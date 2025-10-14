package openai_compatible

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// ConvertClaudeRequest converts Claude Messages API request to OpenAI format for OpenAI-compatible adapters
func ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// Convert Claude Messages API request to OpenAI format first
	openaiRequest := &model.GeneralOpenAIRequest{
		Model:               request.Model,
		MaxCompletionTokens: &request.MaxTokens,
		Temperature:         request.Temperature,
		TopP:                request.TopP,
		Stream:              request.Stream != nil && *request.Stream,
		Stop:                request.StopSequences,
	}

	// Convert system message if present
	if request.System != nil {
		switch system := request.System.(type) {
		case string:
			if system != "" {
				openaiRequest.Messages = append(openaiRequest.Messages, model.Message{
					Role:    "system",
					Content: system,
				})
			}
		case []any:
			// Extract text parts and join; ignore non-text
			var parts []string
			for _, block := range system {
				if blockMap, ok := block.(map[string]any); ok {
					if t, ok := blockMap["type"].(string); ok && t == "text" {
						if text, exists := blockMap["text"]; exists {
							if textStr, ok := text.(string); ok && textStr != "" {
								parts = append(parts, textStr)
							}
						}
					}
				}
			}
			if len(parts) > 0 {
				openaiRequest.Messages = append(openaiRequest.Messages, model.Message{
					Role:    "system",
					Content: strings.Join(parts, "\n"),
				})
			}
		}
	}

	// Convert messages
	for _, msg := range request.Messages {
		openaiMessage := model.Message{Role: msg.Role}

		switch content := msg.Content.(type) {
		case string:
			openaiMessage.Content = content
		case []any:
			var contentParts []model.MessageContent
			for _, block := range content {
				blockMap, ok := block.(map[string]any)
				if !ok {
					continue
				}
				bt, _ := blockMap["type"].(string)
				switch bt {
				case "text":
					if text, exists := blockMap["text"]; exists {
						if textStr, ok := text.(string); ok {
							contentParts = append(contentParts, model.MessageContent{Type: "text", Text: &textStr})
						}
					}
				case "image":
					if source, exists := blockMap["source"]; exists {
						if sourceMap, ok := source.(map[string]any); ok {
							if st, _ := sourceMap["type"].(string); st == "base64" {
								if mt, ok := sourceMap["media_type"].(string); ok {
									if data, ok := sourceMap["data"].(string); ok {
										contentParts = append(contentParts, model.MessageContent{
											Type:     "image_url",
											ImageURL: &model.ImageURL{Url: fmt.Sprintf("data:%s;base64,%s", mt, data)},
										})
									}
								}
							} else if st == "url" {
								if urlStr, ok := sourceMap["url"].(string); ok {
									contentParts = append(contentParts, model.MessageContent{
										Type:     "image_url",
										ImageURL: &model.ImageURL{Url: urlStr},
									})
								}
							}
						}
					}
				case "tool_use":
					if id, ok := blockMap["id"].(string); ok {
						if name, ok := blockMap["name"].(string); ok {
							input := blockMap["input"]
							var argsStr string
							if inputBytes, err := json.Marshal(input); err == nil {
								argsStr = string(inputBytes)
							}
							openaiMessage.ToolCalls = append(openaiMessage.ToolCalls, model.Tool{
								Id:   id,
								Type: "function",
								Function: &model.Function{
									Name:      name,
									Arguments: argsStr,
								},
							})
						}
					}
				case "tool_result":
					if toolCallId, ok := blockMap["tool_call_id"].(string); ok {
						var contentStr string
						switch cc := blockMap["content"].(type) {
						case string:
							contentStr = cc
						case []any:
							for _, item := range cc {
								if itemMap, ok := item.(map[string]any); ok {
									if t, _ := itemMap["type"].(string); t == "text" {
										if txt, ok := itemMap["text"].(string); ok {
											contentStr += txt
										}
									}
								}
							}
						}
						openaiMessage.ToolCallId = toolCallId
						openaiMessage.Content = contentStr
					}
				default:
					// ignore unknown block types gracefully
				}
			}
			if len(contentParts) > 0 {
				openaiMessage.Content = contentParts
			} else if openaiMessage.Content == nil {
				// Ensure content is present for providers requiring it
				openaiMessage.Content = ""
			}
		default:
			if b, err := json.Marshal(content); err == nil {
				openaiMessage.Content = string(b)
			} else {
				openaiMessage.Content = ""
			}
		}

		openaiRequest.Messages = append(openaiRequest.Messages, openaiMessage)
	}

	// Convert tools if present
	if len(request.Tools) > 0 {
		var tools []model.Tool
		for _, claudeTool := range request.Tools {
			tool := model.Tool{
				Type: "function",
				Function: &model.Function{
					Name:        claudeTool.Name,
					Description: claudeTool.Description,
					Parameters:  claudeTool.InputSchema.(map[string]any),
				},
			}
			tools = append(tools, tool)
		}
		openaiRequest.Tools = tools
	}

	// Convert tool choice if present
	if request.ToolChoice != nil {
		openaiRequest.ToolChoice = normalizeClaudeToolChoice(request.ToolChoice)
	}

	// Mark this as a Claude Messages conversion for response handling
	c.Set(ctxkey.ClaudeMessagesConversion, true)
	c.Set(ctxkey.OriginalClaudeRequest, request)

	return openaiRequest, nil
}

