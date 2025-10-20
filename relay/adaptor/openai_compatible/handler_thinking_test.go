package openai_compatible

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/model"
)

// TestHandler_NonStream_ThinkingParam verifies non-stream handler respects thinking and reasoning_format
func TestHandler_NonStream_ThinkingParam(t *testing.T) {
	// Build a simple non-stream response with a single choice containing <think>
	respStruct := SlimTextResponse{
		Choices: []TextResponseChoice{
			{
				Index:        0,
				Message:      structToMessage("before <think>xyz</think> after"),
				FinishReason: "stop",
			},
		},
		Usage: modelUsage(0, 0),
	}
	b, _ := json.Marshal(respStruct)

	// Prepare gin context with query params
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/v1/chat/completions?thinking=true&reasoning_format=reasoning_content", nil)
	c.Request = req

	// Fake upstream http.Response
	upstream := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(b))),
	}

	if err, _ := Handler(c, upstream, 0, "gpt-4"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Decode written JSON
	var out SlimTextResponse
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to unmarshal out: %v", err)
	}

	if len(out.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(out.Choices))
	}
	msg := out.Choices[0].Message

	// Expect content cleaned of think tags
	if got := msg.StringContent(); got != "before  after" {
		t.Fatalf("unexpected cleaned content: %q", got)
	}
	// Expect reasoning mapped to reasoning_content field
	if msg.ReasoningContent == nil || *msg.ReasoningContent != "xyz" {
		t.Fatalf("expected reasoning_content=xyz, got %#v", msg.ReasoningContent)
	}
}

// TestHandler_NonStream_OmitsEmptyErrorField verifies that the handler does not emit the error field
// when upstream responses omit it, preserving OpenAI compatibility for clients that gate on its presence.
func TestHandler_NonStream_OmitsEmptyErrorField(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	c.Request = req

	respStruct := SlimTextResponse{
		Choices: []TextResponseChoice{
			{
				Index:        0,
				Message:      structToMessage("plain response"),
				FinishReason: "stop",
			},
		},
		Usage: modelUsage(5, 7),
	}
	b, _ := json.Marshal(respStruct)

	upstream := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(string(b))),
	}

	if err, _ := Handler(c, upstream, 0, "gpt-4"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to decode handler output: %v", err)
	}
	if _, exists := out["error"]; exists {
		t.Fatalf("expected no error field in handler output, got %s", w.Body.String())
	}
}

// Helpers
func structToMessage(s string) model.Message {
	return model.Message{Content: s, Role: "assistant"}
}

func modelUsage(p, c int) model.Usage {
	return model.Usage{PromptTokens: p, CompletionTokens: c, TotalTokens: p + c}
}
