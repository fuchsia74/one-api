package anthropic

import (
	"encoding/json"
	"testing"
)

// Test that an Anthropic SSE error event parses into StreamResponse with Error fields populated.
func TestParseStreamErrorEvent(t *testing.T) {
	payload := `{"type":"error","error":{"details":null,"type":"overloaded_error","message":"Overloaded"},"request_id":"req_011CSoQiKNJjFPYGZGMask1g"}`
	var sr StreamResponse
	if err := json.Unmarshal([]byte(payload), &sr); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if sr.Type != "error" {
		t.Fatalf("expected type error, got %s", sr.Type)
	}
	if sr.Error.Type != "overloaded_error" {
		t.Fatalf("expected error.type overloaded_error, got %s", sr.Error.Type)
	}
	if sr.Error.Message != "Overloaded" {
		t.Fatalf("expected error.message Overloaded, got %s", sr.Error.Message)
	}
	if sr.RequestId == "" {
		t.Fatalf("expected request_id to be set")
	}
}
