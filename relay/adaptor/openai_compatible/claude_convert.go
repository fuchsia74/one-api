package openai_compatible

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// ConvertOpenAIResponseToClaudeResponse converts an OpenAI-compatible response
// (Chat Completions or Response API) into Claude Messages JSON http.Response.
func ConvertOpenAIResponseToClaudeResponse(_ *gin.Context, resp *http.Response) (*http.Response, *relaymodel.ErrorWithStatusCode) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError)
	}
	_ = resp.Body.Close()

	// 1) Try Response API format first
	var responseAPIResp responseAPIResponse
	if err := json.Unmarshal(body, &responseAPIResp); err == nil && responseAPIResp.Object == "response" {
		claudeResp := responseAPIResponseToClaude(&responseAPIResp)
		return marshalClaudeHTTPResponse(resp, claudeResp)
	}

	// 2) Fallback: Chat Completions format
	var chatResp chatTextResponse
	if err := json.Unmarshal(body, &chatResp); err == nil && len(chatResp.Choices) > 0 {
		claudeResp := chatResponseToClaude(&chatResp)
		return marshalClaudeHTTPResponse(resp, claudeResp)
	}

	// 3) Unknown format – return original payload (controller may handle error)
	newResp := &http.Response{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
	return newResp, nil
}

// responseAPIResponseToClaude maps OpenAI Response API response to ClaudeMessages response
func responseAPIResponseToClaude(r *responseAPIResponse) relaymodel.ClaudeResponse {
	out := relaymodel.ClaudeResponse{
		ID:         r.Id,
		Type:       "message",
		Role:       "assistant",
		Model:      r.Model,
		Content:    []relaymodel.ClaudeContent{},
		StopReason: "end_turn",
	}

	if r.Usage != nil {
		out.Usage = relaymodel.ClaudeUsage{
			InputTokens:  r.Usage.InputTokens,
			OutputTokens: r.Usage.OutputTokens,
		}
	}

	for _, item := range r.Output {
		switch item.Type {
		case "message":
			if item.Role == "assistant" {
				for _, c := range item.Content {
					if c.Type == "output_text" && c.Text != "" {
						out.Content = append(out.Content, relaymodel.ClaudeContent{Type: "text", Text: c.Text})
					}
				}
			}
		case "reasoning":
			for _, s := range item.Summary {
				if s.Type == "summary_text" && s.Text != "" {
					out.Content = append(out.Content, relaymodel.ClaudeContent{Type: "thinking", Thinking: s.Text})
				}
			}
		case "function_call":
			// Map to Claude tool_use block
			input := json.RawMessage(item.Arguments)
			out.Content = append(out.Content, relaymodel.ClaudeContent{
				Type:  "tool_use",
				ID:    item.CallId,
				Name:  item.Name,
				Input: input,
			})
		}
	}

	return out
}

// chatResponseToClaude maps OpenAI Chat Completion response to ClaudeMessages response
func chatResponseToClaude(r *chatTextResponse) relaymodel.ClaudeResponse {
	out := relaymodel.ClaudeResponse{
		ID:         r.Id,
		Type:       "message",
		Role:       "assistant",
		Model:      r.Model,
		Content:    []relaymodel.ClaudeContent{},
		StopReason: "end_turn",
		Usage: relaymodel.ClaudeUsage{
			InputTokens:  r.Usage.PromptTokens,
			OutputTokens: r.Usage.CompletionTokens,
		},
	}

	for _, choice := range r.Choices {
		// Text content
		if choice.Message.Content != nil {
			switch content := choice.Message.Content.(type) {
			case string:
				if content != "" {
					out.Content = append(out.Content, relaymodel.ClaudeContent{Type: "text", Text: content})
				}
			case []relaymodel.MessageContent:
				for _, part := range content {
					if part.Type == "text" && part.Text != nil && *part.Text != "" {
						out.Content = append(out.Content, relaymodel.ClaudeContent{Type: "text", Text: *part.Text})
					}
				}
			}
		}

		// Tool calls -> tool_use blocks
		if len(choice.Message.ToolCalls) > 0 {
			for _, tc := range choice.Message.ToolCalls {
				var input json.RawMessage
				if tc.Function.Arguments != nil {
					switch v := tc.Function.Arguments.(type) {
					case string:
						input = json.RawMessage(v)
					default:
						if b, err := json.Marshal(v); err == nil {
							input = json.RawMessage(b)
						}
					}
				}
				out.Content = append(out.Content, relaymodel.ClaudeContent{
					Type:  "tool_use",
					ID:    tc.Id,
					Name:  tc.Function.Name,
					Input: input,
				})
			}
		}

		// Map finish reason
		switch choice.FinishReason {
		case "stop":
			out.StopReason = "end_turn"
		case "length":
			out.StopReason = "max_tokens"
		case "tool_calls":
			out.StopReason = "tool_use"
		case "content_filter":
			out.StopReason = "stop_sequence"
		}
	}

	return out
}

