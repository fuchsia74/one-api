package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// RewriteClaudeMessagesPrefix returns a middleware that rewrites any request path
// starting with the given prefix to the canonical Claude Messages endpoint: /v1/messages.
// It then re-dispatches the request to the engine so the canonical route and its
// middlewares handle it, and aborts the current handler chain.
//
// Example prefixes observed in the wild:
//   - /v1/v1/messages
//   - /openai/v1/messages
//   - /openai/v1/v1/messages
//   - /api/v1/v1/messages
func RewriteClaudeMessagesPrefix(prefix string, engine *gin.Engine) gin.HandlerFunc {
	normalized := strings.TrimSuffix(prefix, "/")

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// Only handle when the incoming path actually matches the prefix.
		if strings.HasPrefix(path, normalized) {
			// Rewrite the request path to the canonical endpoint.
			c.Request.URL.Path = "/v1/messages"
			// Re-dispatch to let the canonical route handle the request.
			engine.HandleContext(c)
			// Stop further processing in the current chain.
			c.Abort()
			return
		}

		c.Next()
	}
}
