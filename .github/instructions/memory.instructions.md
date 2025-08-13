---
applyTo: "**/*"
---

# Memory

This document contains essential, abstract, and up-to-date information about the project, collaboratively maintained by all developers. It is regularly updated to reflect key architectural decisions, subtle implementation details, and recent developments. Outdated content is removed to maintain clarity and relevance.


## Backend & Testing Developments (2025-08)

- **Log API & Cached Token Details:**
    - The backend log API now returns `cached_prompt_tokens` and `cached_completion_tokens` for each log entry. These fields are persisted in the database and surfaced in API responses. The frontend displays these values as tooltips in the Prompt/Completion columns of the logs table for transparency.
    - AutoMigrate ensures new columns are created automatically. No manual DB migration is needed for cached token fields.
    - All log and billing changes are covered by updated unit tests. Any new log field or billing logic must be reflected in both backend API and frontend UI.

- **Test & Race Condition Policy:**
    - All changes must pass `go test -race ./...` before merge. Any test that fails due to argument mismatch, floating-point precision, or race must be fixed immediately.
    - For floating-point comparisons in tests, always use a tolerance (epsilon) instead of strict equality to avoid failures due to precision errors.
    - If a function signature changes (e.g., new arguments to billing functions), update all test calls accordingly. Use zero or default values for new arguments in legacy/compatibility tests.

- **Migration/Edge-Case Test Handling:**
    - Some tests (e.g., migration, timestamp, or error recovery tests) are intentionally designed to fail or log errors to verify error handling. These include tests for invalid JSON, duplicate keys, or constraint violations.
    - These edge-case tests should not be removed, but failures in these tests do not indicate a problem with business logic. Only address these if the test intent changes or if they block CI/CD pipelines.

- **Handover Best Practices:**
    - When handing over, ensure the new assistant is aware of the following:
        - All log/billing API changes must be reflected in both backend and frontend, with tooltips or UI cues for new fields.
        - All tests must pass with `go test -race ./...` except for intentional edge-case/migration tests.
        - If a test fails due to a function signature change, update all test invocations across the codebase.
        - Use tolerance for floating-point assertions in Go tests.
        - Do not remove or silence migration/edge-case tests unless their intent is obsolete.
        - Always keep this file concise, abstract, and up-to-date—remove outdated details as new patterns emerge.



## Frontend API Path Unification & Verification (2025-08)

- **API Path Convention:**
    - All frontend API calls must use explicit, full URLs with the `/api` prefix. The shared Axios client no longer sets a `baseURL`.
    - Every API call (GET, POST, PUT, DELETE) must include `/api/` in the path. This applies to all pages, components, and utility functions.
    - Inline comments (`// Unified API call - complete URL with /api prefix`) are used to clarify this convention for maintainers.

- **Verification & Migration:**
    - A verification script (`grep -r "api\.get|api\.post|api\.put|api\.delete" ... | grep -v "/api/" | grep -v "Unified API call"`) is used to ensure no legacy or missing `/api` prefixes remain.
    - As of August 2025, all API calls in the modern frontend have been verified and fixed to use the `/api` prefix. The migration is complete and consistent.
    - Any future code or third-party integration must follow this invariant. If the backend route structure changes, a full review is required.

- **Subtle Implementation Details & Risks:**
    - Any missed API call without `/api` will fail (404 or unexpected behavior). All new code must be checked for compliance.
    - Components using `fetch()` or other HTTP clients directly must also use the `/api` prefix.
    - Tests, mocks, and documentation must be kept in sync with this convention.
    - If the backend changes the `/api` prefix, both frontend and backend must be updated in lockstep.

- **Handover Guidance:**
    - When handing over, ensure the new assistant is aware of the explicit API path requirement, the verification process, and the need to keep this invariant in all future work. Use the verification script after any major refactor or dependency update.

## Frontend Authentication, Validation, and Testing (2025-08)

- **Login & Registration:**
    - TOTP (Two-Factor Authentication) is strictly validated in the login flow. The UI disables the submit button unless a 6-digit code is entered when required. TOTP state is managed separately from the form state.
    - Success messages (e.g., after registration or password reset) are passed via navigation state and displayed on the login page (not for direct URL access).
    - Email validation in registration is regex-based and enforced before sending verification codes. The "Send Code" button is disabled unless a valid email is entered.
    - Form error handling is decoupled from form context, improving maintainability.

