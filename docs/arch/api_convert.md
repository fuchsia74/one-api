# API Format Conversion Guide

one-api provides transparent, bidirectional conversion across the three chat-style APIs it exposes:

- **OpenAI Chat Completions** (`/v1/chat/completions`)
- **OpenAI Responses** (`/v1/responses`)
- **Claude Messages** (`/v1/messages`)

Regardless of the entrypoint a caller chooses, the platform delivers the reply using the same format the caller used while freely routing the upstream call to whatever protocol the target model or channel actually understands.

---

## 1. Supported Conversions at a Glance

| User Request Format | Possible Upstream Formats                                 | Response Back to User   |
| ------------------- | --------------------------------------------------------- | ----------------------- |
| Chat Completions    | Chat Completions • Responses • Claude Messages            | Always Chat Completions |
| Responses           | Responses • Chat Completions (fallback) • Claude Messages | Always Responses        |
| Claude Messages     | Claude Messages • Chat Completions • Responses            | Always Claude Messages  |

Key points:

- Native-capable channels are contacted using their preferred protocol (for example, OpenAI GPT-4o via Responses, Anthropic Claude via Claude Messages).
- Channels lacking native Responses support automatically fall back to Chat Completions while the controller rebuilds a Responses payload for the caller.
- Cross-family access (Claude client → OpenAI model, OpenAI client → Claude model, etc.) works without user code changes.

---

## 2. Request Routing Overview

```text
Incoming request --> Identify controller (Chat / Response / Claude)
                  --> Parse model + channel metadata
                  --> Apply capability gates + sanitizers
                  --> Convert request if upstream protocol differs
                  --> Call adaptor
                  --> Convert response/stream back to caller format
                  --> Reconcile usage + quota
```

Important building blocks:

- `meta.Meta` stores routing facts (channel, model mapping, URL path, fallback flags).
- Conversion utilities live primarily in `relay/adaptor/openai` and `relay/adaptor/openai_compatible`.
- Controllers (`relay/controller/*.go`) coordinate conversion, quota, and response rewriting.

---

## 3. Entry Point Details

### 3.1 Chat Completions (`/v1/chat/completions`)

1. Controller parses the payload into `relay/model.GeneralOpenAIRequest`.
2. If the downstream channel is OpenAI **and** `IsModelsOnlySupportedByChatCompletionAPI` returns `false` **and** the original URL was `/v1/chat/completions`, the adaptor upgrades the request to a Responses payload via `ConvertChatCompletionToResponseAPI`.
3. The converted request is cached under `ctxkey.ConvertedRequest` so the response path knows it must translate the upstream payload back into Chat Completion format.
4. When the adaptor detects that the channel only offers Chat Completions (search models or third-party compatibles), it forwards the original request unchanged.
5. Response handling mirrors the request decision: Responses bodies are converted with `ResponseAPIHandler`, while vanilla Chat Completion bodies use the standard handler.

### 3.2 Responses (`/v1/responses`)

1. The controller parses the JSON into `openai.ResponseAPIRequest`, then runs `sanitizeResponseAPIRequest` to clear unsupported parameters (for example, reasoning models drop `temperature`/`top_p`).
2. `normalizeResponseAPIRawBody` rewrites the raw JSON in-place so that forbidden fields are removed before the request ever leaves the proxy. This keeps upstream validation errors from leaking to callers.
3. `supportsNativeResponseAPI(meta)` decides whether the channel can speak Responses directly. It currently returns `true` only for official OpenAI endpoints (`api.openai.com`). All other channels, including GPT-OSS deployments and OpenAI-compatible vendors, opt into the Chat Completion fallback.
4. When falling back, the controller converts the request with `ConvertResponseAPIToChatCompletionRequest`, updates `meta.RequestURLPath` to `/v1/chat/completions`, and sets `meta.ResponseAPIFallback = true` to avoid recursive reconversion later in the pipeline.
5. The adaptor call proceeds. If a fallback was used, the upstream Chat Completion response is transformed back into a Responses envelope via `ResponseAPIHandler` (non-streaming) or `ResponseAPIStreamHandler` (streaming). The helper registered under `ctxkey.ResponseRewriteHandler` performs the final rewrite before bytes are flushed to the client.
6. `normalizeResponseAPIRawBody` also deletes `temperature`/`top_p` keys from the raw payload when the sanitized struct dropped them, ensuring double coverage for channels that reject those parameters outright.

### 3.3 Claude Messages (`/v1/messages`)

1. `RelayClaudeMessagesHelper` inspects the requested model to determine the target channel.
2. Anthropic-native channels set `ctxkey.ClaudeMessagesNative` and forward the request untouched.
3. OpenAI-compatible, Gemini, and other providers convert the Claude payload into their preferred format via adaptor-specific `ConvertClaudeRequest` implementations. Most OpenAI compatibles share `openai_compatible.ConvertClaudeRequest`.
4. During response handling, adaptors check `ctxkey.ClaudeMessagesConversion`. When present, they transform the upstream response (or SSE stream) back into Claude Messages events using utilities such as `openai_compatible.HandleClaudeMessagesResponse` and `ConvertOpenAIStreamToClaudeSSE`.
5. The controller always returns Claude-flavoured JSON/SSE to the caller, regardless of the intermediate protocols.

---

## 4. Capability Detection & Sanitisation

