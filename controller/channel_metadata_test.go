package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetChannelMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/channel/metadata", GetChannelMetadata)

	req, _ := http.NewRequest(http.MethodGet, "/api/channel/metadata?type=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if body == "" || (len(body) > 0 && body[0] != '{') {
		t.Fatalf("unexpected body: %s", body)
	}
}
