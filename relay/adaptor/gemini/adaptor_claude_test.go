package gemini

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

func TestConvertNonStreamingToClaudeResponse_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adaptor := &Adaptor{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	geminiResp := ChatResponse{
		Candidates: []ChatCandidate{
			{
				FinishReason: "STOP",
				Content: ChatContent{
					Parts: []Part{{Text: "Hello from Gemini"}},
				},
			},
		},
		UsageMetadata: &UsageMetadata{
			PromptTokenCount:     5,
			CandidatesTokenCount: 7,
			TotalTokenCount:      12,
		},
	}

	bodyBytes, err := json.Marshal(geminiResp)
	if err != nil {
		t.Fatalf("marshal gemini response: %v", err)
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(bodyBytes)),
	}
	resp.Header.Set("Content-Type", "application/json")

	metaInfo := &meta.Meta{ActualModelName: "gemini-1.5-flash", PromptTokens: 5}

	newResp, errResp := adaptor.convertNonStreamingToClaudeResponse(ctx, resp, bodyBytes, metaInfo)
	if errResp != nil {
		t.Fatalf("convert non-streaming returned error: %v", errResp.Error)
	}
	defer newResp.Body.Close()

	convertedBody, err := io.ReadAll(newResp.Body)
	if err != nil {
		t.Fatalf("read converted body: %v", err)
	}

	var claudeResp model.ClaudeResponse
	if err := json.Unmarshal(convertedBody, &claudeResp); err != nil {
		t.Fatalf("unmarshal claude response: %v", err)
	}

	if len(claudeResp.Content) == 0 || claudeResp.Content[0].Text != "Hello from Gemini" {
		t.Fatalf("unexpected content: %+v", claudeResp.Content)
	}
	if claudeResp.StopReason != "end_turn" {
		t.Fatalf("unexpected stop reason: %s", claudeResp.StopReason)
	}
	if claudeResp.Usage.InputTokens != 5 {
		t.Fatalf("unexpected input tokens: %d", claudeResp.Usage.InputTokens)
	}
	if claudeResp.Usage.OutputTokens != 7 {
		t.Fatalf("unexpected output tokens: %d", claudeResp.Usage.OutputTokens)
	}
}

func TestConvertStreamingToClaudeResponse_Basic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adaptor := &Adaptor{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	streamChunk := ChatResponse{
		Candidates: []ChatCandidate{
			{
				FinishReason: "STOP",
				Content: ChatContent{
					Parts: []Part{{Text: "Hello from Gemini"}},
				},
			},
		},
		UsageMetadata: &UsageMetadata{
			PromptTokenCount:     4,
			CandidatesTokenCount: 3,
			TotalTokenCount:      7,
		},
	}

	chunkBytes, err := json.Marshal(streamChunk)
	if err != nil {
		t.Fatalf("marshal stream chunk: %v", err)
	}

	streamBuf := bytes.NewBuffer(nil)
	streamBuf.WriteString("data: ")
	streamBuf.Write(chunkBytes)
	streamBuf.WriteString("\n\n")
	streamBuf.WriteString("data: [DONE]\n\n")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(streamBuf.Bytes())),
	}
	resp.Header.Set("Content-Type", "text/event-stream")

	metaInfo := &meta.Meta{ActualModelName: "gemini-1.5-flash", PromptTokens: 4}

	newResp, errResp := adaptor.convertStreamingToClaudeResponse(ctx, resp, streamBuf.Bytes(), metaInfo)
	if errResp != nil {
		t.Fatalf("convert streaming returned error: %v", errResp.Error)
	}
	defer newResp.Body.Close()

	if ct := newResp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("unexpected content type: %s", ct)
	}

	convertedBody, err := io.ReadAll(newResp.Body)
	if err != nil {
		t.Fatalf("read converted stream: %v", err)
	}

	converted := string(convertedBody)
	if !strings.Contains(converted, "event: message_start") {
		t.Fatalf("missing message_start event: %s", converted)
	}
	if !strings.Contains(converted, "\"text_delta\"") {
		t.Fatalf("missing text delta: %s", converted)
	}
	if !strings.Contains(converted, "Hello from Gemini") {
		t.Fatalf("missing response text: %s", converted)
	}
	if !strings.Contains(converted, "\"input_tokens\":4") {
		t.Fatalf("missing usage input tokens: %s", converted)
	}
	if !strings.Contains(converted, "data: [DONE]") {
		t.Fatalf("missing done marker: %s", converted)
	}
}
