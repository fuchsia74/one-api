package controller

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestResponseStatus ensures nil responses are handled without panics and return zero status.
func TestResponseStatus(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		require.Equal(t, 0, responseStatus(nil))
	})

	t.Run("non-nil response", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusTeapot}
		require.Equal(t, http.StatusTeapot, responseStatus(resp))
	})
}
