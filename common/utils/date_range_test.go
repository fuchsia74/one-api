package utils

import (
	"testing"
	"time"
)

func TestNormalizeDateRange(t *testing.T) {
	t.Run("single day", func(t *testing.T) {
		s, e, err := NormalizeDateRange("2025-01-15", "2025-01-15", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e-s != 24*3600 {
			t.Fatalf("expected 1 day span, got %d", e-s)
		}
	})

	t.Run("multi day inclusive", func(t *testing.T) {
		s, e, err := NormalizeDateRange("2025-01-01", "2025-01-03", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e-s != 3*24*3600 {
			t.Fatalf("expected 3 day span, got %d", e-s)
		}
	})

	t.Run("leap day", func(t *testing.T) {
		s, e, err := NormalizeDateRange("2024-02-28", "2024-03-01", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e-s != 3*24*3600 {
			t.Fatalf("expected 3 day span across leap day, got %d", e-s)
		}
	})

	t.Run("max days exceeded", func(t *testing.T) {
		_, _, err := NormalizeDateRange("2025-01-01", "2025-01-10", 5)
		if err == nil {
			t.Fatalf("expected error for exceeding max days")
		}
	})

	t.Run("invalid order", func(t *testing.T) {
		_, _, err := NormalizeDateRange("2025-01-10", "2025-01-01", 10)
		if err == nil {
			t.Fatalf("expected error for reversed dates")
		}
	})
}

// Ensure UTC correctness by comparing boundaries explicitly.
func TestNormalizeDateRangeUTC(t *testing.T) {
	s, e, err := NormalizeDateRange("2025-05-05", "2025-05-05", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if time.Unix(s, 0).UTC().Hour() != 0 {
		t.Fatalf("start not at midnight UTC")
	}
	if time.Unix(e, 0).UTC().Hour() != 0 {
		t.Fatalf("endExclusive not at midnight UTC")
	}
}
