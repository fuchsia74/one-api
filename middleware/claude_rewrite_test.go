package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Canonical handler
	v1 := r.Group("/v1")
	v1.POST("/messages", func(c *gin.Context) { c.String(200, "ok") })
	return r
}

func TestRewriteClaudeMessagesPrefix(t *testing.T) {
	engine := setupTestEngine()
	engine.Use(RewriteClaudeMessagesPrefix("/v1/v1/messages", engine))
	engine.Use(RewriteClaudeMessagesPrefix("/openai/v1/messages", engine))
	engine.Use(RewriteClaudeMessagesPrefix("/openai/v1/v1/messages", engine))
	engine.Use(RewriteClaudeMessagesPrefix("/api/v1/v1/messages", engine))

	cases := []string{
		"/v1/v1/messages",
		"/openai/v1/messages",
		"/openai/v1/v1/messages",
		"/api/v1/v1/messages",
	}

	for _, path := range cases {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, w.Code)
		}
		if body := w.Body.String(); body != "ok" {
			t.Fatalf("expected body 'ok' for %s, got %q", path, body)
		}
	}
}

func TestRewriteNonMatchingPassThrough(t *testing.T) {
	engine := setupTestEngine()
	engine.Use(RewriteClaudeMessagesPrefix("/v1/v1/messages", engine))
	// Register an unrelated route to ensure pass-through works
	engine.GET("/healthz", func(c *gin.Context) { c.String(200, "healthy") })

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != "healthy" {
		t.Fatalf("unexpected body: %q", body)
	}
}
