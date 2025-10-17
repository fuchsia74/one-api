# One API

## Synopsis

One‑API is a **single‑endpoint gateway** that lets you manage and call dozens of AI SaaS models without the headache of custom adapters. 🌐 Simply change the `model_name` and you can reach OpenAI, Anthropic, Gemini, Groq, DeepSeek, and many others—all through the same request format.

![](https://s3.laisky.com/uploads/2025/07/oneapi.drawio.png)

```plain
=== One-API Regression Matrix ===

Variant                        gpt-4o-mini  gpt-5-mini   claude-3.5-haiku  gemini-2.5-flash  openai/gpt-oss-20b  deepseek-chat
Chat (stream=false)            PASS 1.51s   PASS 10.63s  PASS 1.89s        PASS 0.94s        PASS 0.54s          PASS 1.91s
Chat (stream=true)             PASS 1.60s   PASS 9.82s   PASS 0.99s        PASS 3.42s        PASS 0.57s          PASS 2.39s
Chat Tools (stream=false)      PASS 4.00s   PASS 6.26s   PASS 1.75s        PASS 2.03s        PASS 0.62s          PASS 4.56s
Chat Tools (stream=true)       PASS 2.67s   PASS 5.42s   PASS 1.70s        PASS 1.24s        PASS 0.64s          PASS 2.60s
Response (stream=false)        PASS 2.02s   PASS 9.31s   PASS 4.73s        PASS 3.53s        PASS 1.28s          PASS 4.84s
Response (stream=true)         PASS 1.85s   PASS 9.97s   PASS 3.35s        PASS 1.96s        PASS 1.14s          PASS 3.47s
Response Tools (stream=false)  PASS 3.17s   PASS 4.86s   PASS 1.71s        PASS 1.66s        PASS 1.01s          PASS 2.48s
Response Tools (stream=true)   PASS 1.57s   PASS 3.51s   PASS 1.64s        PASS 2.13s        PASS 1.10s          PASS 3.52s
Claude (stream=false)          PASS 1.40s   PASS 8.66s   PASS 2.42s        PASS 1.31s        PASS 0.81s          PASS 2.40s
Claude (stream=true)           PASS 1.43s   PASS 5.32s   PASS 1.32s        PASS 1.22s        PASS 0.89s          PASS 3.81s
Claude Tools (stream=false)    PASS 1.30s   PASS 7.90s   PASS 2.31s        PASS 3.68s        PASS 0.56s          PASS 3.89s
Claude Tools (stream=true)     PASS 2.83s   PASS 8.73s   PASS 1.51s        PASS 3.77s        PASS 0.47s          PASS 2.66s

Totals  | Requests: 72 | Passed: 72 | Failed: 0 | Skipped: 0

2025-10-15T23:34:30Z    INFO    oneapi-test     test/main.go:31 all tests passed

```

### Why this fork exists

The original author stopped maintaining the project, leaving critical PRs and new features unaddressed. As a long‑time contributor, I’ve forked the repository and rebuilt the core to keep the ecosystem alive and evolving.

### What’s new

- **🔧 Complete billing overhaul** – per‑channel pricing for the same model, discounted rates for cached inputs, and transparent usage reports.
- **🖥️ Refreshed UI/UX** – a fully rewritten front‑end that makes tenant, quota, and cost management a breeze.
- **🔀 Transparent API conversion** – send a request in **ChatCompletion**, **Response**, or **Claude Messages** format and One‑API will automatically translate it to the target provider’s native schema.
- **🔄 Drop‑in database compatibility** – the original One‑API schema is fully supported, so you can migrate without data loss or schema changes.
- **🐳 Multi‑architecture support** – runs on Linux x86_64, ARM64, and Windows out of the box.

Docker images available on Docker Hub:

- `ppcelery/one-api:latest`
- `ppcelery/one-api:arm64-latest`

One‑API empowers developers to:

- Centralize **tenant management** and **quota control**.
- Route AI calls to the right model with a single endpoint.
- Track usage and costs in real time.

Also welcome to register and use my deployed one-api gateway, which supports various mainstream models. For usage instructions, please refer to <https://wiki.laisky.com/projects/gpt/pay/>.

- [One API](#one-api)
  - [Synopsis](#synopsis)
    - [Why this fork exists](#why-this-fork-exists)
    - [What’s new](#whats-new)
  - [Tutorial](#tutorial)
    - [Docker Compose Deployment](#docker-compose-deployment)
    - [Kubernetes Deployment](#kubernetes-deployment)
  - [Contributors](#contributors)
  - [New Features](#new-features)
    - [Universal Features](#universal-features)
      - [Support update user's remained quota](#support-update-users-remained-quota)
      - [Get request's cost](#get-requests-cost)
      - [Support Tracing info in logs](#support-tracing-info-in-logs)
      - [Support Cached Input](#support-cached-input)
        - [Support Anthropic Prompt caching](#support-anthropic-prompt-caching)
      - [Automatically Enable Thinking and Customize Reasoning Format via URL Parameters](#automatically-enable-thinking-and-customize-reasoning-format-via-url-parameters)
        - [Reasoning Format - reasoning-content](#reasoning-format---reasoning-content)
        - [Reasoning Format - reasoning](#reasoning-format---reasoning)
        - [Reasoning Format - thinking](#reasoning-format---thinking)
    - [OpenAI Features](#openai-features)
      - [(Merged) Support gpt-vision](#merged-support-gpt-vision)
      - [Support openai images edits](#support-openai-images-edits)
      - [Support OpenAI o1/o1-mini/o1-preview](#support-openai-o1o1-minio1-preview)
      - [Support gpt-4o-audio](#support-gpt-4o-audio)
      - [Support OpenAI web search models](#support-openai-web-search-models)
      - [Support gpt-image-1's image generation \& edits](#support-gpt-image-1s-image-generation--edits)
      - [Support o3-mini \& o3 \& o4-mini \& gpt-4.1 \& o3-pro \& reasoning content](#support-o3-mini--o3--o4-mini--gpt-41--o3-pro--reasoning-content)
      - [Support OpenAI Response API](#support-openai-response-api)
      - [Support gpt-5 family](#support-gpt-5-family)
      - [Support o3-deep-research \& o4-mini-deep-research](#support-o3-deep-research--o4-mini-deep-research)
      - [Support Codex Cli](#support-codex-cli)
    - [Anthropic (Claude) Features](#anthropic-claude-features)
      - [(Merged) Support aws claude](#merged-support-aws-claude)
      - [Support claude-3-7-sonnet \& thinking](#support-claude-3-7-sonnet--thinking)
        - [Stream](#stream)
        - [Non-Stream](#non-stream)
      - [Support /v1/messages Claude Messages API](#support-v1messages-claude-messages-api)
        - [Support Claude Code](#support-claude-code)
    - [Support Claude 4.x Models](#support-claude-4x-models)
    - [Google (Gemini \& Vertex) Features](#google-gemini--vertex-features)
      - [Support gemini-2.0-flash-exp](#support-gemini-20-flash-exp)
      - [Support gemini-2.0-flash](#support-gemini-20-flash)
      - [Support gemini-2.0-flash-thinking-exp-01-21](#support-gemini-20-flash-thinking-exp-01-21)
      - [Support Vertex Imagen3](#support-vertex-imagen3)
      - [Support gemini multimodal output #2197](#support-gemini-multimodal-output-2197)
      - [Support gemini-2.5-pro](#support-gemini-25-pro)
      - [Support GCP Vertex gloabl region and gemini-2.5-pro-preview-06-05](#support-gcp-vertex-gloabl-region-and-gemini-25-pro-preview-06-05)
      - [Support gemini-2.5-flash-image-preview \& imagen-4 series](#support-gemini-25-flash-image-preview--imagen-4-series)
    - [OpenCode Support](#opencode-support)
    - [AWS Features](#aws-features)
      - [Support AWS cross-region inferences](#support-aws-cross-region-inferences)
      - [Support AWS BedRock Inference Profile](#support-aws-bedrock-inference-profile)
    - [Replicate Features](#replicate-features)
      - [Support replicate flux \& remix](#support-replicate-flux--remix)
      - [Support replicate chat models](#support-replicate-chat-models)
    - [DeepSeek Features](#deepseek-features)
      - [Support deepseek-reasoner](#support-deepseek-reasoner)
    - [OpenRouter Features](#openrouter-features)
      - [Support OpenRouter's reasoning content](#support-openrouters-reasoning-content)
    - [Coze Features](#coze-features)
      - [Support coze oauth authentication](#support-coze-oauth-authentication)
    - [XAI / Grok Features](#xai--grok-features)
      - [Support XAI/Grok Text \& Image Models](#support-xaigrok-text--image-models)
    - [Black Forest Labs Features](#black-forest-labs-features)
      - [Support black-forest-labs/flux-kontext-pro](#support-black-forest-labsflux-kontext-pro)
  - [Bug Fixes \& Enterprise-Grade Improvements (Including Security Enhancements)](#bug-fixes--enterprise-grade-improvements-including-security-enhancements)

## Tutorial

### Docker Compose Deployment

Run one-api using docker-compose:

```yaml
oneapi:
  image: ppcelery/one-api:latest
  restart: unless-stopped
  logging:
    driver: "json-file"
    options:
      max-size: "10m"
  environment:
    # --- Session & Security ---
    # (optional) SESSION_SECRET set a fixed session secret so that user sessions won't be invalidated after server restart
    SESSION_SECRET: xxxxxxx
    # (optional) ENABLE_COOKIE_SECURE enable secure cookies, must be used with HTTPS
    ENABLE_COOKIE_SECURE: "true"
    # (optional) COOKIE_MAXAGE_HOURS sets the session cookie's max age in hours. Default is `168` (7 days); adjust to control session lifetime.
    COOKIE_MAXAGE_HOURS: 168

    # --- Core Runtime ---
    # (optional) PORT override the listening port used by the HTTP server, default is `3000`
    PORT: 3000
    # (optional) GIN_MODE set Gin runtime mode; defaults to release when unset
    GIN_MODE: release
    # (optional) SHUTDOWN_TIMEOUT_SEC controls how long to wait for graceful shutdown and drains (seconds)
    SHUTDOWN_TIMEOUT_SEC: 360
    # (optional) DEBUG enable debug mode
    DEBUG: "true"
    # (optional) DEBUG_SQL display SQL logs
    DEBUG_SQL: "true"
    # (optional) ENABLE_PROMETHEUS_METRICS expose /metrics for Prometheus scraping when true
    ENABLE_PROMETHEUS_METRICS: "true"
    # (optional) LOG_RETENTION_DAYS set log retention days; default is not to delete any logs
    LOG_RETENTION_DAYS: 7
    # (optional) TRACE_RENTATION_DAYS retain trace records for the specified number of days; default is 30 and 0 disables cleanup
    TRACE_RENTATION_DAYS: 30

    # --- Storage & Cache ---
    # (optional) SQL_DSN set SQL database connection; leave empty to use SQLite (supports mysql, postgresql, sqlite3)
    SQL_DSN: "postgres://laisky:xxxxxxx@1.2.3.4/oneapi"
    # (optional) SQLITE_PATH override SQLite file path when SQL_DSN is empty
    SQLITE_PATH: "/data/one-api.db"
    # (optional) SQL_MAX_IDLE_CONNS tune database idle connection pool size
    SQL_MAX_IDLE_CONNS: 200
    # (optional) SQL_MAX_OPEN_CONNS tune database max open connections
    SQL_MAX_OPEN_CONNS: 2000
    # (optional) SQL_MAX_LIFETIME tune database connection lifetime in seconds
    SQL_MAX_LIFETIME: 300
    # (optional) REDIS_CONN_STRING set Redis cache connection
    REDIS_CONN_STRING: redis://100.122.41.16:6379/1
    # (optional) REDIS_PASSWORD set Redis password when authentication is required
    REDIS_PASSWORD: ""
    # (optional) SYNC_FREQUENCY refresh in-memory caches every N seconds when enabled
    SYNC_FREQUENCY: 600
    # (optional) MEMORY_CACHE_ENABLED force memory cache usage even without Redis
    MEMORY_CACHE_ENABLED: "true"

    # --- Usage & Billing ---
    # (optional) ENFORCE_INCLUDE_USAGE require upstream API responses to include usage field
    ENFORCE_INCLUDE_USAGE: "true"
    # (optional) PRECONSUME_TOKEN_FOR_BACKGROUND_REQUEST reserve quota for background requests that report usage later
    PRECONSUME_TOKEN_FOR_BACKGROUND_REQUEST: 15000
    # (optional) DEFAULT_MAX_TOKEN set the default maximum number of tokens for requests, default is 2048
    DEFAULT_MAX_TOKEN: 2048
    # (optional) DEFAULT_USE_MIN_MAX_TOKENS_MODEL opt-in to the min/max token contract for supported channels
    DEFAULT_USE_MIN_MAX_TOKENS_MODEL: "false"

    # --- Rate Limiting ---
    # (optional) GLOBAL_API_RATE_LIMIT maximum API requests per IP within three minutes, default is 1000
    GLOBAL_API_RATE_LIMIT: 1000
    # (optional) GLOBAL_WEB_RATE_LIMIT maximum web page requests per IP within three minutes, default is 1000
    GLOBAL_WEB_RATE_LIMIT: 1000
    # (optional) GLOBAL_RELAY_RATE_LIMIT /v1 API ratelimit for each token
    GLOBAL_RELAY_RATE_LIMIT: 1000
    # (optional) GLOBAL_CHANNEL_RATE_LIMIT whether to ratelimit per channel; 0 is unlimited, 1 enables rate limiting
    GLOBAL_CHANNEL_RATE_LIMIT: 1
    # (optional) CRITICAL_RATE_LIMIT tighten rate limits for admin-only APIs (seconds window matches defaults)
    CRITICAL_RATE_LIMIT: 20

    # --- Channel Automation ---
    # (optional) CHANNEL_SUSPEND_SECONDS_FOR_429 set the suspension duration (seconds) after receiving a 429 error, default is 60 seconds
    CHANNEL_SUSPEND_SECONDS_FOR_429: 60
    # (optional) CHANNEL_TEST_FREQUENCY run automatic channel health checks every N seconds (0 disables)
    CHANNEL_TEST_FREQUENCY: 0
    # (optional) BATCH_UPDATE_ENABLED enable background batch quota updater
    BATCH_UPDATE_ENABLED: "false"
    # (optional) BATCH_UPDATE_INTERVAL batch quota flush interval in seconds
    BATCH_UPDATE_INTERVAL: 5

    # --- Frontend & Proxies ---
    # (optional) FRONTEND_BASE_URL redirect page requests to specified address, server-side setting only
    FRONTEND_BASE_URL: https://oneapi.laisky.com
    # (optional) RELAY_PROXY forward upstream model calls through an HTTP proxy
    RELAY_PROXY: ""
    # (optional) USER_CONTENT_REQUEST_PROXY proxy for fetching user-provided assets
    USER_CONTENT_REQUEST_PROXY: ""
    # (optional) USER_CONTENT_REQUEST_TIMEOUT timeout (seconds) for fetching user assets
    USER_CONTENT_REQUEST_TIMEOUT: 30

    # --- Media & Pagination ---
    # (optional) MAX_ITEMS_PER_PAGE maximum items per page, default is 100
    MAX_ITEMS_PER_PAGE: 100
    # (optional) MAX_INLINE_IMAGE_SIZE_MB set the maximum allowed image size (in MB) for inlining images as base64, default is 30
    MAX_INLINE_IMAGE_SIZE_MB: 30

    # --- Integrations ---
    # (optional) OPENROUTER_PROVIDER_SORT set sorting method for OpenRouter Providers, default is throughput
    OPENROUTER_PROVIDER_SORT: throughput
    # (optional) LOG_PUSH_API set the API address for pushing error logs to external services
    LOG_PUSH_API: "https://gq.laisky.com/query/"
    LOG_PUSH_TYPE: "oneapi"
    LOG_PUSH_TOKEN: "xxxxxxx"

  volumes:
    - /var/lib/oneapi:/data
  ports:
    - 3000:3000
```

The initial default account and password are `root` / `123456`. Listening port can be configured via the `PORT` environment variable, default is `3000`.

> [!TIP]
>
> For production environments, consider using proper secret management solutions instead of hardcoding sensitive values in environment variables.

### Kubernetes Deployment

The Kubernetes deployment guide has been moved into a dedicated document:

- [docs/manuals/k8s.md](docs/manuals/k8s.md)

## Contributors

<a href="https://github.com/Laisky/one-api/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=Laisky/one-api" />
</a>

## New Features

### Universal Features

#### Support update user's remained quota

You can update the used quota using the API key of any token, allowing other consumption to be aggregated into the one-api for centralized management.

![](https://s3.laisky.com/uploads/2024/12/oneapi-update-quota.png)

#### Get request's cost

Each chat completion request will include a `X-Oneapi-Request-Id` in the returned headers. You can use this request id to request `GET /api/cost/request/:request_id` to get the cost of this request.

The returned structure is:

```go
type UserRequestCost struct {
  Id          int     `json:"id"`
  CreatedTime int64   `json:"created_time" gorm:"bigint"`
  UserID      int     `json:"user_id"`
  RequestID   string  `json:"request_id"`
  Quota       int64   `json:"quota"`
  CostUSD     float64 `json:"cost_usd" gorm:"-"`
}
```

#### Support Tracing info in logs

![](https://s3.laisky.com/uploads/2025/08/tracing.png)

#### Support Cached Input

Now supports cached input, which can significantly reduce the cost.

![](https://s3.laisky.com/uploads/2025/08/cached_input.png)

##### Support Anthropic Prompt caching

- <https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching>

#### Automatically Enable Thinking and Customize Reasoning Format via URL Parameters

Supports two URL parameters: `thinking` and `reasoning_format`.

- `thinking`: Whether to enable thinking mode, disabled by default.
- `reasoning_format`: Specifies the format of the returned reasoning.
  - `reasoning_content`: DeepSeek official API format, returned in the `reasoning_content` field.
  - `reasoning`: OpenRouter format, returned in the `reasoning` field.
  - `thinking`: Claude format, returned in the `thinking` field.

##### Reasoning Format - reasoning-content

![](https://s3.laisky.com/uploads/2025/02/reasoning_format-reasoning_content.png)

##### Reasoning Format - reasoning

![](https://s3.laisky.com/uploads/2025/02/reasoning_format-reasoning.png)

##### Reasoning Format - thinking

![](https://s3.laisky.com/uploads/2025/02/reasoning_format-thinking.png)

### OpenAI Features

#### (Merged) Support gpt-vision

#### Support openai images edits

- [feat: support openai images edits api #1369](https://github.com/songquanpeng/one-api/pull/1369)

![](https://s3.laisky.com/uploads/2024/12/oneapi-image-edit.png)

#### Support OpenAI o1/o1-mini/o1-preview

- [feat: add openai o1 #1990](https://github.com/songquanpeng/one-api/pull/1990)

#### Support gpt-4o-audio

- [feat: support gpt-4o-audio #2032](https://github.com/songquanpeng/one-api/pull/2032)

![](https://s3.laisky.com/uploads/2025/01/oneapi-audio-1.png)

![](https://s3.laisky.com/uploads/2025/01/oneapi-audio-2.png)

#### Support OpenAI web search models

- [feature: support openai web search models #2189](https://github.com/songquanpeng/one-api/pull/2189)

support `gpt-4o-search-preview` & `gpt-4o-mini-search-preview`

![](https://s3.laisky.com/uploads/2025/03/openai-websearch-models-1.png)

![](https://s3.laisky.com/uploads/2025/03/openai-websearch-models-2.png)

#### Support gpt-image-1's image generation & edits

![](https://s3.laisky.com/uploads/2025/04/gpt-image-1-2.png)

![](https://s3.laisky.com/uploads/2025/04/gpt-image-1-3.png)

![](https://s3.laisky.com/uploads/2025/04/gpt-image-1-1.png)

#### Support o3-mini & o3 & o4-mini & gpt-4.1 & o3-pro & reasoning content

- [feat: extend support for o3 models and update model ratios #2048](https://github.com/songquanpeng/one-api/pull/2048)

![](https://s3.laisky.com/uploads/2025/06/o3-pro.png)

#### Support OpenAI Response API

**Partially supported, still in development.**

![](https://s3.laisky.com/uploads/2025/07/response-api.png)

#### Support gpt-5 family

gpt-5-chat-latest / gpt-5 / gpt-5-mini / gpt-5-nano / gpt-5-codex / gpt-5-pro

#### Support o3-deep-research & o4-mini-deep-research

![](https://s3.laisky.com/uploads/2025/09/o4-mini-deep-research.png)

#### Support Codex Cli

```sh
# vi $HOME/.codex/config.toml

model = "gemini-2.5-flash"
model_provider = "laisky"

[model_providers.laisky]
# Name of the provider that will be displayed in the Codex UI.
name = "Laisky"
# The path `/chat/completions` will be amended to this URL to make the POST
# request for the chat completions.
base_url = "https://oneapi.laisky.com/v1"
# If `env_key` is set, identifies an environment variable that must be set when
# using Codex with this provider. The value of the environment variable must be
# non-empty and will be used in the `Bearer TOKEN` HTTP header for the POST request.
env_key = "sk-xxxxxxx"
# Valid values for wire_api are "chat" and "responses". Defaults to "chat" if omitted.
wire_api = "responses"
# If necessary, extra query params that need to be added to the URL.
# See the Azure example below.
query_params = {}

```

### Anthropic (Claude) Features

#### (Merged) Support aws claude

- [feat: support aws bedrockruntime claude3 #1328](https://github.com/songquanpeng/one-api/pull/1328)
- [feat: add new claude models #1910](https://github.com/songquanpeng/one-api/pull/1910)

![](https://s3.laisky.com/uploads/2024/12/oneapi-claude.png)

#### Support claude-3-7-sonnet & thinking

- [feat: support claude-3-7-sonnet #2143](https://github.com/songquanpeng/one-api/pull/2143/files)
- [feat: support claude thinking #2144](https://github.com/songquanpeng/one-api/pull/2144)

By default, the thinking mode is not enabled. You need to manually pass the `thinking` field in the request body to enable it.

##### Stream

![](https://s3.laisky.com/uploads/2025/02/claude-thinking.png)

##### Non-Stream

![](https://s3.laisky.com/uploads/2025/02/claude-thinking-non-stream.png)

#### Support /v1/messages Claude Messages API

![](https://s3.laisky.com/uploads/2025/07/claude_messages.png)

##### Support Claude Code

```sh
export ANTHROPIC_MODEL="openai/gpt-oss-120b"
export ANTHROPIC_BASE_URL="https://oneapi.laisky.com/"
export ANTHROPIC_AUTH_TOKEN="sk-xxxxxxx"
```

You can use any model you like for Claude Code, even if the model doesn’t natively support the Claude Messages API.

### Support Claude 4.x Models

![](https://s3.laisky.com/uploads/2025/09/claude-sonnet-4-5.png)

claude-opus-4-0 / claude-opus-4-1 / claude-sonnet-4-0 / claude-sonnet-4-5 / claude-haiku-4-5

### Google (Gemini & Vertex) Features

#### Support gemini-2.0-flash-exp

- [feat: add gemini-2.0-flash-exp #1983](https://github.com/songquanpeng/one-api/pull/1983)

![](https://s3.laisky.com/uploads/2024/12/oneapi-gemini-flash.png)

#### Support gemini-2.0-flash

- [feat: support gemini-2.0-flash #2055](https://github.com/songquanpeng/one-api/pull/2055)

#### Support gemini-2.0-flash-thinking-exp-01-21

- [feature: add deepseek-reasoner & gemini-2.0-flash-thinking-exp-01-21 #2045](https://github.com/songquanpeng/one-api/pull/2045)

#### Support Vertex Imagen3

- [feat: support vertex imagen3 #2030](https://github.com/songquanpeng/one-api/pull/2030)

![](https://s3.laisky.com/uploads/2025/01/oneapi-imagen3.png)

#### Support gemini multimodal output #2197

- [feature: support gemini multimodal output #2197](https://github.com/songquanpeng/one-api/pull/2197)

![](https://s3.laisky.com/uploads/2025/03/gemini-multimodal.png)

#### Support gemini-2.5-pro

#### Support GCP Vertex gloabl region and gemini-2.5-pro-preview-06-05

![](https://s3.laisky.com/uploads/2025/06/gemini-2.5-pro-preview-06-05.png)

#### Support gemini-2.5-flash-image-preview & imagen-4 series

![](https://s3.laisky.com/uploads/2025/09/gemini-banana.png)

### OpenCode Support

<p align="center">
  <a href="https://opencode.ai">
    <picture>
      <source srcset="https://github.com/sst/opencode/raw/dev/packages/console/app/src/asset/logo-ornate-dark.svg" media="(prefers-color-scheme: dark)">
      <source srcset="https://github.com/sst/opencode/raw/dev/packages/console/app/src/asset/logo-ornate-light.svg" media="(prefers-color-scheme: light)">
      <img src="https://github.com/sst/opencode/raw/dev/packages/console/app/src/asset/logo-ornate-light.svg" alt="OpenCode logo">
    </picture>
  </a>
</p>

[opencode.ai](https://opencode.ai) is an AI coding agent built for the terminal. OpenCode is fully open source, giving you control and `freedom` to use any provider, any model, and any editor. It's available as both a CLI and TUI.

One‑API integrates seamlessly with OpenCode: you can connect any One‑API endpoint and use all your unified models through OpenCode's interface (both CLI and TUI).

To get started, create or edit `~/.config/opencode/opencode.json` like this:

**Using OpenAI SDK:**

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "one-api": {
      "npm": "@ai-sdk/openai",
      "name": "One API",
      "options": {
        "baseURL": "https://oneapi.laisky.com/v1",
        "apiKey": "<ONEAPI_TOKEN_KEY>"
      },
      "models": {
        "gpt-4.1-2025-04-14": {
          "name": "GPT 4.1"
        }
      }
    }
  }
}
```

**Using Anthropic SDK:**

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "one-api-anthropic": {
      "npm": "@ai-sdk/anthropic",
      "name": "One API (Anthropic)",
      "options": {
        "baseURL": "https://oneapi.laisky.com/v1",
        "apiKey": "<ONEAPI_TOKEN_KEY>"
      },
      "models": {
        "claude-sonnet-4-5": {
          "name": "Claude Sonnet 4.5"
        }
      }
    }
  }
}
```

### AWS Features

#### Support AWS cross-region inferences

- [fix: support aws cross region inferences #2182](https://github.com/songquanpeng/one-api/pull/2182)

#### Support AWS BedRock Inference Profile

![](https://s3.laisky.com/uploads/2025/07/aws-inference-profile.png)

### Replicate Features

#### Support replicate flux & remix

- [feature: 支持 replicate 的绘图 #1954](https://github.com/songquanpeng/one-api/pull/1954)
- [feat: image edits/inpaiting 支持 replicate 的 flux remix #1986](https://github.com/songquanpeng/one-api/pull/1986)

![](https://s3.laisky.com/uploads/2024/12/oneapi-replicate-1.png)

![](https://s3.laisky.com/uploads/2024/12/oneapi-replicate-2.png)

![](https://s3.laisky.com/uploads/2024/12/oneapi-replicate-3.png)

#### Support replicate chat models

- [feat: 支持 replicate chat models #1989](https://github.com/songquanpeng/one-api/pull/1989)

### DeepSeek Features

#### Support deepseek-reasoner

- [feature: add deepseek-reasoner & gemini-2.0-flash-thinking-exp-01-21 #2045](https://github.com/songquanpeng/one-api/pull/2045)

### OpenRouter Features

#### Support OpenRouter's reasoning content

- [feat: support OpenRouter reasoning #2108](https://github.com/songquanpeng/one-api/pull/2108)

By default, the thinking mode is automatically enabled for the deepseek-r1 model, and the response is returned in the open-router format.

![](https://s3.laisky.com/uploads/2025/02/openrouter-reasoning.png)

### Coze Features

#### Support coze oauth authentication

- [feat: support coze oauth authentication](https://github.com/Laisky/one-api/pull/52)

### XAI / Grok Features

#### Support XAI/Grok Text & Image Models

![](https://s3.laisky.com/uploads/2025/08/groq.png)

### Black Forest Labs Features

#### Support black-forest-labs/flux-kontext-pro

![](https://s3.laisky.com/uploads/2025/05/flux-kontext-pro.png)

## Bug Fixes & Enterprise-Grade Improvements (Including Security Enhancements)

- [BUGFIX: Several issues when updating tokens #1933](https://github.com/songquanpeng/one-api/pull/1933)
- [feat(audio): count whisper-1 quota by audio duration #2022](https://github.com/songquanpeng/one-api/pull/2022)
- [fix: Fix issue where high-quota users using low-quota tokens aren't pre-charged, causing large token deficits under high concurrency #25](https://github.com/Laisky/one-api/pull/25)
- [fix: channel test false negative #2065](https://github.com/songquanpeng/one-api/pull/2065)
- [fix: resolve "bufio.Scanner: token too long" error by increasing buffer size #2128](https://github.com/songquanpeng/one-api/pull/2128)
- [feat: Enhance VolcEngine channel support with bot model #2131](https://github.com/songquanpeng/one-api/pull/2131)
- [fix: models API returns models in deactivated channels #2150](https://github.com/songquanpeng/one-api/pull/2150)
- [fix: Automatically close channel when connection fails](https://github.com/Laisky/one-api/pull/34)
- [fix: update EmailDomainWhitelist submission logic #33](https://github.com/Laisky/one-api/pull/33)
- [fix: send ByAll](https://github.com/Laisky/one-api/pull/35)
- [fix: oidc token endpoint request body #2106 #36](https://github.com/Laisky/one-api/pull/36)

> [!NOTE]
>
> For additional enterprise-grade improvements, including security enhancements (e.g., [vulnerability fixes](https://github.com/Laisky/one-api/pull/126)), you can also view these pull requests [here](https://github.com/Laisky/one-api/pulls?q=is%3Apr+is%3Aclosed).