// normalizeClaudeToolChoice adapts Claude tool_choice payloads to the OpenAI ChatCompletion schema.
// Claude clients often set type=tool with a top-level name; OpenAI-compatible upstreams expect
// type=function with the name nested under the function field.
func normalizeClaudeToolChoice(choice any) any {
	switch src := choice.(type) {
	case map[string]any:
		cloned := make(map[string]any, len(src))
		for k, v := range src {
			cloned[k] = v
		}

		typeVal, _ := cloned["type"].(string)
		switch strings.ToLower(typeVal) {
		case "tool":
			name, _ := cloned["name"].(string)
			var fn map[string]any
			if existing, ok := cloned["function"].(map[string]any); ok {
				fn = cloneMap(existing)
			} else {
				fn = map[string]any{}
			}
			if name != "" && fn["name"] == nil {
				fn["name"] = name
			}
			if len(fn) > 0 {
				cloned["function"] = fn
			}
			cloned["type"] = "function"
			delete(cloned, "name")
		case "function":
			if name, ok := cloned["name"].(string); ok && name != "" {
				fn, _ := cloned["function"].(map[string]any)
				if fn == nil {
					fn = map[string]any{}
				}
				if fn["name"] == nil {
					fn["name"] = name
				}
				cloned["function"] = fn
				delete(cloned, "name")
			}
		}
		return cloned
	default:
		return choice
	}
}

// cloneMap returns a shallow copy of a map[string]any. It avoids mutating caller state when normalizing payloads.
func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	cloned := make(map[string]any, len(input))
	for k, v := range input {
		cloned[k] = v
	}
	return cloned
}

// HandleClaudeMessagesResponse handles Claude Messages response conversion for OpenAI-compatible adapters
// This should be called in the adapter's DoResponse method when ClaudeMessagesConversion flag is set
func HandleClaudeMessagesResponse(c *gin.Context, resp *http.Response, meta *meta.Meta, handler func(*gin.Context, *http.Response, int, string) (*model.ErrorWithStatusCode, *model.Usage)) (*model.Usage, *model.ErrorWithStatusCode) {
	// Check if this is a Claude Messages conversion
	if isClaudeConversion, exists := c.Get(ctxkey.ClaudeMessagesConversion); !exists || !isClaudeConversion.(bool) {
		// Not a Claude Messages conversion, proceed normally
		err, usage := handler(c, resp, meta.PromptTokens, meta.ActualModelName)
		return usage, err
	}

	// Claude Messages conversion path
	if meta.IsStream {
		// Convert OpenAI-compatible SSE to Claude-native SSE, write to client, return usage
		usage, convErr := ConvertOpenAIStreamToClaudeSSE(c, resp, meta.PromptTokens, meta.ActualModelName)
		if convErr != nil {
			return nil, convErr
		}
		return usage, nil
	}

	// Non-stream: convert to Claude JSON and let controller forward it
	claudeResp, convErr := ConvertOpenAIResponseToClaudeResponse(c, resp)
	if convErr != nil {
		return nil, convErr
	}
	c.Set(ctxkey.ConvertedResponse, claudeResp)
	return nil, nil
}
