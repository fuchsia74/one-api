package openai

import (
	"encoding/json"
	"testing"
)

func TestResponseAPIInputUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected ResponseAPIInput
	}{
		{
			name:     "String input",
			jsonData: `{"input": "Write a one-sentence bedtime story about a unicorn."}`,
			expected: ResponseAPIInput{"Write a one-sentence bedtime story about a unicorn."},
		},
		{
			name:     "Array input with single string",
			jsonData: `{"input": ["Hello world"]}`,
			expected: ResponseAPIInput{"Hello world"},
		},
		{
			name:     "Array input with message object",
			jsonData: `{"input": [{"role": "user", "content": "Hello"}]}`,
			expected: ResponseAPIInput{map[string]any{
				"role":    "user",
				"content": "Hello",
			}},
		},
		{
			name:     "Array input with multiple items",
			jsonData: `{"input": ["Hello", {"role": "user", "content": "World"}]}`,
			expected: ResponseAPIInput{
				"Hello",
				map[string]any{
					"role":    "user",
					"content": "World",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request struct {
				Input ResponseAPIInput `json:"input"`
			}

			err := json.Unmarshal([]byte(tt.jsonData), &request)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			if len(request.Input) != len(tt.expected) {
				t.Errorf("Expected input length %d, got %d", len(tt.expected), len(request.Input))
				return
			}

			for i, expectedItem := range tt.expected {
				actualItem := request.Input[i]

				// For string comparison
				if expectedStr, ok := expectedItem.(string); ok {
					if actualStr, ok := actualItem.(string); ok {
						if expectedStr != actualStr {
							t.Errorf("Expected input[%d] to be %q, got %q", i, expectedStr, actualStr)
						}
					} else {
						t.Errorf("Expected input[%d] to be string, got %T", i, actualItem)
					}
					continue
				}

				// For map comparison (simplified)
				if expectedMap, ok := expectedItem.(map[string]any); ok {
					if actualMap, ok := actualItem.(map[string]any); ok {
						for key, expectedValue := range expectedMap {
							if actualValue, exists := actualMap[key]; !exists || actualValue != expectedValue {
								t.Errorf("Expected input[%d][%s] to be %v, got %v", i, key, expectedValue, actualValue)
							}
						}
					} else {
						t.Errorf("Expected input[%d] to be map, got %T", i, actualItem)
					}
				}
			}
		})
	}
}

func TestResponseAPIRequestValidation(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid string input",
			jsonData: `{
				"model": "gpt-4o-mini",
				"input": "Write a story"
			}`,
			expectError: false,
		},
		{
			name: "Valid array input",
			jsonData: `{
				"model": "gpt-4o-mini",
				"input": [{"role": "user", "content": "Write a story"}]
			}`,
			expectError: false,
		},
		{
			name: "Valid prompt input",
			jsonData: `{
				"model": "gpt-4o-mini",
				"prompt": {
					"id": "pmpt_123",
					"variables": {"name": "John"}
				}
			}`,
			expectError: false,
		},
		{
			name: "Invalid - both input and prompt",
			jsonData: `{
				"model": "gpt-4o-mini",
				"input": "Write a story",
				"prompt": {"id": "pmpt_123"}
			}`,
			expectError: false, // JSON unmarshaling will succeed, validation happens later
		},
		{
			name: "Invalid - neither input nor prompt",
			jsonData: `{
				"model": "gpt-4o-mini"
			}`,
			expectError: false, // JSON unmarshaling will succeed, validation happens later
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request ResponseAPIRequest
			err := json.Unmarshal([]byte(tt.jsonData), &request)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Basic validation that the struct was populated correctly
			if err == nil {
				if request.Model == "" {
					t.Errorf("Model should not be empty")
				}
			}
		})
	}
}
