# API Format Conversion Guide

This guide explains how one-api transparently converts between different AI API formats, enabling users and developers to interact with a wide range of models using familiar interfaces. It covers both usage guidelines and implementation details.

- [API Format Conversion Guide](#api-format-conversion-guide)
  - [1. Introduction](#1-introduction)
  - [2. Usage Overview](#2-usage-overview)
    - [2.1 Supported API Styles](#21-supported-api-styles)
    - [2.2 How It Works for Users](#22-how-it-works-for-users)
  - [3. Conversion Architecture](#3-conversion-architecture)
    - [3.1 High-Level Flow](#31-high-level-flow)
  - [Overview](#overview)
  - [4. User-Facing Behavior \& Guidelines](#4-user-facing-behavior--guidelines)
    - [4.1 Model/Endpoint Selection](#41-modelendpoint-selection)
    - [4.2 Streaming, Function Calling, and Structured Content](#42-streaming-function-calling-and-structured-content)
    - [4.3 Example Usage](#43-example-usage)
  - [5. Implementation Details (For Developers)](#5-implementation-details-for-developers)
    - [5.1 OpenAI ↔ Response API Conversion](#51-openai--response-api-conversion)
    - [5.2 Claude Messages API Conversion](#52-claude-messages-api-conversion)
    - [5.3 Model Support Detection](#53-model-support-detection)
    - [5.4 Response API Fallback via ChatCompletion](#54-response-api-fallback-via-chatcompletion)
  - [6. Data Structure Mappings](#6-data-structure-mappings)
    - [6.1 Request Conversion](#61-request-conversion)
    - [6.2 Response Conversion (Non-streaming)](#62-response-conversion-non-streaming)
    - [6.3 Response Conversion (Streaming)](#63-response-conversion-streaming)
    - [6.4 Request Format Mapping](#64-request-format-mapping)
    - [6.5 Response Format Mapping](#65-response-format-mapping)
    - [6.6 Status Mapping](#66-status-mapping)
  - [7. Advanced Features](#7-advanced-features)
    - [7.1 Function Calling Support](#71-function-calling-support)
    - [7.2 Streaming Implementation](#72-streaming-implementation)
    - [7.3 Example Conversion](#73-example-conversion)
    - [7.4 Event Processing](#74-event-processing)
    - [7.5 Deduplication Strategy](#75-deduplication-strategy)
  - [8. Error Handling \& Fallbacks](#8-error-handling--fallbacks)
    - [8.1 Parse Errors](#81-parse-errors)
    - [8.2 API Errors](#82-api-errors)
    - [8.3 Fallback Mechanisms](#83-fallback-mechanisms)
  - [9. Testing \& Extensibility](#9-testing--extensibility)
    - [9.1 Test Coverage](#91-test-coverage)
    - [9.2 Integration Tests](#92-integration-tests)
  - [10. Context Management \& Configuration](#10-context-management--configuration)
    - [10.1 Context Keys](#101-context-keys)
    - [10.2 Context Flow](#102-context-flow)
    - [10.3 Model Support Integration](#103-model-support-integration)
      - [Current Implementation](#current-implementation)
      - [Model Categories](#model-categories)
        - [ChatCompletion-Only Models (API: `/v1/chat/completions`)](#chatcompletion-only-models-api-v1chatcompletions)
        - [Response API Compatible Models (API: `/v1/responses`)](#response-api-compatible-models-api-v1responses)
      - [Integration Points](#integration-points)
        - [1. Request Processing](#1-request-processing)
        - [2. URL Generation](#2-url-generation)
      - [Implementation Strategy](#implementation-strategy)
  - [11. Performance Considerations](#11-performance-considerations)
    - [11.1 Memory Management](#111-memory-management)
    - [11.2 Processing Efficiency](#112-processing-efficiency)
  - [12. Future Enhancements](#12-future-enhancements)
    - [12.1 Dynamic Model Support Detection](#121-dynamic-model-support-detection)
    - [12.2 Enhanced Error Recovery](#122-enhanced-error-recovery)
    - [12.3 Performance Optimizations](#123-performance-optimizations)
  - [13. Configuration Reference](#13-configuration-reference)
    - [13.1 Channel Type Detection](#131-channel-type-detection)
    - [13.2 Relay Mode Detection](#132-relay-mode-detection)
  - [14. Summary](#14-summary)
- [Claude Messages API Conversion Architecture](#claude-messages-api-conversion-architecture)
  - [Overview](#overview-1)
  - [Recent Changes (2025-08)](#recent-changes-2025-08)
  - [Problem Statement](#problem-statement)
  - [Architecture](#architecture)
    - [High-Level Flow](#high-level-flow)
    - [Key Components](#key-components)
      - [1. Claude Messages Controller](#1-claude-messages-controller)
      - [2. Adapter Conversion Interface](#2-adapter-conversion-interface)
      - [3. Shared OpenAI-Compatible Conversion](#3-shared-openai-compatible-conversion)
  - [Adapter Implementation Patterns](#adapter-implementation-patterns)
    - [Pattern 1: Native Claude Support (Anthropic)](#pattern-1-native-claude-support-anthropic)
    - [Pattern 2: OpenAI-Compatible Conversion](#pattern-2-openai-compatible-conversion)
    - [Pattern 3: Custom Conversion (Gemini)](#pattern-3-custom-conversion-gemini)
  - [Supported Adapters](#supported-adapters)
    - [✅ Native Claude Messages Support](#-native-claude-messages-support)
    - [✅ OpenAI-Compatible Conversion](#-openai-compatible-conversion)
    - [✅ Custom Conversion](#-custom-conversion)
    - [❌ Limited or No Support](#-limited-or-no-support)
    - [📋 Test Results Summary](#-test-results-summary)
  - [Data Structure Mappings](#data-structure-mappings)
    - [Claude Messages to OpenAI Format](#claude-messages-to-openai-format)
    - [Message Content Conversion](#message-content-conversion)
      - [Text Content](#text-content)
      - [Structured Content](#structured-content)
      - [Tool Use](#tool-use)
    - [Response Format Conversion](#response-format-conversion)
      - [OpenAI to Claude Messages](#openai-to-claude-messages)
      - [Finish Reason Mapping](#finish-reason-mapping)
  - [Context Management](#context-management)
    - [Context Keys](#context-keys)
    - [Context Flow](#context-flow)
  - [Error Handling](#error-handling)
    - [Conversion Errors](#conversion-errors)
    - [Adapter Errors](#adapter-errors)
    - [Fallback Mechanisms](#fallback-mechanisms)
  - [Performance Considerations](#performance-considerations)
    - [Memory Management](#memory-management)
    - [Processing Efficiency](#processing-efficiency)
  - [Testing](#testing)
    - [Test Coverage](#test-coverage)
  - [Future Enhancements](#future-enhancements)
    - [1. Enhanced Content Support](#1-enhanced-content-support)
    - [2. Performance Optimizations](#2-performance-optimizations)
    - [3. Extended Adapter Support](#3-extended-adapter-support)
  - [Summary](#summary)

## 1. Introduction

one-api provides seamless compatibility between OpenAI, Claude, Gemini, and other AI APIs. It automatically converts requests and responses between formats, so you can use your preferred API style regardless of the underlying model or provider.

## 2. Usage Overview

### 2.1 Supported API Styles

- **OpenAI ChatCompletion API** (`/v1/chat/completions`)
- **OpenAI Response API** (`/v1/responses`)
- **Claude Messages API** (`/v1/messages`)

### 2.2 How It Works for Users

- You send requests in your preferred format (e.g., OpenAI ChatCompletion or Claude Messages).
- one-api detects the target model/provider and automatically converts the request/response as needed.
- You receive responses in the same format you used for your request.

**Example:**

- You can use the Claude Messages API to access OpenAI-compatible models, or use OpenAI's API to access Claude models, with no code changes on your side.

## 3. Conversion Architecture

### 3.1 High-Level Flow

```plaintext
User Request (Any Supported API)
  ↓
[Route to Appropriate Controller]
  ↓
[Detect Model/Provider]
  ↓
[Convert Request Format if Needed]
  ↓
[Send to Upstream]
  ↓
[Convert Response Format if Needed]
  ↓
User Response (Original API Format)
```

## Overview

The system supports:

- **OpenAI ↔ Response API**: Converts between ChatCompletion and Response API formats.
- **Claude Messages API**: Converts Claude Messages requests/responses to/from OpenAI, Gemini, and other formats.

## 4. User-Facing Behavior & Guidelines

### 4.1 Model/Endpoint Selection

- For most models, you can use either the ChatCompletion or Claude Messages API.
- Some models (e.g., OpenAI search models) only support ChatCompletion; others (e.g., Claude) only support Claude Messages.
- one-api automatically routes and converts as needed.

### 4.2 Streaming, Function Calling, and Structured Content

- Streaming, function calling, and structured content are fully supported and converted between formats.
- Usage (token counting) is as accurate as possible, including tool/function arguments.

### 4.3 Example Usage

**Using Claude Messages API with OpenAI-compatible models:**

```sh
export ANTHROPIC_MODEL="openai/gpt-4o"
export ANTHROPIC_BASE_URL="https://oneapi.laisky.com/"
export ANTHROPIC_AUTH_TOKEN="sk-xxxxxxx"
```

## 5. Implementation Details (For Developers)

### 5.1 OpenAI ↔ Response API Conversion

**Location:** `relay/adaptor/openai/adaptor.go`

**Key Logic:**

- Converts ChatCompletion requests to Response API format for compatible models.
- Stores converted request in context for response detection.

**Key Condition:**

- Only converts when relay mode is ChatCompletion and channel type is OpenAI, and the model supports Response API.

### 5.2 Claude Messages API Conversion

**Location:** `relay/controller/claude_messages.go`, `relay/adaptor/openai_compatible/claude_messages.go`

**Key Logic:**

- All OpenAI-compatible adapters use a shared handler (`HandleClaudeMessagesResponse`) for both streaming and non-streaming Claude Messages conversion.
- Streaming conversion uses `ConvertOpenAIStreamToClaudeSSE` to emit Claude-native SSE events and accumulate all relevant content for usage calculation.
- Usage calculation includes both text and tool call arguments.

**Controller Fallback Logic:**

- The controller attempts to extract usage from the Claude response body (JSON), then from SSE (stream), and finally falls back to prompt-only estimation if all else fails.

### 5.3 Model Support Detection

**Function:** `IsModelsOnlySupportedByChatCompletionAPI(model string) bool` (see `relay/adaptor/openai/response_model.go`)

**Behavior:**

- Returns true for models that only support ChatCompletion (e.g., search models), false otherwise.

**Integration Points:** Used in request processing and URL generation to determine conversion logic.

1. **Request Processing** - `adaptor.go:117`:

```go
if relayMode == relaymode.ChatCompletions &&
   meta.ChannelType == channeltype.OpenAI &&
   !IsModelsOnlySupportedByChatCompletionAPI(meta.ActualModelName) {
    // Proceed with conversion
}
```

2. **URL Generation** - `adaptor.go:84`:

```go
if meta.Mode == relaymode.ChatCompletions &&
   meta.ChannelType == channeltype.OpenAI &&
   !IsModelsOnlySupportedByChatCompletionAPI(meta.ActualModelName) {
    responseAPIPath := "/v1/responses"
    return GetFullRequestURL(meta.BaseURL, responseAPIPath, meta.ChannelType), nil
}
return GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
```

**Current Behavior:**

- ✅ **Search Models**: Models containing "gpt" and "-search-" use ChatCompletion API (`/v1/chat/completions`)
- ✅ **Regular Models**: All other models use Response API (`/v1/responses`)
- ✅ **URL Consistency**: Endpoint selection matches conversion logic
- ✅ **Test Coverage**: Comprehensive tests verify both URL generation and conversion consistency

### 5.4 Response API Fallback via ChatCompletion

- **When it triggers:** If a `/v1/responses` request targets a channel whose upstream does not yet speak the Response API (e.g., Azure, third-party OpenAI compatibles), the controller routes the call through `relayResponseAPIThroughChat`.
- **Request conversion:** `ConvertResponseAPIToChatCompletionRequest` collapses text-only `input` arrays back into simple message strings, retains multimodal items (images/audio) as structured content, preserves tool definitions/reasoning settings, and rejects currently unsupported `background` or prompt template fields. Streaming is disabled on this path.
- **Response rewriting:** After the adapter completes the ChatCompletion call, the controller registers `ctxkey.ResponseRewriteHandler`, which takes the upstream `SlimTextResponse` and rehydrates a Response API envelope (status, output array, usage). The original payload is stored under `ctxkey.ResponseAPIRequestOriginal` so instructions, text formatting, and tool-choice metadata can be echoed in the response.
- **Quota + metrics:** The fallback shares the ChatCompletion pre-consume / post-billing flow and reconciles provisional quota once upstream usage arrives, ensuring billing parity with native Response API calls.
- **Regression coverage:** `TestConvertResponseAPIToChatCompletionRequest` verifies conversion fidelity, while `TestRenderChatResponseAsResponseAPI` ensures the rewrite handler emits valid Response API JSON.

## 6. Data Structure Mappings

### 6.1 Request Conversion

**Function:** `ConvertChatCompletionToResponseAPI()`

**Key Transformations:**

- Messages → Input array
- System message → Instructions field
- Tools → Response API tool format
- Function call history → Text summaries
- Parameters mapping (temperature, top_p, etc.)

**Function Call History Handling:**
The Response API doesn't support ChatCompletion's function call history format. The converter creates text summaries:

```plaintext
Previous function calls:
- Called get_current_datetime({}) → {"year":2025,"month":6,"day":12}
- Called get_weather({"location":"Boston"}) → {"temperature":22,"condition":"sunny"}
```

> **Fallback note:** When the controller must relay a Response API request through a ChatCompletion-only upstream, `ConvertResponseAPIToChatCompletionRequest` performs the inverse mapping—collapsing text-only segments back into message strings, preserving multimodal content, and carrying over tool definitions / reasoning settings.

### 6.2 Response Conversion (Non-streaming)

**Function:** `ConvertResponseAPIToChatCompletion()`

**Handler:** `ResponseAPIHandler()`

**Key Transformations:**

- Output array → Choices array
- Message content → Choice message content
- Function calls → Tool calls
- Status → Finish reason
- Usage field mapping

### 6.3 Response Conversion (Streaming)

**Function:** `ConvertResponseAPIStreamToChatCompletion()`

**Handler:** `ResponseAPIStreamHandler()`

**Stream Event Processing:**

- `response.output_text.delta` → Content deltas
- `response.reasoning_summary_text.delta` → Reasoning deltas
- `response.completed` → Usage information
- Function call events → Tool call deltas

### 6.4 Request Format Mapping

| ChatCompletion Field   | Response API Field  | Notes                                |
| ---------------------- | ------------------- | ------------------------------------ |
| `messages`             | `input`             | Array of message objects             |
| `messages[0]` (system) | `instructions`      | System message moved to instructions |
| `tools`                | `tools`             | Tool format conversion required      |
| `max_tokens`           | `max_output_tokens` | Direct mapping                       |
| `temperature`          | `temperature`       | Direct mapping                       |
| `stream`               | `stream`            | Direct mapping                       |
| `user`                 | `user`              | Direct mapping                       |

### 6.5 Response Format Mapping

| Response API Field              | ChatCompletion Field           | Notes             |
| ------------------------------- | ------------------------------ | ----------------- |
| `output[].content[].text`       | `choices[].message.content`    | Text content      |
| `output[].summary[].text`       | `choices[].message.reasoning`  | Reasoning content |
| `output[].type="function_call"` | `choices[].message.tool_calls` | Function calls    |
| `status`                        | `choices[].finish_reason`      | Status mapping    |
| `usage.input_tokens`            | `usage.prompt_tokens`          | Token usage       |
| `usage.output_tokens`           | `usage.completion_tokens`      | Token usage       |

### 6.6 Status Mapping

| Response API Status | ChatCompletion finish_reason | Notes                                    |
| ------------------- | ---------------------------- | ---------------------------------------- |
| `completed`         | `stop` or `tool_calls`       | `tool_calls` when function calls present |
| `failed`            | `stop`                       |                                          |
| `incomplete`        | `length`                     |                                          |
| `cancelled`         | `stop`                       |                                          |

## 7. Advanced Features

### 7.1 Function Calling Support

1. ChatCompletion tools → Response API tools (format conversion)
2. Function call history → Text summaries in input
3. Tool choice → Tool choice (preserved)

### 7.2 Streaming Implementation

1. Response API function_call output → ChatCompletion tool_calls
2. Call ID mapping with prefix handling (`fc_` ↔ `call_`)
3. Function name and arguments preservation
4. Finish reason set to `tool_calls` when functions present

Unified OpenAI-compatible streaming handler:

- All OpenAI-compatible adapters delegate streaming to a shared handler in `relay/adaptor/openai_compatible`.
- Optional thinking extraction is controlled via URL parameter `?thinking=true` and the extracted reasoning is mapped to the field specified by `?reasoning_format=` (supports `reasoning_content`, `reasoning`, and `thinking`).
- When upstream usage is missing/partial, token usage is computed from streamed text plus tool call arguments.

### 7.3 Example Conversion

**Input (ChatCompletion)**:

```json
{
  "model": "gpt-4",
  "messages": [
    { "role": "user", "content": "What's the weather?" },
    {
      "role": "assistant",
      "tool_calls": [
        { "id": "call_123", "function": { "name": "get_weather" } }
      ]
    },
    { "role": "tool", "tool_call_id": "call_123", "content": "Sunny, 22°C" }
  ]
}
```

**Converted to Response API**:

```json
{
  "model": "gpt-4",
  "input": [
    { "role": "user", "content": "What's the weather?" },
    {
      "role": "assistant",
      "content": "Previous function calls:\n- Called get_weather() → Sunny, 22°C"
    }
  ]
}
```

### 7.4 Event Processing

The streaming handler processes different event types:

- **Delta Events**: `response.output_text.delta`, `response.reasoning_summary_text.delta`

  - Converted to ChatCompletion streaming chunks
  - Content accumulated for token counting

- **Completion Events**: `response.output_text.done`, `response.content_part.done`

  - Discarded to prevent duplicate content
  - Only usage information from `response.completed` is forwarded

- **Function Call Events**: Function call streaming support
  - Converted to tool_call deltas in ChatCompletion format

### 7.5 Deduplication Strategy

Response API emits both delta and completion events. The implementation:

1. Only processes delta events for content streaming
2. Discards completion events to prevent duplication
3. Forwards usage information from final completion events

## 8. Error Handling & Fallbacks

### 8.1 Parse Errors

- Request conversion errors wrapped with `ErrorWrapper()`
- Response parsing errors logged and processing continues
- Malformed chunks skipped with debug logging

### 8.2 API Errors

- Response API errors passed through unchanged
- Error format preserved for client compatibility

### 8.3 Fallback Mechanisms

- Token usage calculation fallback when API doesn't provide usage
- Content extraction fallback for malformed responses
- Response API requests automatically downgrade to ChatCompletion when a channel lacks native Response API support; responses are rewritten back through `ResponseRewriteHandler` to keep client contracts intact.

## 9. Testing & Extensibility

### 9.1 Test Coverage

**Location**: `relay/adaptor/openai/response_model_test.go`

**Key Test Categories**:

- `TestConvertChatCompletionToResponseAPI()` - Request conversion
- `TestConvertResponseAPIToChatCompletion()` - Response conversion
- `TestConvertResponseAPIStreamToChatCompletion()` - Streaming conversion
- `TestFunctionCallWorkflow()` - End-to-end function calling
- `TestChannelSpecificConversion()` - Channel type filtering
- `TestConvertResponseAPIToChatCompletionRequest()` - Fallback request conversion coverage
- `TestRenderChatResponseAsResponseAPI()` - ChatCompletion→Response rewrite validation

### 9.2 Integration Tests

**Location**: `relay/adaptor/openai/channel_conversion_test.go`

Tests conversion behavior for different channel types:

- OpenAI: Conversion enabled
- Azure, AI360, etc.: Conversion disabled

## 10. Context Management & Configuration

### 10.1 Context Keys

**Location**: `common/ctxkey/key.go`

**Key Constant**: `ConvertedRequest = "converted_request"`

**Usage**:

- Request phase: Store converted ResponseAPI request
- Response phase: Detect need for response conversion
- Response API fallback registers `ResponseRewriteHandler` (rewrapper callback) and `ResponseAPIRequestOriginal` (original payload snapshot) so the controller can reshape upstream ChatCompletion responses back into Response API format.

### 10.2 Context Flow

1. **Request**: `c.Set(ctxkey.ConvertedRequest, responseAPIRequest)`
2. **Response**: `c.Get(ctxkey.ConvertedRequest)` to detect conversion need

### 10.3 Model Support Integration

#### Current Implementation

✅ **Function**: `IsModelsOnlySupportedByChatCompletionAPI(modelName string) bool`
**Location**: `relay/adaptor/openai/response_model.go:15`

**Model Detection Logic**:

```go
func IsModelsOnlySupportedByChatCompletionAPI(actualModel string) bool {
	switch {
	case strings.Contains(actualModel, "gpt") && strings.Contains(actualModel, "-search-"):
		return true
	default:
		return false
	}
}
```

#### Model Categories

##### ChatCompletion-Only Models (API: `/v1/chat/completions`)

These models return `true` from `IsModelsOnlySupportedByChatCompletionAPI()`:

- ✅ **Search Models**: `gpt-4-search-*`, `gpt-4o-search-*`, `gpt-3.5-turbo-search-*`
- 🔍 **Pattern**: Contains both "gpt" and "-search-"
- 📍 **Endpoint**: `https://api.openai.com/v1/chat/completions`
- 🔄 **Conversion**: **Disabled** - Request stays in ChatCompletion format

##### Response API Compatible Models (API: `/v1/responses`)

These models return `false` from `IsModelsOnlySupportedByChatCompletionAPI()`:

- ✅ **Regular GPT Models**: `gpt-4`, `gpt-4o`, `gpt-3.5-turbo`
- ✅ **Reasoning Models**: `o1-preview`, `o1-mini`, `o3`
- ✅ **All Other Models**: Any model not matching the ChatCompletion-only pattern
- 📍 **Endpoint**: `https://api.openai.com/v1/responses`
- 🔄 **Conversion**: **Enabled** - ChatCompletion → Response API → ChatCompletion

#### Integration Points

##### 1. Request Processing

**Location**: `relay/adaptor/openai/adaptor.go:117`

**✅ Current Implementation**:

```go
if relayMode == relaymode.ChatCompletions &&
   meta.ChannelType == channeltype.OpenAI &&
   !IsModelsOnlySupportedByChatCompletionAPI(meta.ActualModelName) {
   // Proceed with Response API conversion
}
```

##### 2. URL Generation

**Location**: `relay/adaptor/openai/adaptor.go:84`

**✅ Current Implementation**:

```go
if meta.Mode == relaymode.ChatCompletions &&
   meta.ChannelType == channeltype.OpenAI &&
   !IsModelsOnlySupportedByChatCompletionAPI(meta.ActualModelName) {
   responseAPIPath := "/v1/responses"
   return GetFullRequestURL(meta.BaseURL, responseAPIPath, meta.ChannelType), nil
}
return GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
```

#### Implementation Strategy

✅ **Completed**:

1. ✅ Function implementation with search model detection
2. ✅ Integration in request conversion logic
3. ✅ Integration in URL generation logic
4. ✅ Comprehensive test coverage
5. ✅ Documentation updates

🔄 **Future Enhancements**:

1. **Dynamic Model Detection**: API-based model capability queries
2. **Configuration-Driven**: External configuration for model support mapping
3. **Runtime Updates**: Dynamic model support updates without code changes
4. **Enhanced Patterns**: More sophisticated model pattern matching

## 11. Performance Considerations

### 11.1 Memory Management

- Streaming buffers: 1MB buffer for large messages
- Content accumulation: Separate tracking for reasoning vs content
- Context storage: Minimal object stored in gin context

### 11.2 Processing Efficiency

- Single-pass conversion: Request and response converted once
- Lazy evaluation: Conversion only when needed
- Early detection: Context check before processing

## 12. Future Enhancements

### 12.1 Dynamic Model Support Detection

- API-based model capability detection
- Configuration-driven model support mapping
- Runtime model support updates

### 12.2 Enhanced Error Recovery

- Partial response recovery for streaming failures
- Automatic fallback to ChatCompletion for unsupported features

### 12.3 Performance Optimizations

- Response format detection optimization
- Memory usage optimization for large responses
- Caching for repeated conversions

## 13. Configuration Reference

### 13.1 Channel Type Detection

**Location**: `relay/channeltype/define.go`

**OpenAI Channel Type**: `channeltype.OpenAI = 1`

### 13.2 Relay Mode Detection

**Location**: `relay/relaymode/`

**ChatCompletion Mode**: `relaymode.ChatCompletions`

## 14. Summary

The API conversion system provides transparent, bidirectional conversion between ChatCompletion, Response API, and Claude Messages formats, enabling:

1. **Backward Compatibility**: Users can continue using ChatCompletion API
2. **Forward Compatibility**: Access to Response API features and models
3. **Selective Conversion**: Model-specific conversion control
4. **Full Feature Support**: Function calling, streaming, reasoning content
5. **Error Resilience**: Comprehensive error handling and fallbacks

The implementation maintains familiar API interfaces for users, while leveraging advanced capabilities and compatibility under the hood. Both users and developers benefit from seamless integration, robust error handling, and extensibility.

# Claude Messages API Conversion Architecture

## Overview

The Claude Messages API conversion system enables users to access various AI models through the standardized Claude Messages API format (`/v1/messages`). This system automatically converts Claude Messages requests to the appropriate format for each adapter (OpenAI, Gemini, Groq, etc.) and converts responses back to Claude Messages format.

## Recent Changes (2025-08)

- **Unified Conversion Handler:** All OpenAI-compatible adapters now use a shared handler (`HandleClaudeMessagesResponse`) for both streaming and non-streaming Claude Messages conversion. This ensures consistent conversion and usage calculation across all adapters.
- **Streaming Conversion Logic:** Streaming responses are converted using `ConvertOpenAIStreamToClaudeSSE`, which emits Claude-native SSE events (`message_start`, `content_block_start`, `content_block_delta`, etc.) and accumulates text, tool arguments, and "thinking" blocks for accurate token usage calculation.
- **Usage Calculation:** Token usage calculation now includes tool call arguments in addition to text, for both streaming and non-streaming responses. If the upstream does not provide usage, or provides incomplete usage, the system computes/fills in the missing values.
- **Controller Fallback Logic:** The controller attempts to extract usage from the Claude response body (JSON), then from SSE (stream), and finally falls back to prompt-only estimation if all else fails.
- **Adapter Implementation Requirement:** All OpenAI-compatible adapters must delegate response handling to the shared handler. This is now the required pattern for new adapters.
- **Billing Impact:** These changes may affect billing and quotas, as more accurate (and sometimes higher) token usage is now reported. All new adapters must follow the unified conversion and usage calculation logic.

## Problem Statement

Different AI providers use different API formats:

- **Anthropic**: Native Claude Messages API format
- **OpenAI-compatible providers**: OpenAI ChatCompletion format
- **Google**: Gemini API format
- **Other providers**: Various proprietary formats

The system needs to:

1. Accept requests in Claude Messages API format
2. Convert to the appropriate format for each adapter
3. Convert responses back to Claude Messages format
4. Maintain full feature compatibility including function calling, streaming, and structured content

## Architecture

### High-Level Flow

```plaintext
User Request (Claude Messages API)
    ↓
[Route to Claude Messages Controller]
    ↓
[Determine Target Adapter]
    ↓
┌─ If Anthropic Adapter ─→ Native Processing
│
└─ If Other Adapter
    ↓
[Convert to Adapter Format]
    ↓
[Send to Upstream via Adapter]
    ↓
[Adapter Response]
    ↓
[Convert back to Claude Messages]
    ↓
User Response (Claude Messages API)
```

### Key Components

#### 1. Claude Messages Controller

**Location**: `relay/controller/claude_messages.go`

**Entry Point**: `RelayClaudeMessagesHelper()` method

**Key Responsibilities**:

- Accept Claude Messages API requests at `/v1/messages`
- Route to appropriate adapter based on model
- Handle response conversion coordination
- Manage streaming and non-streaming responses
- Extract usage information from the response body (JSON or SSE), including tool call arguments, and compute usage if missing.

#### 2. Adapter Conversion Interface

**Interface Methods**:

- `ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error)`
- `DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (*model.Usage, *model.ErrorWithStatusCode)`

**Conversion Patterns**:

1. **Native Support** (Anthropic):

   - Sets `ClaudeMessagesNative` flag
   - Uses native Claude handlers directly

2. **Conversion Support** (OpenAI, Gemini, etc.):
   - Sets `ClaudeMessagesConversion` flag
   - Converts request format in `ConvertClaudeRequest`

- Converts response format in `DoResponse` using the shared handler (`HandleClaudeMessagesResponse`).
- For streaming, uses `ConvertOpenAIStreamToClaudeSSE` to emit Claude-native SSE events and accumulate all relevant content for usage calculation.
- For non-streaming, converts to Claude-native JSON and stores in context for the controller to forward.
- Usage calculation includes both text and tool call arguments.

#### 3. Shared OpenAI-Compatible Conversion

**Location**: `relay/adaptor/openai_compatible/claude_messages.go`

**Function**: `ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error)`

**Used by**: All OpenAI-compatible adapters (DeepSeek, Groq, Mistral, XAI, etc.)

## Adapter Implementation Patterns

### Pattern 1: Native Claude Support (Anthropic)

```go
func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
    // Set native processing flags
    c.Set(ctxkey.ClaudeMessagesNative, true)
    c.Set(ctxkey.ClaudeDirectPassthrough, true)

    return request, nil
}
```

### Pattern 2: OpenAI-Compatible Conversion

```go
func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
    // Use shared OpenAI-compatible conversion
    return openai_compatible.ConvertClaudeRequest(c, request)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
  // Use the shared Claude Messages response handler for both streaming and non-streaming
  return openai_compatible.HandleClaudeMessagesResponse(c, resp, meta, func(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
    if meta.IsStream {
      return openai_compatible.StreamHandler(c, resp, promptTokens, modelName)
    }
    return openai_compatible.Handler(c, resp, promptTokens, modelName)
  })
}
```

### Pattern 3: Custom Conversion (Gemini)

```go
func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
    // Convert to OpenAI format first
    openaiRequest := convertClaudeToOpenAI(request)

    // Set conversion flags
    c.Set(ctxkey.ClaudeMessagesConversion, true)
    c.Set(ctxkey.OriginalClaudeRequest, request)

    // Use Gemini's existing conversion logic
    return a.ConvertRequest(c, relaymode.ChatCompletions, openaiRequest)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (*model.Usage, *model.ErrorWithStatusCode) {
    // Check for Claude Messages conversion
    if isClaudeConversion, exists := c.Get(ctxkey.ClaudeMessagesConversion); exists && isClaudeConversion.(bool) {
        // Convert Gemini response to Claude format
        claudeResp, convertErr := a.convertToClaudeResponse(c, resp, meta)
        if convertErr != nil {
            return nil, convertErr
        }

        // Set converted response for controller to use
        c.Set(ctxkey.ConvertedResponse, claudeResp)
        return nil, nil
    }

    // Normal processing for non-Claude requests
    return a.normalDoResponse(c, resp, meta)
}
```

## Supported Adapters

### ✅ Native Claude Messages Support

- **Anthropic**: Native Claude Messages API support

### ✅ OpenAI-Compatible Conversion

These adapters use the shared `openai_compatible.ConvertClaudeRequest()`:

- **DeepSeek**: `relay/adaptor/deepseek/adaptor.go`
- **Moonshot**: `relay/adaptor/moonshot/adaptor.go`
- **Groq**: `relay/adaptor/groq/adaptor.go`
- **Mistral**: `relay/adaptor/mistral/adaptor.go`
- **XAI**: `relay/adaptor/xai/adaptor.go`
- **TogetherAI**: `relay/adaptor/togetherai/adaptor.go`
- **OpenRouter**: `relay/adaptor/openrouter/adaptor.go`
- **SiliconFlow**: `relay/adaptor/siliconflow/adaptor.go`
- **Doubao**: `relay/adaptor/doubao/adaptor.go`
- **StepFun**: `relay/adaptor/stepfun/adaptor.go`
- **Novita**: `relay/adaptor/novita/adaptor.go`
- **AIProxy**: `relay/adaptor/aiproxy/adaptor.go`
- **LingYiWanWu**: `relay/adaptor/lingyiwanwu/adaptor.go`
- **AI360**: `relay/adaptor/ai360/adaptor.go`

### ✅ Custom Conversion

These adapters implement custom Claude Messages conversion logic:

- **Anthropic**: Native Claude support with direct pass-through
- **OpenAI**: Full conversion with response format transformation
- **Gemini**: Custom conversion with response format transformation
- **Ali**: Custom conversion implementation
- **Baidu**: Custom conversion implementation
- **Zhipu**: Custom conversion implementation
- **Xunfei**: Custom conversion implementation
- **Tencent**: Custom conversion implementation
- **AWS**: Custom conversion for Bedrock Claude models
- **VertexAI**: Custom conversion with sub-adapter routing
- **Replicate**: Custom conversion implementation
- **Cohere**: Custom conversion implementation
- **Cloudflare**: Uses shared OpenAI-compatible response handlers for streaming and non-streaming; retains custom request URL/model mapping
- **Palm**: Basic text-only conversion support
- **Ollama**: Basic text-only conversion support
- **Coze**: Basic text-only conversion support

### ❌ Limited or No Support

These adapters have limited or no Claude Messages support:

- **DeepL**: Translation service, not applicable for chat completion
- **Minimax**: Stub implementation, returns "not implemented" error
- **Baichuan**: Stub implementation, returns "not implemented" error

### 📋 Test Results Summary

Based on comprehensive testing:

- **✅ Fully Working**: 25+ adapters with complete Claude Messages support
- **⚠️ Configuration Required**: Some adapters (Baidu, Tencent, VertexAI) require valid API keys/configuration
- **❌ Not Applicable**: 3 adapters (DeepL, Minimax, Baichuan) correctly return appropriate errors

## Data Structure Mappings

### Claude Messages to OpenAI Format

| Claude Messages Field | OpenAI Field  | Notes                              |
| --------------------- | ------------- | ---------------------------------- |
| `model`               | `model`       | Direct mapping                     |
| `max_tokens`          | `max_tokens`  | Direct mapping                     |
| `messages`            | `messages`    | Message format conversion required |
| `system`              | `messages[0]` | System message as first message    |
| `tools`               | `tools`       | Tool format conversion required    |
| `tool_choice`         | `tool_choice` | Direct mapping                     |
| `temperature`         | `temperature` | Direct mapping                     |
| `top_p`               | `top_p`       | Direct mapping                     |
| `stream`              | `stream`      | Direct mapping                     |
| `stop_sequences`      | `stop`        | Direct mapping                     |

### Message Content Conversion

#### Text Content

```json
// Claude Messages
{"role": "user", "content": "Hello"}

// OpenAI
{"role": "user", "content": "Hello"}
```

#### Structured Content

```json
// Claude Messages
{
  "role": "user",
  "content": [
    {"type": "text", "text": "Hello"},
    {"type": "image", "source": {"type": "base64", "media_type": "image/jpeg", "data": "..."}}
  ]
}

// OpenAI
{
  "role": "user",
  "content": [
    {"type": "text", "text": "Hello"},
    {"type": "image_url", "image_url": {"url": "data:image/jpeg;base64,..."}}
  ]
}
```

#### Tool Use

```json
// Claude Messages
{
  "role": "assistant",
  "content": [
    {"type": "tool_use", "id": "toolu_123", "name": "get_weather", "input": {"location": "NYC"}}
  ]
}

// OpenAI
{
  "role": "assistant",
  "tool_calls": [
    {"id": "toolu_123", "type": "function", "function": {"name": "get_weather", "arguments": "{\"location\":\"NYC\"}"}}
  ]
}
```

### Response Format Conversion

#### OpenAI to Claude Messages

| OpenAI Field                    | Claude Messages Field | Notes                                                    |
| ------------------------------- | --------------------- | -------------------------------------------------------- |
| `id`                            | `id`                  | Generate Claude-style ID if missing                      |
| `choices[0].message.content`    | `content[0].text`     | Text content                                             |
| `choices[0].message.tool_calls` | `content[].tool_use`  | Tool calls conversion                                    |
| `choices[0].finish_reason`      | `stop_reason`         | Reason mapping required                                  |
| `usage.prompt_tokens`           | `usage.input_tokens`  | Direct mapping                                           |
| `usage.completion_tokens`       | `usage.output_tokens` | Direct mapping                                           |
| tool call arguments             | included in usage     | Token usage calculation now includes tool call arguments |

#### Finish Reason Mapping

| OpenAI finish_reason | Claude stop_reason | Notes               |
| -------------------- | ------------------ | ------------------- |
| `stop`               | `end_turn`         | Normal completion   |
| `length`             | `max_tokens`       | Token limit reached |
| `tool_calls`         | `tool_use`         | Function calling    |
| `content_filter`     | `stop_sequence`    | Content filtered    |

## Context Management

### Context Keys

**Location**: `common/ctxkey/key.go`

**Key Constants**:

- `ClaudeMessagesConversion = "claude_messages_conversion"`
- `ClaudeMessagesNative = "claude_messages_native"`
- `ClaudeDirectPassthrough = "claude_direct_passthrough"`
- `OriginalClaudeRequest = "original_claude_request"`
- `ConvertedResponse = "converted_response"`

### Context Flow

1. **Request Phase**:

   - `ConvertClaudeRequest()` sets conversion flags
   - Original request stored for reference

2. **Response Phase**:

   - `DoResponse()` checks conversion flags
   - Converts response format if needed
   - Sets converted response in context

3. **Controller Phase**:
   - Controller checks for converted response
   - Uses converted response or falls back to native handlers

## Error Handling

### Conversion Errors

- Request conversion errors wrapped with proper error types
- Response parsing errors logged with debug information
- Malformed content handled gracefully with fallbacks
- Usage calculation fallbacks: If usage is missing, the system computes it from all available content (including tool call arguments). If still missing, only prompt tokens are counted.

### Adapter Errors

- Upstream adapter errors passed through unchanged
- Error format preserved for client compatibility
- Proper HTTP status codes maintained

### Fallback Mechanisms

- Token usage calculation fallback when adapter doesn't provide usage
- Content extraction fallback for malformed responses
- Default values for missing required fields

## Performance Considerations

### Memory Management

- Minimal context storage for conversion flags
- Efficient message content transformation
- Streaming support with proper buffer management

### Processing Efficiency

- Single-pass conversion for request and response
- Lazy evaluation - conversion only when needed
- Early detection of conversion requirements

## Testing

### Test Coverage

**Locations**:

- `relay/adaptor/gemini/adaptor_test.go` - Gemini conversion tests
- `relay/adaptor/openai_compatible/claude_messages_test.go` - Shared conversion tests
- Individual adapter test files for specific conversion logic

**Key Test Categories**:

- Request format conversion
- Response format conversion
- Streaming conversion
- Function calling workflows
- Error handling scenarios

## Future Enhancements

### 1. Enhanced Content Support

- Support for more Claude Messages content types
- Better handling of complex structured content
- Improved image and file handling

### 2. Performance Optimizations

- Response format detection optimization
- Memory usage optimization for large messages
- Caching for repeated conversions

### 3. Extended Adapter Support

- Support for more specialized adapters
- Dynamic adapter capability detection
- Runtime adapter registration

## Summary

The Claude Messages API conversion system provides:

1. **Universal Access**: Single API endpoint for multiple AI providers
2. **Format Transparency**: Automatic format conversion between different APIs
3. **Feature Preservation**: Full support for function calling, streaming, and structured content
4. **Extensible Architecture**: Easy addition of new adapters
5. **Error Resilience**: Comprehensive error handling and fallbacks

The implementation allows users to interact with various AI models through the familiar Claude Messages API while maintaining compatibility with each provider's native capabilities.
