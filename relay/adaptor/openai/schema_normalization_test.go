package openai

import (
	"reflect"
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
)

func TestNormalizeStructuredJSONSchema_RemovesNumericBounds(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"score": map[string]any{
				"type":             "number",
				"minimum":          0,
				"maximum":          1,
				"exclusiveMinimum": 0,
			},
		},
	}

	normalized, changed := NormalizeStructuredJSONSchema(schema, channeltype.OpenAI)
	if !changed {
		t.Fatalf("expected schema normalization to report changes")
	}

	props, ok := normalized["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected properties map, got %T", normalized["properties"])
	}

	score, ok := props["score"].(map[string]any)
	if !ok {
		t.Fatalf("expected score map, got %T", props["score"])
	}

	if _, exists := score["minimum"]; exists {
		t.Fatalf("minimum key should be removed")
	}
	if _, exists := score["maximum"]; exists {
		t.Fatalf("maximum key should be removed")
	}
	if _, exists := score["exclusiveMinimum"]; exists {
		t.Fatalf("exclusiveMinimum key should be removed")
	}
}

func TestNormalizeStructuredJSONSchema_AzureAddsAdditionalProperties(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"topic": map[string]any{"type": "string"},
		},
	}

	normalized, changed := NormalizeStructuredJSONSchema(schema, channeltype.Azure)
	if !changed {
		t.Fatalf("expected azure normalization to set additionalProperties=false")
	}

	if val, ok := normalized["additionalProperties"].(bool); !ok || val {
		t.Fatalf("expected additionalProperties=false, got %v (type %T)", normalized["additionalProperties"], normalized["additionalProperties"])
	}
}

func TestNormalizeStructuredJSONSchema_NoChangeWhenClean(t *testing.T) {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"topic": map[string]any{"type": "string"},
		},
	}

	copy := reflect.ValueOf(schema).Interface().(map[string]any)
	normalized, changed := NormalizeStructuredJSONSchema(schema, channeltype.OpenAI)
	if changed {
		t.Fatalf("expected no changes for already clean schema")
	}

	if !reflect.DeepEqual(normalized, copy) {
		t.Fatalf("schema should remain unchanged")
	}
}
