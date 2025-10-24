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

func TestMessageStringContent_OutputJSON(t *testing.T) {
	m := Message{
		Role: "assistant",
		Content: []any{
			map[string]any{
				"type":         "output_json_delta",
				"partial_json": "{\"topic\":\"AI\"",
			},
			map[string]any{
				"type":         "output_json_delta",
				"partial_json": ",\"confidence\":0.9}",
			},
		},
	}

	if got := m.StringContent(); got != "{\"topic\":\"AI\",\"confidence\":0.9}" {
		t.Fatalf("unexpected string content: %s", got)
	}

	parts := m.ParseContent()
	if len(parts) != 1 {
		t.Fatalf("expected 1 content part, got %d", len(parts))
	}
	if parts[0].Text == nil || *parts[0].Text != "{\"topic\":\"AI\",\"confidence\":0.9}" {
		t.Fatalf("expected JSON text to be preserved, got %+v", parts[0].Text)
	}
}
