package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/adaptor/openai_compatible"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func TestRenderChatResponseAsResponseAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(ctxkey.RequestId, "req_123")

	textResp := &openai_compatible.SlimTextResponse{
		Choices: []openai_compatible.TextResponseChoice{
			{
				Message:      relaymodel.Message{Role: "assistant", Content: "Hello there"},
				FinishReason: "stop",
			},
		},
		Usage: relaymodel.Usage{PromptTokens: 12, CompletionTokens: 8, TotalTokens: 20},
	}

	parallel := true
	request := &openai.ResponseAPIRequest{ParallelToolCalls: &parallel}
	meta := &metalib.Meta{ActualModelName: "gpt-fallback"}

	if err := renderChatResponseAsResponseAPI(c, http.StatusOK, textResp, request, meta); err != nil {
		t.Fatalf("unexpected error rendering response: %v", err)
	}

	var resp openai.ResponseAPIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if resp.Model != "gpt-fallback" {
		t.Fatalf("expected model gpt-fallback, got %s", resp.Model)
	}
	if resp.Status != "completed" {
		t.Fatalf("expected status completed, got %s", resp.Status)
	}
	if len(resp.Output) != 1 {
		t.Fatalf("expected single output item, got %d", len(resp.Output))
	}
	if resp.Output[0].Type != "message" {
		t.Fatalf("expected message output, got %s", resp.Output[0].Type)
	}
	if len(resp.Output[0].Content) == 0 || resp.Output[0].Content[0].Text != "Hello there" {
		t.Fatalf("expected message content preserved, got %#v", resp.Output[0].Content)
	}
	if resp.Usage == nil || resp.Usage.TotalTokens != 20 {
		t.Fatalf("expected usage to be carried over, got %#v", resp.Usage)
	}
	if !resp.ParallelToolCalls {
		t.Fatalf("expected parallel tool calls to be true")
	}
}
