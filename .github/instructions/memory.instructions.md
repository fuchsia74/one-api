---
applyTo: "**/*"
---

# Memory

This document contains essential, abstract, and up-to-date information about the project, collaboratively maintained by all developers. It is regularly updated to reflect key architectural decisions, subtle implementation details, and recent developments. Outdated content is removed to maintain clarity and relevance.


## Pricing & Billing Architecture (2025-07)

- **Pricing Unit Standardization:** All model pricing, quota, and billing calculations use "per 1M tokens". This is reflected in backend, UI, and documentation. Always keep user-facing messages and docs in sync with backend logic.
- **Centralized Model Pricing:** All channels/adaptors use a shared `ModelRatios` constant. Local pricing maps are deprecated. Supported model lists are derived from the shared pricing map keys.
- **Default/Fallback Pricing:** Unified fallback logic for unknown models. No local overrides.
- **VertexAI Aggregation:** VertexAI pricing aggregates all subadaptors. Any omission propagates to VertexAI.

## General Project Practices

- **Error Handling:** Always use `github.com/Laisky/errors/v2` for error wrapping. Never return bare errors.
- **Context Keys:** All context keys are pre-defined in `common/ctxkey/key.go`.
- **Package Management:** Use package managers only. Never edit package files by hand.
- **Testing:** All bug fixes/features require updated unit tests. No temporary scripts.
- **Time Handling:** Always use UTC for server, DB, and API time.
- **Golang ORM:** Use `gorm.io/gorm` for writes; prefer SQL for reads to minimize DB load.

## Handover Guidance

- **Claude Messages API:** Universal conversion and billing parity. See `docs/arch/api_billing.md` and `docs/arch/api_convert.md` for details.
- **Billing Architecture:** Four-layer pricing (channel overrides > adapter defaults > global > fallback).
- **Adaptor Pattern:** All new API formats follow the Claude Messages pattern: interface method + universal conversion + context marking.

---
**Recent Developments (2025-08):**

- **Frontend Pagination Bug (2025-08):**
    - **Root Cause:** Pagination controls in all management tables (Tokens, Channels, Users, Logs, Redemptions) were hidden due to a legacy CSS rule: `.ui.table tfoot .ui.pagination { display: none !important; }`.
    - **Symptoms:** Pagination logic and API were correct, but the UI did not show pagination controls, leading to confusion and apparent missing functionality.
    - **Resolution:** The problematic CSS rule was removed from `web/default/src/index.css`. All table components now use server-side pagination, and the Semantic UI Pagination component is visible and functional in the table footer.
    - **Subtlety:** Always check for global or legacy CSS overrides when UI elements are unexpectedly missing, especially with third-party UI libraries. Avoid inline style overrides; prefer maintainable CSS fixes.
    - **Best Practice:** Pagination logic should match the data loading strategy (server-side vs. client-side). For server-side pagination, do not slice data in the component; render the API result directly.

- Major refactor to unify and clarify model pricing logic, reduce duplication, and standardize on "per 1M tokens" as the pricing unit. All adaptors now use shared pricing maps and fallback logic. This change is critical for maintainability and billing accuracy.
- When handing over, ensure the new assistant is aware of the pricing unit change, the centralized pricing logic, and the importance of keeping documentation and UI in sync with backend logic.

## Claude Messages API: Universal Conversion

- All Claude Messages API requests (`/v1/messages`) are routed to adapters implementing `ConvertClaudeRequest(c, request)` and `DoResponse(c, resp, meta)`. Conversion state is tracked via context keys. Data mapping is bi-directional (Claude ↔ OpenAI), with full support for function calling, streaming, and tool use. Billing, quota, and token counting are handled identically to ChatCompletion. See `relay/controller/claude_messages.go` and related files for reference.
- New adapters follow the Claude Messages pattern: interface method + universal conversion + context marking. Specialized adapters (e.g., DeepL, Palm, Ollama) are excluded from Claude Messages support.

## Gemini Adapter: Function Schema Cleaning

