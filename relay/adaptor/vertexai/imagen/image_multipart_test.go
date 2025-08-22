package imagen

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestConvertMultipartImageEditRequest ensures we can parse required fields from multipart form
// and construct an Imagen create request without panicking.
func TestConvertMultipartImageEditRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()

	// Build a multipart body with image, mask, prompt, model and response_format=b64_json
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// required file parts
	imgPart, err := writer.CreateFormFile("image", "img.png")
	if err != nil {
		t.Fatalf("create image part: %v", err)
	}
	_, _ = imgPart.Write([]byte("PNG"))

	maskPart, err := writer.CreateFormFile("mask", "mask.png")
	if err != nil {
		t.Fatalf("create mask part: %v", err)
	}
	_, _ = maskPart.Write([]byte("PNG"))

	// required fields
	if err := writer.WriteField("prompt", "Edit this image"); err != nil {
		t.Fatalf("write prompt: %v", err)
	}
	if err := writer.WriteField("model", "imagen-3.0"); err != nil {
		t.Fatalf("write model: %v", err)
	}
	if err := writer.WriteField("response_format", "b64_json"); err != nil {
		t.Fatalf("write response_format: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/v1/images/edits", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	converted, err := ConvertMultipartImageEditRequest(c)
	if err != nil {
		t.Fatalf("ConvertMultipartImageEditRequest error: %v", err)
	}
	if converted == nil || len(converted.Instances) == 0 {
		t.Fatalf("expected non-nil converted request with instances")
	}
}
