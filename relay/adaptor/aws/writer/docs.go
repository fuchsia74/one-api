/*
Package aws provides AWS Bedrock integration for Writer language models.

This package implements support for Writer's Palmyra models (X4 and X5) through
AWS Bedrock, enabling OpenAI-compatible API access to these models.

# Supported Models

Writer models available through AWS Bedrock:
  - writer.palmyra-x4-v1:0
  - writer.palmyra-x5-v1:0

# Supported Parameters

Writer models support the following parameters:
  - max_tokens: Maximum number of tokens to generate
  - temperature: Randomness in generation (0.0 to 1.0)
  - top_p: Nucleus sampling parameter (0.0 to 1.0)
  - stop: Stop sequences to terminate generation

# Features

Writer models provide:
  - Simple text/chat completion capabilities
  - Both streaming and non-streaming responses
  - Token usage tracking for billing
  - OpenAI-compatible request/response format
  - No reasoning content (unlike DeepSeek/OpenAI models)

# Usage

Writer models are accessed through the standard One API endpoints with
AWS Bedrock configuration. The adapter handles conversion between OpenAI
format and Writer's native format automatically.

# Implementation Details

The Writer adapter follows the standard AWS Bedrock adapter pattern:
  - Uses AWS Converse API for optimal token counting
  - Supports both streaming and non-streaming modes
  - Provides proper error handling and usage tracking
  - Implements simple text content processing (no reasoning)
*/
package aws