- Gemini API rejects OpenAI-style function schemas with unsupported fields (`additionalProperties`, `description`, `strict`). Recursive cleaning removes `additionalProperties` everywhere, and `description`/`strict` only at the top level.

## Handover Guidance

- **Critical Files:**
  - `relay/controller/claude_messages.go`, `relay/adaptor/interface.go`, `common/ctxkey/key.go`, `docs/arch/api_billing.md`, `docs/arch/api_convert.md`
- **Subtle Details:**
  - All pricing, quota, and billing logic must be kept in sync with documentation and UI. Any backend change must be reflected in user-facing docs and messages.
  - The four-layer pricing and fallback logic is critical for maintainability and billing accuracy. Never bypass it.
  - DB pool settings and billing timeout are tuned for high concurrency; operators must monitor and adjust for their environment.
  - All adapters use the shared pricing map and fallback logic—no local overrides.
  - For any new adapter or API, follow the Claude Messages and pricing patterns strictly.

---

**Recent Developments (2025-08):**

- **Frontend Pagination Bug:**
    - Pagination controls in management tables were hidden due to a legacy CSS rule. The rule was removed; all tables now use server-side pagination and visible controls.
    - Always check for global/legacy CSS overrides when UI elements are missing. Prefer maintainable CSS fixes over inline overrides.
    - Pagination logic must match data loading strategy. For server-side pagination, render API results directly.

- **Model Pricing Refactor:**
    - All pricing, quota, and billing logic standardized to "per 1M tokens". Shared pricing maps and fallback logic are used everywhere. Documentation and UI must always match backend logic.

- **Table Sorting & Data Consistency:**
    - All management tables (Users, Channels, Tokens, Redemptions, Logs) now use server-side sorting and pagination. Sorting is unified via dropdown and clickable column headers with icons. All tables fetch data in real-time from the server—no local cache is used for editing or display.
    - "Fill Related Models" and "Fill All Models" buttons in channel editing serve distinct purposes: Related adds only models supported by the current channel/adaptor; All adds all available models. Both deduplicate automatically. This logic is implemented in the frontend and must be maintained.
    - When adding models to a channel, always deduplicate and never overwrite existing selections unless explicitly cleared.

- **Subtle Implementation Details:**
    - All model lists for channel/adaptor editing are fetched in real-time from the backend, never from local cache. Channel-specific models are fetched via `/api/models` and mapped by channel type.
    - Table sorting dropdowns and icons are visually unified across all management tables. Sorting is always server-side and applies to all data, not just the current page.
    - All bug fixes and features require updated unit tests. No temporary scripts are allowed.

- **Build Issue Resolution:**
    - Corrupted `TokensTableCompact.js` was rebuilt with proper imports, JSX, and sorting logic. Always validate file integrity after major refactors.

- **Handover Best Practices:**
    - When handing over, ensure the new assistant is aware of the pricing unit change, centralized pricing logic, table sorting/pagination patterns, and the importance of keeping documentation and UI in sync with backend logic. Remove outdated content and keep this file concise and abstract.

## Claude Messages API: Universal Conversion

- All Claude Messages API requests (`/v1/messages`) are routed to adapters that implement `ConvertClaudeRequest(c, request)` and `DoResponse(c, resp, meta)`. Anthropic uses native passthrough; most others use OpenAI-compatible or custom conversion.
- Conversion state is tracked via context keys in `common/ctxkey/key.go` (e.g., `ClaudeMessagesConversion`, `ConvertedResponse`, `OriginalClaudeRequest`).
- Data mapping is bi-directional: Claude → OpenAI (system, messages, tools, tool_choice, etc.) and OpenAI → Claude (choices, tool_calls, finish_reason, usage, etc.). Gemini and some others use custom mapping.
- Full support for function calling, streaming, structured/multimodal content, and tool use. Billing, quota, and token counting are handled identically to ChatCompletion, including image token calculation and fallback strategies.
- All errors are wrapped with `github.com/Laisky/errors/v2` and surfaced with context. Malformed content is handled gracefully with fallbacks.
- New adapters should follow the Claude Messages pattern: interface method + universal conversion + context marking. Specialized adapters (e.g., DeepL, Palm, Ollama) are excluded from Claude Messages support.
- See `relay/controller/claude_messages.go`, `relay/adaptor/interface.go`, `relay/adaptor/openai_compatible/claude_messages.go`, `relay/adaptor/gemini/adaptor.go`, `common/ctxkey/key.go`, and `docs/arch/api_convert.md` for reference.