| Concern                                 | Implementation                                                                        | Notes                                                                            |
| --------------------------------------- | ------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| Model supports only Chat Completions    | `IsModelsOnlySupportedByChatCompletionAPI` (`relay/adaptor/openai/response_model.go`) | Matches search/audio GPT variants; skips Responses conversion.                   |
| Channel supports native Responses       | `supportsNativeResponseAPI` (`relay/controller/response.go`)                          | Accepts only OpenAI first-party endpoints. Others use Chat fallback.             |
| Reasoning models reject sampling params | `sanitizeResponseAPIRequest` + `sanitizeChatCompletionRequest`                        | Clears `temperature` and `top_p` for GPT-5 / o-series models.                    |
| Raw payload must match sanitised struct | `normalizeResponseAPIRawBody`                                                         | Removes stripped keys from the outbound JSON to avoid upstream 400s.             |
| Fallback recursion protection           | `meta.ResponseAPIFallback`                                                            | Prevents a Chat fallback request from being re-upgraded to Responses downstream. |

These safeguards execute before every upstream call, so the same rules apply to retries and multi-channel failovers.

---

## 5. Streaming Behaviour

- **Responses → Chat fallback:** `ResponseAPIStreamHandler` rebuilds SSE sequences (`response.created`, `response.output_text.delta`, etc.) from Chat Completion chunks and emits the `response.completed` summary once usage is known. The helper also ensures a terminating `data: [DONE]` envelope for clients.
- **Chat → Responses upgrade:** When `/v1/chat/completions` requests are upgraded to Responses, `ResponseAPIDirectStreamHandler` passes through upstream Responses SSE untouched.
- **Claude Messages:** `ConvertOpenAIStreamToClaudeSSE` produces Claude-native event types (`message_start`, `content_block_delta`, …) while accumulating text and tool call arguments for billing.

---

## 6. Error Handling & Billing

1. Controllers pre-consume quota using the same logic regardless of protocol. Response fallback calls use the Chat Completion quota helpers but reconcile against the final Responses usage once conversion completes.
2. All adaptor errors are wrapped with `openai.ErrorWrapper` (or the channel equivalent) so HTTP status codes and machine-readable error bodies survive conversions.
3. Token accounting prioritises upstream usage. When upstream omits it, the system estimates totals from streamed text, tool call arguments, and prompt size.
4. Billing post-processing funnels through `billing.PostConsumeQuotaDetailed`, which now receives the original Responses model name even after a fallback path.

---

## 7. Context Keys & Runtime Flags

| Key                                                               | Purpose                                                                                                 |
| ----------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `ctxkey.ConvertedRequest`                                         | Stores the converted request (Responses or Chat) so the response side can pick the proper formatter.    |
| `ctxkey.ConvertedResponse`                                        | Holds a fully converted response object awaiting controller flush (used heavily by Claude conversions). |
| `ctxkey.ResponseRewriteHandler`                                   | Callback used in Responses fallback to stream Chat output back as Responses SSE.                        |
| `ctxkey.ResponseAPIRequestOriginal`                               | Snapshot of the user’s original Responses payload for metadata echoes.                                  |
| `ctxkey.ClaudeMessagesConversion` / `ctxkey.ClaudeMessagesNative` | Flags describing the Claude pipeline path.                                                              |
| `meta.ResponseAPIFallback`                                        | Marks that the active request already fell back to Chat Completions.                                    |

Refer to `common/ctxkey/key.go` and `relay/meta/relay_meta.go` for the authoritative list.

---

## 8. Testing Coverage

Relevant test suites validating the behaviour above:

- `relay/adaptor/openai/channel_conversion_test.go` – ensures Chat ↔ Responses conversion toggles correctly for OpenAI, Azure, GPT-OSS, etc.
- `relay/controller/response_fallback_test.go`
  - `TestRelayResponseAPIHelper_FallbackStreaming` exercises SSE rewriting.
  - `TestNormalizeResponseAPIRawBody_RemovesUnsupportedParams` guards sanitisation.
- `relay/adaptor/openai/response_model_test.go` – bidirectional conversion tests for Chat ↔ Responses, including function calling and streaming.
- `relay/controller/claude_messages_test.go` plus adaptor-specific suites – validate Claude ↔ OpenAI/Gemini conversions.
- `cmd/test` regression sweep – cross-api smoke tests that hit every public adaptor.

Run everything with:

```bash
GOFLAGS=-race go test ./...
```

(Controllers and adaptors can also be targeted individually for faster iteration.)

---

## 9. Summary & Further Reading

- All three chat-style APIs can be used interchangeably from the client’s perspective; one-api will translate on the fly.
- Capability detection ensures each upstream sees only the fields it supports, falling back to Chat Completions when Responses is unavailable.
- Streaming and billing remain accurate across conversions thanks to shared handlers and usage reconciliation.
- Internal flags (`ConvertedRequest`, `ResponseAPIFallback`, `ClaudeMessagesConversion`, etc.) tie the request/response lifecycles together.

For deeper implementation insight, explore:

- `relay/controller/response.go`
- `relay/adaptor/openai/adaptor.go`
- `relay/controller/claude_messages.go`
- `relay/adaptor/openai_compatible/claude_messages.go`

These modules contain the authoritative logic that keeps the three API formats in sync.
