package common

import (
	"testing"
)

// TestGenerateVerificationCodeLength tests that GenerateVerificationCode returns a code of the requested length.
func TestGenerateVerificationCodeLength(t *testing.T) {
	lengths := []int{0, 1, 4, 8, 16, 32}
	for _, length := range lengths {
		code := GenerateVerificationCode(length)
		if length == 0 {
			// Should return full UUID (length 32)
			if len(code) != 32 {
				t.Errorf("Expected code length 32 for length=0, got %d", len(code))
			}
		} else {
			if len(code) != length {
				t.Errorf("Expected code length %d, got %d", length, len(code))
			}
		}
	}
}

// TestGenerateVerificationCodeUniqueness tests that GenerateVerificationCode generates unique codes.
func TestGenerateVerificationCodeUniqueness(t *testing.T) {
	codes := make(map[string]struct{})
	for range 100 {
		code := GenerateVerificationCode(8)
		if _, exists := codes[code]; exists {
			t.Errorf("Duplicate code generated: %s", code)
		}
		codes[code] = struct{}{}
	}
}

// TestGenerateVerificationCodeZeroLength tests that length=0 returns a valid UUID.
func TestGenerateVerificationCodeZeroLength(t *testing.T) {
	code := GenerateVerificationCode(0)
	if len(code) != 32 {
		t.Errorf("Expected UUID length 32 for length=0, got %d", len(code))
	}
}