## Pricing & Billing Architecture (2025-07)

- **Pricing Unit Standardization:** All model pricing, quota, and billing calculations are now standardized to use "per 1M tokens" (1 million tokens) instead of "per 1K tokens". This is reflected in all code, comments, and documentation. Double-check all user-facing messages and documentation for consistency.
- **Centralized Model Pricing:** Each channel/adaptor now imports and uses a shared `ModelRatios` constant from its respective `constants.go` or subadaptor. Local, hardcoded pricing maps have been removed to avoid duplication and drift.
- **Model List Generation:** Supported model lists are always derived from the keys of the shared pricing maps, ensuring pricing and support are always in sync.
- **Default/Fallback Pricing:** All adaptors use a unified fallback (e.g., `5 * ratio.MilliTokensUsd`) for unknown models. If a model is missing from the shared map, it will use this fallback.
- **VertexAI Aggregation:** VertexAI pricing is now aggregated from all subadaptors (Claude, Imagen, Gemini, Veo) and includes VertexAI-specific models. Any omission in a subadaptor will propagate to VertexAI.
- **Critical Subtleties:**
  - If any model is missing from the shared pricing map, it may become unsupported or use fallback pricing.
  - Models with non-token-based pricing (e.g., per image/video) require special handling and may not fit the token-based pattern.
  - All documentation and UI must be kept in sync with the new pricing unit to avoid confusion.

## Gemini Adapter: Function Schema Cleaning

- Gemini API rejects OpenAI-style function schemas with unsupported fields (`additionalProperties`, `description`, `strict`).
- Recursive cleaning removes `additionalProperties` everywhere, and `description`/`strict` only at the top level. Cleaned parameters are type-asserted before assignment.
- Only remove `description`/`strict` at the top; nested objects may require them.

## General Project Practices

- **Error Handling:** Always use `github.com/Laisky/errors/v2` for error wrapping; never return bare errors.
- **Context Keys:** All context keys must be pre-defined in `common/ctxkey/key.go`.
- **Package Management:** Use package managers (npm, pip, etc.), never edit package files by hand.
- **Testing:** All bug fixes/features must be covered by unit tests. No temporary scripts. Unit tests must be updated to cover new issues and features.
- **Time Handling:** Always use UTC for server, DB, and API time.
- **Golang ORM:** Use `gorm.io/gorm` for writes; prefer SQL for reads to minimize DB load.

## Handover Guidance

- **Claude Messages API:** Fully production-ready, with universal conversion and billing parity. See `docs/arch/api_billing.md` and `docs/arch/api_convert.md` for details.
- **Billing Architecture:** Four-layer pricing (channel overrides > adapter defaults > global > fallback).
- **Adaptor Pattern:** All new API formats should follow the Claude Messages pattern: interface method + universal conversion + context marking.
- **Critical Files:**
  - `relay/controller/claude_messages.go`
  - `relay/adaptor/interface.go`
  - `common/ctxkey/key.go`
  - `docs/arch/api_billing.md`
  - `docs/arch/api_convert.md`

---
**Recent Developments (2025-07):**

- Major refactor to unify and clarify model pricing logic, reduce duplication, and standardize on "per 1M tokens" as the pricing unit. All adaptors now use shared pricing maps and fallback logic. This change is critical for maintainability and billing accuracy.
- When handing over, ensure the new assistant is aware of the pricing unit change, the centralized pricing logic, and the importance of keeping documentation and UI in sync with backend logic.
