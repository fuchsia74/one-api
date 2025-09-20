package openai

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// faultyWriter simulates a client disconnect by returning an error on Write
type faultyWriter struct{ gin.ResponseWriter }

func (fw faultyWriter) Write(b []byte) (int, error) { return 0, errors.New("client disconnected") }

func TestHandlerReturnsUsageOnWriteFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Wrap the recorder with faulty writer to simulate disconnect
	c.Writer = faultyWriter{ResponseWriter: c.Writer}

	// Build a minimal upstream response with usage
	body := []byte(`{"choices":[{"message":{"content":"hi"}}],"usage":{"prompt_tokens":5,"completion_tokens":7,"total_tokens":12}}`)
	resp := &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body)), Header: http.Header{}}

	err, usage := Handler(c, resp, 5, "gpt-4o")
	if err == nil {
		t.Fatalf("expected error due to write failure, got nil")
	}
	if usage == nil {
		t.Fatalf("expected usage to be returned on write failure")
	}
	if usage.PromptTokens != 5 || usage.CompletionTokens == 0 {
		t.Fatalf("unexpected usage: %+v", *usage)
	}
}