func marshalClaudeHTTPResponse(orig *http.Response, payload relaymodel.ClaudeResponse) (*http.Response, *relaymodel.ErrorWithStatusCode) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, ErrorWrapper(fmt.Errorf("marshal_claude_response: %w", err), "marshal_claude_response_failed", http.StatusInternalServerError)
	}
	newResp := &http.Response{
		StatusCode: orig.StatusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(b)),
	}
	// Copy headers and set content type/length
	for k, v := range orig.Header {
		newResp.Header[k] = v
	}
	newResp.Header.Set("Content-Type", "application/json")
	newResp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
	return newResp, nil
}

// ConvertOpenAIStreamToClaudeSSE reads an OpenAI-compatible chat completion/response-api SSE stream
// and writes Claude-native SSE events to the client, returning computed usage.
func ConvertOpenAIStreamToClaudeSSE(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*relaymodel.Usage, *relaymodel.ErrorWithStatusCode) {
	_ = gmw.GetLogger(c)

	// Prepare client for SSE
	common.SetEventStreamHeaders(c)

	scanner := bufio.NewScanner(resp.Body)
	buffer := make([]byte, 1024*1024)
	scanner.Buffer(buffer, len(buffer))
	scanner.Split(bufio.ScanLines)

	accumText := ""
	accumThinking := ""
	accumToolArgs := ""
	var usage *relaymodel.Usage

	// Track content blocks and indices
	nextIndex := 0
	thinkingIndex := -1
	textIndex := -1
	toolStarted := map[string]int{} // tool_call_id -> index

	// Emit message_start
	msgStart := map[string]any{
		"type": "message_start",
		"message": map[string]any{
			"type":    "message",
			"role":    "assistant",
			"model":   modelName,
			"content": []any{},
		},
	}
	if b, err := json.Marshal(msgStart); err == nil {
		c.Writer.Write([]byte("data: "))
		c.Writer.Write(b)
		c.Writer.Write([]byte("\n\n"))
		c.Writer.(http.Flusher).Flush()
	}

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}

		// Parse OpenAI-compatible streaming chunk
		var chunk ChatCompletionsStreamResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}

		// Process choices
		for _, choice := range chunk.Choices {
			// Thinking delta
			if choice.Delta.Thinking != nil && *choice.Delta.Thinking != "" {
				if thinkingIndex == -1 {
					// Start thinking block at next index
					start := map[string]any{
						"type":          "content_block_start",
						"index":         nextIndex,
						"content_block": map[string]any{"type": "thinking", "thinking": ""},
					}
					if b, e := json.Marshal(start); e == nil {
						c.Writer.Write([]byte("data: "))
						c.Writer.Write(b)
						c.Writer.Write([]byte("\n\n"))
						c.Writer.(http.Flusher).Flush()
					}
					thinkingIndex = nextIndex
					nextIndex++
				}
				thinkingDelta := *choice.Delta.Thinking
				accumThinking += thinkingDelta
				delta := map[string]any{
					"type":  "content_block_delta",
					"index": thinkingIndex,
					"delta": map[string]any{"type": "thinking_delta", "thinking": thinkingDelta},
				}
				if b, e := json.Marshal(delta); e == nil {
					c.Writer.Write([]byte("data: "))
					c.Writer.Write(b)
					c.Writer.Write([]byte("\n\n"))
					c.Writer.(http.Flusher).Flush()
				}
			}

			// Signature delta (attached to thinking block)
			if choice.Delta.Signature != nil && *choice.Delta.Signature != "" {
				if thinkingIndex == -1 {
					// Start thinking block to attach signature
					start := map[string]any{
						"type":          "content_block_start",
						"index":         nextIndex,
						"content_block": map[string]any{"type": "thinking", "thinking": ""},
					}
					if b, e := json.Marshal(start); e == nil {
						c.Writer.Write([]byte("data: "))
						c.Writer.Write(b)
						c.Writer.Write([]byte("\n\n"))
						c.Writer.(http.Flusher).Flush()
					}
					thinkingIndex = nextIndex
					nextIndex++
				}
				sig := *choice.Delta.Signature
				delta := map[string]any{
					"type":  "content_block_delta",
					"index": thinkingIndex,
					"delta": map[string]any{"type": "signature_delta", "signature": sig},
				}
				if b, e := json.Marshal(delta); e == nil {
					c.Writer.Write([]byte("data: "))
					c.Writer.Write(b)
					c.Writer.Write([]byte("\n\n"))
					c.Writer.(http.Flusher).Flush()
				}
			}

			// Text delta
			deltaText := choice.Delta.StringContent()
			if deltaText != "" {
				if textIndex == -1 {
					// Start text content block at next index
					start := map[string]any{
						"type":          "content_block_start",
						"index":         nextIndex,
						"content_block": map[string]any{"type": "text", "text": ""},
					}
					if b, e := json.Marshal(start); e == nil {
						c.Writer.Write([]byte("data: "))
						c.Writer.Write(b)
						c.Writer.Write([]byte("\n\n"))
						c.Writer.(http.Flusher).Flush()
					}
					textIndex = nextIndex
					nextIndex++
				}
				accumText += deltaText
				delta := map[string]any{
					"type":  "content_block_delta",
					"index": textIndex,
					"delta": map[string]any{"type": "text_delta", "text": deltaText},
				}
				if b, e := json.Marshal(delta); e == nil {
					c.Writer.Write([]byte("data: "))
					c.Writer.Write(b)
					c.Writer.Write([]byte("\n\n"))
					c.Writer.(http.Flusher).Flush()
				}
			}

			// Tool call deltas
			if len(choice.Delta.ToolCalls) > 0 {
				for _, tc := range choice.Delta.ToolCalls {
					id := tc.Id
					if id == "" {
						id = fmt.Sprintf("tool_%d", nextIndex)
					}
					idx, exists := toolStarted[id]
					if !exists {
						// Start a new tool_use block
						idx = nextIndex
						toolStarted[id] = idx
						nextIndex++
						start := map[string]any{
							"type":  "content_block_start",
							"index": idx,
							"content_block": map[string]any{
								"type": "tool_use",
								"id":   id,
								"name": func() string {
									if tc.Function != nil {
										return tc.Function.Name
									}
									return ""
								}(),
								"input": map[string]any{},
							},
						}
						if b, e := json.Marshal(start); e == nil {
							c.Writer.Write([]byte("data: "))
							c.Writer.Write(b)
							c.Writer.Write([]byte("\n\n"))
							c.Writer.(http.Flusher).Flush()
						}
					}

					// Delta arguments
					var argStr string
					if tc.Function != nil && tc.Function.Arguments != nil {
						switch v := tc.Function.Arguments.(type) {
						case string:
							argStr = v
						default:
							if b, e := json.Marshal(v); e == nil {
								argStr = string(b)
							}
						}
					}
					if argStr != "" {
						accumToolArgs += argStr
						delta := map[string]any{
							"type":  "content_block_delta",
							"index": idx,
							"delta": map[string]any{"type": "input_json_delta", "partial_json": argStr},
						}
						if b, e := json.Marshal(delta); e == nil {
							c.Writer.Write([]byte("data: "))
							c.Writer.Write(b)
							c.Writer.Write([]byte("\n\n"))
							c.Writer.(http.Flusher).Flush()
						}
					}
				}
			}
		}

		// Usage delta
		if chunk.Usage != nil {
			usage = chunk.Usage
			msgDelta := map[string]any{
				"type": "message_delta",
				"usage": map[string]any{
					"input_tokens":  usage.PromptTokens,
					"output_tokens": usage.CompletionTokens,
				},
			}
			if b, e := json.Marshal(msgDelta); e == nil {
				c.Writer.Write([]byte("data: "))
				c.Writer.Write(b)
				c.Writer.Write([]byte("\n\n"))
				c.Writer.(http.Flusher).Flush()
			}
		}
	}

	// Close any started blocks
	if thinkingIndex >= 0 {
		stop := map[string]any{"type": "content_block_stop", "index": thinkingIndex}
		if b, e := json.Marshal(stop); e == nil {
			c.Writer.Write([]byte("data: "))
			c.Writer.Write(b)
			c.Writer.Write([]byte("\n\n"))
			c.Writer.(http.Flusher).Flush()
		}
	}
	if textIndex >= 0 {
		stop := map[string]any{"type": "content_block_stop", "index": textIndex}
		if b, e := json.Marshal(stop); e == nil {
			c.Writer.Write([]byte("data: "))
			c.Writer.Write(b)
			c.Writer.Write([]byte("\n\n"))
			c.Writer.(http.Flusher).Flush()
		}
	}
	for _, idx := range toolStarted {
		stop := map[string]any{"type": "content_block_stop", "index": idx}
		if b, e := json.Marshal(stop); e == nil {
			c.Writer.Write([]byte("data: "))
			c.Writer.Write(b)
			c.Writer.Write([]byte("\n\n"))
			c.Writer.(http.Flusher).Flush()
		}
	}

	// Finalize usage if upstream omitted
	if usage == nil {
		completion := CountTokenText(accumText, modelName) + CountTokenText(accumThinking, modelName) + CountTokenText(accumToolArgs, modelName)
		usage = &relaymodel.Usage{PromptTokens: promptTokens, CompletionTokens: completion, TotalTokens: promptTokens + completion}
	} else if usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	// message_stop and [DONE]
	msgStop := map[string]any{"type": "message_stop"}
	if b, e := json.Marshal(msgStop); e == nil {
		c.Writer.Write([]byte("data: "))
		c.Writer.Write(b)
		c.Writer.Write([]byte("\n\n"))
		c.Writer.(http.Flusher).Flush()
	}
	c.Writer.Write([]byte("data: [DONE]\n\n"))
	c.Writer.(http.Flusher).Flush()
	_ = resp.Body.Close()
	return usage, nil
}

// --- Minimal local response types to avoid import cycles ---

type responseAPIResponse struct {
	Id        string              `json:"id"`
	Object    string              `json:"object"`
	Model     string              `json:"model"`
	Output    []responseAPIOutput `json:"output"`
	Usage     *responseAPIUsage   `json:"usage,omitempty"`
	CreatedAt int64               `json:"created_at"`
	Status    string              `json:"status"`
}

type responseAPIUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type responseAPIOutput struct {
	Type      string                 `json:"type"`
	Role      string                 `json:"role,omitempty"`
	Content   []responseAPIContent   `json:"content,omitempty"`
	Summary   []responseAPIContent   `json:"summary,omitempty"`
	CallId    string                 `json:"call_id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Arguments string                 `json:"arguments,omitempty"`
	Tools     []relaymodel.Tool      `json:"tools,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type responseAPIContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type chatTextResponse struct {
	Id      string           `json:"id"`
	Model   string           `json:"model"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Choices []chatTextChoice `json:"choices"`
	Usage   relaymodel.Usage `json:"usage"`
}

type chatTextChoice struct {
	Index        int                `json:"index"`
	Message      relaymodel.Message `json:"message"`
	FinishReason string             `json:"finish_reason"`
}
