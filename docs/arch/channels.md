# Channels: Technical Implementation Guide

- [Channels: Technical Implementation Guide](#channels-technical-implementation-guide)
  - [Overview](#overview)
  - [Data model and cache](#data-model-and-cache)
  - [Channel selection algorithm](#channel-selection-algorithm)
  - [Retry](#retry)
  - [Failure handling](#failure-handling)
  - [Temporary suspension vs auto-disable policy](#temporary-suspension-vs-auto-disable-policy)
  - [Metrics integration](#metrics-integration)
  - [Configuration knobs](#configuration-knobs)
  - [Operational guidance](#operational-guidance)
  - [Known limitations](#known-limitations)

## Overview

This document describes how One API routes requests across channels and how failures trigger retries and channel state changes. It maps directly to the implementation in:

- `controller/relay.go` (request flow, retry driver)
- `model/cache.go` (in‑memory channel selection)
- `model/ability.go` (per group/model/channel capability and suspension)
- `monitor/*` (auto disable decisions and notifications)

Terminology:

- Channel: a provider connection (row in `channels` table) that supports one or more models for one or more groups.
- Ability: tuple (group, model, channel_id) that is enabled/disabled and can be temporarily suspended.

## Data model and cache

- Channels are stored in `channels` with fields including `status`, `group`, `models` (CSV), and `priority`.
- Abilities are stored in `abilities` with fields: `group`, `model`, `channel_id`, `enabled`, `priority`, `suspend_until`.
- The in‑memory cache (`model.InitChannelCache`) builds `group2model2channels`:
  - Only channels with `status = enabled` are considered.
  - Only abilities where `enabled = true` and `suspend_until` is nil or in the past are included.
  - For each (group, model), channels are sorted by priority descending (higher number = higher priority).
- Cache refresh runs every `SYNC_FREQUENCY` seconds; see [Configuration knobs](#configuration-knobs).

## Channel selection algorithm

Selection uses the in‑memory cache when `MEMORY_CACHE_ENABLED` is true; otherwise DB queries are used by the non‑cache variants.

- Default selection (outside of the retry handler) prefers highest priority:
  - Among available candidates for (group, model), pick highest priority (max value); if multiple at that priority, choose randomly.
  - If `DEFAULT_USE_MIN_MAX_TOKENS_MODEL` is true, within the highest priority tier, prefer the smallest `max_tokens` (from `Model Configs`) and randomize among ties. This optimizes for lower capacity usage by default.
- Retry selection (see next section) alters the target tier and/or filters by `max_tokens` depending on error class.

Priority semantics (as implemented): higher integer value = higher priority. “Ignore first priority” in code means “skip the current highest priority tier and try lower tiers.”

## Retry

The retry driver lives in `controller/relay.go::Relay` and executes after an initial attempt fails.

Entry and gating

- Initial attempt: `relayHelper` performs the request against the selected channel.
- If it fails with `bizErr != nil`, we may retry unless a specific channel was forced:
  - If `SpecificChannelId` is present in the request context, retries are disabled and the error is returned immediately.

Retry budget

- Base attempts derive from `RETRY_TIMES` (default 0 = no generic retries).
- Special handling by error class:
  - 429 Too Many Requests (rate limit): if `RETRY_TIMES > 0`, the attempt budget is doubled to probe more alternatives.
  - 413 Request Entity Too Large (capacity): attempt budget becomes “all other channels” for the same (group, model), i.e., `len(channels) - 1`. If the cache lookup fails, fall back to 1 retry. This path activates regardless of `RETRY_TIMES`.
  - 5xx or network transport errors (server/transport transient): keep base budget; avoid reusing the same ability first.
  - Client request errors (4xx due to user input, e.g., schema/validation): no generic retries.

Per‑attempt selection strategy

- A failed channel is added to an in‑memory exclusion set for the duration of this request. Subsequent selections avoid these channels.
- Strategy depends on the classified error of the most recent attempt:
  - Rate‑limit (429):
    - Prefer lower priority tiers first to escape localized throttling. If no lower tier available, fall back to the highest priority tier among remaining candidates.
  - Capacity (413):
    - Prefer channels whose `max_tokens` for the requested model differs from the failed ones (channels with no `max_tokens` limit are also eligible). This avoids immediately retrying channels with the same capacity constraint that just failed.
  - Server/transport transient (5xx/network):
    - Avoid the exact (channel, model) ability first; probe other abilities in the same tier to maintain performance, then drop to lower tiers if needed.
  - Client request errors (4xx due to user input):
    - Do not retry; surface the error.

Request replay

- The original request body is cached (`common.GetRequestBody`) and the HTTP request body is reset for each retry.

Attempt accounting and exit

- Each attempt (initial and retries) records Prometheus metrics; on success the handler returns immediately.
- If selection fails (no candidates) or we exhaust the budget, the last error is returned. For 429 after exhausting retries, the message is rewritten to be more actionable:
  - If multiple channels were tried: “All available channels (N) for this model are currently rate limited, please try again later”.
  - Otherwise: “The current group load is saturated, please try again later”.
- The response always appends the `request_id` for traceability.

## Failure handling

Error classification and side‑effects (per attempt)

- After each failure, `processChannelRelayError` runs asynchronously and classifies the error origin. Actions are scoped to the specific (group, model, channel_id) “ability” unless the issue is proven to be channel‑wide/fatal.

Classification (examples):

- Client request error:
  - 400 and similar schema/validation errors; vendor type `invalid_request_error`.
  - Action: no suspension, no auto‑disable. Emit failure metric only.
- Rate limit (channel‑origin transient):
  - 429, or vendor‑specific rate‑limit types.
  - Action: suspend the ability for `CHANNEL_SUSPEND_SECONDS_FOR_429`.
- Capacity (model/token/context window limits):
  - 413, or explicit vendor messages for token/context overflow.
  - Action: no suspension by default; rely on retry strategy to pick channels with larger `max_tokens`.
- Server/transport transient (channel‑origin):
  - 5xx responses; network timeouts/EOF/connection reset; upstream gateway failures.
  - Action: suspend the ability for a short window `CHANNEL_SUSPEND_SECONDS_FOR_5XX`.
- Auth/permission/quota (potentially channel‑wide):
  - 401/403; error types `authentication_error`, `permission_error`, `insufficient_quota`; known vendor strings like “API key not valid/expired”, “organization restricted”, “已欠费/余额不足”.
  - Action: suspend the ability for `CHANNEL_SUSPEND_SECONDS_FOR_AUTH`. If `monitor.ShouldDisableChannel` deems the condition fatal (e.g., invalid API key or deactivated account), auto‑disable the entire channel.

Debugging aids

- With `DEBUG=true`, the retry loop logs the exclusion set and attempt ordering. If selection fails, it queries the DB for the excluded channels’ suspension status to help diagnose cache vs. DB discrepancies.

Concurrency note

- There is a known race on `bizErr` mutation before response serialization; this is safe in practice because the mutation occurs on the last reference prior to write, but it is noted in the code.

## Temporary suspension vs auto-disable policy

Suspension (temporary, per ability)

- Scope: always the ability (group, model, channel_id), not the entire channel.
- Triggers: channel‑origin/transient classes (rate‑limit 429, server/transport 5xx/network), and optionally auth/quota/permission depending on severity.
- Actions:
  - 429: set `abilities.suspend_until = now + CHANNEL_SUSPEND_SECONDS_FOR_429`.
  - 5xx/network: set `abilities.suspend_until = now + CHANNEL_SUSPEND_SECONDS_FOR_5XX`.
  - Auth/quota/permission: set `abilities.suspend_until = now + CHANNEL_SUSPEND_SECONDS_FOR_AUTH`, unless immediately escalated to auto‑disable by policy.
- Effect: the ability is excluded from selection until suspension expires and the cache refreshes.

Auto‑disable (persistent, per channel)

- Gate: `AUTOMATIC_DISABLE_CHANNEL_ENABLED` must be true.
- Reserved for fatal channel‑wide conditions verified by `monitor.ShouldDisableChannel`:
  - Invalid API key, account deactivated, hard permission denials, permanent organization restrictions, clear vendor policy violations.
- Action: `monitor.DisableChannel` sets `channels.status = auto_disabled` and sends a notification (email or message pusher).

## Metrics integration

- For each request:
  - `RecordChannelRequest` increments in‑flight counters by channel/type and decrements later.
  - `RecordRelayRequest` records success/failure, usage, user quota, and per‑model latency when successful.
  - Post‑refactor: failures should be tagged with error class (client/rate‑limit/capacity/server/auth) to aid dashboards and alerting.

## Configuration knobs

Environment variables (see `common/config/config.go`):

- `CHANNEL_SUSPEND_SECONDS_FOR_429` (int seconds, default 60): ability suspension window after 429.
- `CHANNEL_SUSPEND_SECONDS_FOR_5XX` (int seconds, default 30): ability suspension window after transient 5xx/network errors.
- `CHANNEL_SUSPEND_SECONDS_FOR_AUTH` (int seconds, default 300): ability suspension window after auth/quota/permission errors, unless escalated to channel‑wide auto‑disable.
- `MEMORY_CACHE_ENABLED` (bool): enable in‑memory channel cache. Auto‑enabled when Redis is enabled.
- `SYNC_FREQUENCY` (int seconds, default 600): cache refresh interval for channels and abilities.
- `DEBUG` (bool): verbose retry diagnostics and DB suspension dumps.
- `ENABLE_PROMETHEUS_METRICS` (bool, default true): enable Prometheus metrics.
- `AUTOMATIC_DISABLE_CHANNEL_ENABLED` (bool, default false): allow auto‑disabling channels on fatal errors.
- `DEFAULT_USE_MIN_MAX_TOKENS_MODEL` (bool, default false): default selection prefers smaller `max_tokens` within top priority tier.

## Operational guidance

- Priority assignment: higher integer = higher priority. Place primary channels at higher numbers; backups use lower numbers. The retry engine will intentionally drop to lower tiers first on 429 to escape local rate limits.
- Max tokens configuration: populate `Model Configs` with realistic `max_tokens` per model. This enables the 413 path to move to channels with larger capacity.
- Pinning to a specific channel disables retries: if a request carries a specific channel id (populated into `SpecificChannelId`), the system will not try alternatives.
- Cache consistency: suspensions take effect for new requests after the next cache refresh; the current request already excludes the failed channel via its local exclusion set.
- Prefer ability‑level suspension for transient issues; reserve channel‑wide disable for fatal vendor/account problems.

## Known limitations

- With `RETRY_TIMES = 0`, only 413 errors will trigger multi‑channel retries (due to explicit override). Other errors will not retry unless you set a positive budget.
- The in‑memory cache is eventually consistent (refresh interval). A freshly suspended ability may remain in the cache until the next sync; selection still avoids it within the same request.
- Error string heuristics for classification/auto‑disable are provider‑dependent and may need updates as provider messages evolve.
