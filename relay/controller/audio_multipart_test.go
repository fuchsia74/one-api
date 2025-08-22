package controller

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestExtractAudioModelFromMultipart verifies that we correctly parse the `model` field
// from multipart/form-data for audio transcription/translation requests.
func TestExtractAudioModelFromMultipart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()

	// Build a multipart body with a dummy file field and model
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	// file part (content is irrelevant for this test)
	fw, err := writer.CreateFormFile("file", "a.mp3")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = fw.Write([]byte("dummy"))

	// model field
	if err := writer.WriteField("model", "gpt-4o-mini-transcribe"); err != nil {
		t.Fatalf("write field: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	got := extractAudioModelFromMultipart(c)
	if got != "gpt-4o-mini-transcribe" {
		t.Fatalf("expected model 'gpt-4o-mini-transcribe', got '%s'", got)
	}
}
