package openai_compatible

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/render"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/model"
)

const (
	DataPrefix       = "data: "
	DataPrefixLength = len(DataPrefix)
	Done             = "[DONE]"
)

// ChatCompletionsStreamResponse represents the streaming response structure
type ChatCompletionsStreamResponse struct {
	Id      string                                `json:"id"`
	Object  string                                `json:"object"`
	Created int64                                 `json:"created"`
	Model   string                                `json:"model"`
	Choices []ChatCompletionsStreamResponseChoice `json:"choices"`
	Usage   *model.Usage                          `json:"usage,omitempty"`
}

type ChatCompletionsStreamResponseChoice struct {
	Index        int           `json:"index"`
	Delta        model.Message `json:"delta"`
	FinishReason *string       `json:"finish_reason"`
}

// SlimTextResponse represents the non-streaming response structure
type SlimTextResponse struct {
	Choices     []TextResponseChoice `json:"choices"`
	model.Usage `json:"usage"`
	Error       model.Error `json:"error"`
}

type TextResponseChoice struct {
	Index        int           `json:"index"`
	Message      model.Message `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// NormalizeDataLine normalizes SSE data lines
func NormalizeDataLine(data string) string {
	if strings.HasPrefix(data, "data:") {
		content := strings.TrimLeft(data[len("data:"):], " ")
		return "data: " + content
	}
	return data
}

// ErrorWrapper creates an error response
func ErrorWrapper(err error, code string, statusCode int) *model.ErrorWithStatusCode {
	return &model.ErrorWithStatusCode{
		Error: model.Error{
			Message: err.Error(),
			Type:    "one_api_error",
			Code:    code,
		},
		StatusCode: statusCode,
	}
}

// CountTokenText estimates token count (simplified implementation)
func CountTokenText(text string, modelName string) int {
	// Simple estimation: ~4 characters per token
	return len(text) / 4
}

// GetFullRequestURL constructs the full request URL for OpenAI-compatible APIs
func GetFullRequestURL(baseURL string, requestURL string, channelType int) string {
	if channelType == channeltype.OpenAICompatible {
		return fmt.Sprintf("%s%s", strings.TrimSuffix(baseURL, "/"), strings.TrimPrefix(requestURL, "/v1"))
	}
	return fmt.Sprintf("%s%s", baseURL, requestURL)
}

// StreamHandler processes streaming responses from OpenAI-compatible APIs
func StreamHandler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	responseText := ""
	toolArgsText := ""
	var usage *model.Usage

	logger := gmw.GetLogger(c).With(
		zap.String("model", modelName),
	)

	// Check if response content type indicates an error (non-streaming response)
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") &&
		!strings.Contains(contentType, "text/event-stream") {
		logger.Error("unexpected content type for streaming request, possible error response",
			zap.String("content_type", contentType),
			zap.Int("status_code", resp.StatusCode))

		// Read response as potential error
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return ErrorWrapper(err, "read_error_response_failed", http.StatusInternalServerError), nil
		}

		logger.Error("received error response in stream handler",
			zap.ByteString("response_body", responseBody))

		// Try to parse as error response
		var errorResponse SlimTextResponse
		if err := json.Unmarshal(responseBody, &errorResponse); err == nil && errorResponse.Error.Type != "" {
			return &model.ErrorWithStatusCode{
				Error:      errorResponse.Error,
				StatusCode: resp.StatusCode,
			}, nil
		}

		// Return generic error if parsing fails
		return ErrorWrapper(errors.Errorf("unexpected non-streaming response: %s", string(responseBody)),
			"unexpected_response_format", resp.StatusCode), nil
	}

	scanner := bufio.NewScanner(resp.Body)
	buffer := make([]byte, 1024*1024) // 1MB buffer
	scanner.Buffer(buffer, len(buffer))
	scanner.Split(bufio.ScanLines)

	common.SetEventStreamHeaders(c)
	doneRendered := false
	chunksProcessed := 0

	for scanner.Scan() {
		data := NormalizeDataLine(scanner.Text())
		logger.Debug("processing streaming chunk",
			zap.String("chunk_data", data),
			zap.Int("chunks_processed", chunksProcessed))

		if len(data) < DataPrefixLength {
			continue
		}

		if data[:DataPrefixLength] != DataPrefix && data[:DataPrefixLength] != Done {
			continue
		}

		if strings.HasPrefix(data[DataPrefixLength:], Done) {
			render.StringData(c, data)
			doneRendered = true
			continue
		}

		// Parse the streaming chunk
		var streamResponse ChatCompletionsStreamResponse
		jsonData := data[DataPrefixLength:]
		if err := json.Unmarshal([]byte(jsonData), &streamResponse); err != nil {
			logger.Warn("failed to parse streaming chunk, skipping",
				zap.String("chunk_data", jsonData),
				zap.Error(err))
			continue // Skip malformed chunks
		}

		chunksProcessed++

		// Accumulate response text and tool call arguments
		for _, choice := range streamResponse.Choices {
			responseText += choice.Delta.StringContent()
			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					if tc.Function != nil && tc.Function.Arguments != nil {
						switch v := tc.Function.Arguments.(type) {
						case string:
							toolArgsText += v
						default:
							if b, e := json.Marshal(v); e == nil {
								toolArgsText += string(b)
							}
						}
					}
				}
			}
		}

		// Accumulate usage information
		if streamResponse.Usage != nil {
			usage = streamResponse.Usage
		}

		// Forward the chunk to client
		render.StringData(c, data)
	}

	if err := scanner.Err(); err != nil {
		return ErrorWrapper(err, "read_stream_failed", http.StatusInternalServerError), usage
	}

	// Check if we processed any chunks - if not, this might indicate an error
	if chunksProcessed == 0 && responseText == "" {
		logger.Error("stream processing completed but no chunks were processed",
			zap.String("model", modelName),
			zap.String("content_type", resp.Header.Get("Content-Type")))
		return ErrorWrapper(errors.Errorf("no streaming data received from upstream"),
			"empty_stream_response", http.StatusInternalServerError), usage
	}

	if !doneRendered {
		render.StringData(c, "data: "+Done)
	}

	if err := resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), usage
	}

	// Calculate or fix usage when missing or incomplete
	if usage == nil {
		// No usage provided by upstream: compute from text
		computed := CountTokenText(responseText, modelName) + CountTokenText(toolArgsText, modelName)
		usage = &model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: computed,
			TotalTokens:      promptTokens + computed,
		}
		logger.Debug("computed usage for stream (no upstream usage)",
			zap.Int("prompt_tokens", usage.PromptTokens),
			zap.Int("completion_tokens", usage.CompletionTokens),
			zap.Int("total_tokens", usage.TotalTokens),
			zap.Int("response_text_len", len(responseText)),
			zap.Int("tool_args_len", len(toolArgsText)))
	} else {
		// Upstream provided some usage; fill missing parts
		if usage.PromptTokens == 0 {
			usage.PromptTokens = promptTokens
		}
		if usage.CompletionTokens == 0 {
			usage.CompletionTokens = CountTokenText(responseText, modelName) + CountTokenText(toolArgsText, modelName)
		}
		if usage.TotalTokens == 0 {
			usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
		}
		logger.Debug("finalized usage for stream (with upstream usage)",
			zap.Int("prompt_tokens", usage.PromptTokens),
			zap.Int("completion_tokens", usage.CompletionTokens),
			zap.Int("total_tokens", usage.TotalTokens),
			zap.Int("response_text_len", len(responseText)),
			zap.Int("tool_args_len", len(toolArgsText)))
	}

	return nil, usage
}

// Handler processes non-streaming responses from OpenAI-compatible APIs
func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	logger := gmw.GetLogger(c)
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}

	logger.Debug("receive from upstream channel", zap.ByteString("response_body", responseBody))
	if err = resp.Body.Close(); err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	// Check if response body is empty
	if len(responseBody) == 0 {
		logger.Error("received empty response body from upstream",
			zap.String("model", modelName),
			zap.Int("status_code", resp.StatusCode))
		return ErrorWrapper(errors.Errorf("empty response body from upstream"),
			"empty_response_body", http.StatusInternalServerError), nil
	}

	var textResponse SlimTextResponse
	if err = json.Unmarshal(responseBody, &textResponse); err != nil {
		logger.Error("failed to unmarshal response body",
			zap.ByteString("response_body", responseBody),
			zap.Error(err))
		return ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	if textResponse.Error.Type != "" {
		logger.Debug("upstream returned error response",
			zap.String("error_type", textResponse.Error.Type),
			zap.String("error_message", textResponse.Error.Message))
		return &model.ErrorWithStatusCode{
			Error:      textResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}

	// Check if response has choices - empty choices might indicate an error
	if len(textResponse.Choices) == 0 {
		logger.Error("response has no choices, possible upstream error",
			zap.String("model", modelName),
			zap.ByteString("response_body", responseBody))
		return ErrorWrapper(errors.Errorf("no choices in response from upstream"),
			"no_choices_in_response", http.StatusInternalServerError), nil
	}

	// Forward the response to client
	c.JSON(resp.StatusCode, textResponse)

	// Calculate usage if not provided
	usage := textResponse.Usage
	if usage.PromptTokens == 0 {
		usage.PromptTokens = promptTokens
	}
	if usage.CompletionTokens == 0 {
		// Calculate completion tokens from response text and tool call arguments
		responseText := ""
		toolArgsText := ""
		for _, choice := range textResponse.Choices {
			responseText += choice.Message.StringContent()
			if len(choice.Message.ToolCalls) > 0 {
				for _, tc := range choice.Message.ToolCalls {
					if tc.Function != nil && tc.Function.Arguments != nil {
						switch v := tc.Function.Arguments.(type) {
						case string:
							toolArgsText += v
						default:
							if b, e := json.Marshal(v); e == nil {
								toolArgsText += string(b)
							}
						}
					}
				}
			}
		}
		usage.CompletionTokens = CountTokenText(responseText, modelName) + CountTokenText(toolArgsText, modelName)
	}
	if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	logger.Debug("finalized usage for non-stream (openai-compatible)",
		zap.Int("prompt_tokens", usage.PromptTokens),
		zap.Int("completion_tokens", usage.CompletionTokens),
		zap.Int("total_tokens", usage.TotalTokens))

	return nil, &usage
}
