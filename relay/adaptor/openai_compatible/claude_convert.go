package openai_compatible

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

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
