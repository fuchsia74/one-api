package model

import "testing"

// Test that ParseContent preserves image detail field for accurate token billing
func TestParseContent_ImageDetailPreserved(t *testing.T) {
	m := Message{
		Role: "user",
		Content: []any{
			map[string]any{
				"type": "image_url",
				"image_url": map[string]any{
					"url":    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAAB",
					"detail": "low",
				},
			},
		},
	}

	parts := m.ParseContent()
	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}
	if parts[0].ImageURL == nil {
		t.Fatalf("expected image URL part")
	}
	if parts[0].ImageURL.Detail != "low" {
		t.Fatalf("expected detail 'low', got '%s'", parts[0].ImageURL.Detail)
	}
}
