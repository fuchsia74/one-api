package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
)

func TestGetTokenKeyParts_ConfiguredPrefix(t *testing.T) {
	old := config.TokenKeyPrefix
	config.TokenKeyPrefix = "sk-"
	defer func() { config.TokenKeyPrefix = old }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer sk-abc-123")
	c.Request = req

	parts := GetTokenKeyParts(c)
	if len(parts) < 2 || parts[0] != "abc" || parts[1] != "123" {
		t.Fatalf("unexpected parts: %#v", parts)
	}
}

func TestGetTokenKeyParts_LegacyPrefix(t *testing.T) {
	old := config.TokenKeyPrefix
	config.TokenKeyPrefix = "custom-"
	defer func() { config.TokenKeyPrefix = old }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer sk-abc-456")
	c.Request = req

	parts := GetTokenKeyParts(c)
	if len(parts) < 2 || parts[0] != "abc" || parts[1] != "456" {
		t.Fatalf("unexpected parts for legacy: %#v", parts)
	}
}
