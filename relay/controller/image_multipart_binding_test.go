package controller

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetImageRequest_MultipartBindsModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("prompt", "a cat"); err != nil {
		t.Fatalf("write prompt: %v", err)
	}
	if err := writer.WriteField("model", "gpt-image-1"); err != nil {
		t.Fatalf("write model: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	imgReq, err := getImageRequest(c, 0)
	if err != nil {
		t.Fatalf("getImageRequest error: %v", err)
	}
	if imgReq == nil || imgReq.Model != "gpt-image-1" {
		t.Fatalf("expected model 'gpt-image-1', got '%v'", imgReq)
	}
}
