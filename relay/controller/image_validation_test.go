package controller

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	metalib "github.com/songquanpeng/one-api/relay/meta"
)

func TestValidateImageRequest_DALLE3_RejectAutoQuality(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := []byte(`{"model":"dall-e-3","prompt":"p","size":"1024x1024","quality":"auto"}`)
	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	ir, err := getImageRequest(c, 0)
	if err != nil {
		t.Fatalf("getImageRequest error: %v", err)
	}

	// meta not used for validation currently
	if got := validateImageRequest(ir, metalib.GetByContext(c)); got == nil {
		t.Fatalf("expected validation error for quality=auto")
	} else if got.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", got.StatusCode)
	}
}
