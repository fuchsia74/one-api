package controller

import (
	"context"
	"net/http"
	"testing"
)

// Test classification helper directly
func TestIsClientContextCancel(t *testing.T) {
	if !isClientContextCancel(http.StatusInternalServerError, context.Canceled) {
		t.Errorf("expected true for context.Canceled")
	}
	if !isClientContextCancel(http.StatusInternalServerError, context.DeadlineExceeded) {
		t.Errorf("expected true for context.DeadlineExceeded")
	}
	if !isClientContextCancel(http.StatusRequestTimeout, nil) {
		t.Errorf("expected true for 408 even if rawErr is nil")
	}
	if isClientContextCancel(http.StatusInternalServerError, nil) {
		t.Errorf("expected false for 500 with nil rawErr")
	}
}