- **Testing Infrastructure:**
    - The modern frontend uses Vitest as the test runner, with `jsdom` as the default environment. All test scripts and TypeScript configs are updated accordingly.
    - All test files must use the correct mocking API for Vitest (e.g., `vi.mock`).
    - The global Vitest setup may affect tests that expect a Node environment or use other runners. All new dependencies must be kept up to date.

- **Subtle Implementation Details & Risks:**
    - TOTP and email validation are stricter; if backend or other clients expect different behavior, login or registration may fail.
    - The use of navigation state for success messages means direct URL access will not show these messages.
    - The refactor of error handling may break custom error handling in other forms if they relied on the previous implementation.
    - The global Vitest setup may affect tests that expect a Node environment or use other runners.

- **Handover Guidance:**
    - When handing over, ensure the new assistant is aware of the stricter validation logic, the decoupled error handling, and the global Vitest test environment. All authentication and registration flows must be tested after changes. Any test runner or environment changes must be validated for compatibility.


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


## Frontend Responsive & Table Architecture (2025-08)

- **Unified Responsive System:**
    - All management tables (Users, Channels, Tokens, Redemptions, Logs) and main pages now use a unified responsive architecture. This includes:
        - `useResponsive` hook for device detection (`isMobile`, `isTablet`, etc.).
        - `ResponsivePageContainer`, `ResponsiveSection`, and `AdaptiveGrid` for consistent, adaptive layouts.
        - Card-based mobile layouts for tables, with label-value pairs and touch-friendly controls.
        - All table action buttons and pagination controls are touch-friendly, visually consistent, and use minimum 44px tap targets.
        - Table columns can be hidden on mobile via `hideColumnsOnMobile` prop; all table cells use `data-label` for accessibility.
        - Pagination is always visible, styled for both desktop and mobile, and never slices data locally for server-side pagination.
        - All table and UI changes must be validated on both desktop and mobile after any refactor.

- **CSS & Tailwind:**
    - `mobile.css` and `tailwind.config.js` are extended for modular, maintainable responsive utilities (custom breakpoints, touch targets, responsive spacing, etc.).
    - Legacy or global CSS overrides that break pagination or table layout have been removed. Always check for such overrides when UI elements are missing.
    - Never use inline style overrides for layout/visibility; always prefer maintainable CSS fixes.
    - Remove outdated or redundant CSS as part of any major UI refactor.

- **Testing & Build:**
    - All bug fixes and features require updated unit tests. No temporary scripts are allowed.
    - Test files must use the correct mocking API for the test runner (e.g., `vi.mock` for Vitest, not `jest.mock`).
    - After major refactors, always validate build and test integrity. Fixes for test runner compatibility (e.g., Vitest vs Jest) should be documented.

- **Subtle Implementation Details:**
    - All model lists for channel/adaptor editing are fetched in real-time from the backend, never from local cache.
    - Table sorting dropdowns and icons are visually unified and always server-side.
    - When adding models to a channel, deduplicate and never overwrite existing selections unless explicitly cleared.
    - Mobile usability and accessibility (touch targets, `data-label`, focus/active states) are first-class concerns.
    - Any new UI/UX or backend logic must be kept in sync with documentation and user-facing messages.

- **Handover Guidance:**
    - When handing over, ensure the new assistant is aware of the unified responsive/table architecture, the importance of data-labels for mobile, the need to keep UI/UX in sync with backend and documentation, and the requirement for proper test/build practices (including test runner compatibility).


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

- **Frontend Authentication & Testing:**
    - Login and registration flows now enforce stricter TOTP and email validation, with improved error and success message handling. The login page now displays navigation-passed success messages and disables submission unless TOTP is valid. Registration email validation is regex-based and enforced before sending codes.
    - The modern frontend has migrated to Vitest with `jsdom` as the default environment. All test scripts, configs, and setup files are updated. All test files must use Vitest APIs.
    - Form error handling is now decoupled from form context, improving maintainability but requiring updates to custom forms.

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
