package openai

import (
	"encoding/json"
	"testing"
)

// TestResponseAPIRequestParsing tests comprehensive parsing of ResponseAPIRequest
// with all different request body formats outlined in the Response API documentation
func TestResponseAPIRequestParsing(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
		validate    func(t *testing.T, req *ResponseAPIRequest)
	}{
		{
			name: "Simple string input",
			jsonData: `{
				"model": "gpt-4.1",
				"input": "Write a one-sentence bedtime story about a unicorn."
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if req.Model != "gpt-4.1" {
					t.Errorf("Expected model 'gpt-4.1', got '%s'", req.Model)
				}
				if len(req.Input) != 1 {
					t.Errorf("Expected input length 1, got %d", len(req.Input))
				}
				if req.Input[0] != "Write a one-sentence bedtime story about a unicorn." {
					t.Errorf("Unexpected input content: %v", req.Input[0])
				}
			},
		},
		{
			name: "String input with instructions",
			jsonData: `{
				"model": "gpt-4.1",
				"instructions": "Talk like a pirate.",
				"input": "Are semicolons optional in JavaScript?"
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if req.Instructions == nil || *req.Instructions != "Talk like a pirate." {
					t.Errorf("Expected instructions 'Talk like a pirate.', got %v", req.Instructions)
				}
				if req.Input[0] != "Are semicolons optional in JavaScript?" {
					t.Errorf("Unexpected input content: %v", req.Input[0])
				}
			},
		},
		{
			name: "Array input with message roles",
			jsonData: `{
				"model": "gpt-4.1",
				"input": [
					{
						"role": "developer",
						"content": "Talk like a pirate."
					},
					{
						"role": "user",
						"content": "Are semicolons optional in JavaScript?"
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if len(req.Input) != 2 {
					t.Errorf("Expected input length 2, got %d", len(req.Input))
				}
				// Verify first message
				msg1, ok := req.Input[0].(map[string]any)
				if !ok {
					t.Errorf("Expected first input to be a message object")
				} else {
					if msg1["role"] != "developer" {
						t.Errorf("Expected first message role 'developer', got '%v'", msg1["role"])
					}
					if msg1["content"] != "Talk like a pirate." {
						t.Errorf("Expected first message content 'Talk like a pirate.', got '%v'", msg1["content"])
					}
				}
			},
		},
		{
			name: "Prompt template with variables",
			jsonData: `{
				"model": "gpt-4.1",
				"prompt": {
					"id": "pmpt_abc123",
					"version": "2",
					"variables": {
						"customer_name": "Jane Doe",
						"product": "40oz juice box"
					}
				}
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if req.Prompt == nil {
					t.Fatal("Expected prompt to be set")
				}
				if req.Prompt.Id != "pmpt_abc123" {
					t.Errorf("Expected prompt id 'pmpt_abc123', got '%s'", req.Prompt.Id)
				}
				if req.Prompt.Version == nil || *req.Prompt.Version != "2" {
					t.Errorf("Expected prompt version '2', got %v", req.Prompt.Version)
				}
				if req.Prompt.Variables["customer_name"] != "Jane Doe" {
					t.Errorf("Expected customer_name 'Jane Doe', got '%v'", req.Prompt.Variables["customer_name"])
				}
			},
		},
		{
			name: "Image input with URL",
			jsonData: `{
				"model": "gpt-4.1-mini",
				"input": [
					{
						"role": "user",
						"content": [
							{"type": "input_text", "text": "what's in this image?"},
							{
								"type": "input_image",
								"image_url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
							}
						]
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if len(req.Input) != 1 {
					t.Errorf("Expected input length 1, got %d", len(req.Input))
				}
				msg, ok := req.Input[0].(map[string]any)
				if !ok {
					t.Fatal("Expected input to be a message object")
				}
				content, ok := msg["content"].([]any)
				if !ok {
					t.Fatal("Expected content to be an array")
				}
				if len(content) != 2 {
					t.Errorf("Expected content length 2, got %d", len(content))
				}
			},
		},
		{
			name: "Image input with base64",
			jsonData: `{
				"model": "gpt-4.1-mini",
				"input": [
					{
						"role": "user",
						"content": [
							{"type": "input_text", "text": "what's in this image?"},
							{
								"type": "input_image",
								"image_url": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAYEBQYFBAYGBQYHBwYIChAKCgkJChQODwwQFxQYGBcUFhYaHSUfGhsjHBYWICwgIyYnKSopGR8tMC0oMCUoKSj/2wBDAQcHBwoIChMKChMoGhYaKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCgoKCj/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCdABmX/9k="
							}
						]
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if len(req.Input) != 1 {
					t.Errorf("Expected input length 1, got %d", len(req.Input))
				}
				msg, ok := req.Input[0].(map[string]any)
				if !ok {
					t.Fatal("Expected input to be a message object")
				}
				content, ok := msg["content"].([]any)
				if !ok {
					t.Fatal("Expected content to be an array")
				}
				if len(content) != 2 {
					t.Errorf("Expected content length 2, got %d", len(content))
				}
				// Verify image content
				imageContent, ok := content[1].(map[string]any)
				if !ok {
					t.Fatal("Expected second content item to be an object")
				}
				if imageContent["type"] != "input_image" {
					t.Errorf("Expected type 'input_image', got '%v'", imageContent["type"])
				}
			},
		},
		{
			name: "Image input with file_id",
			jsonData: `{
				"model": "gpt-4.1-mini",
				"input": [
					{
						"role": "user",
						"content": [
							{"type": "input_text", "text": "what's in this image?"},
							{
								"type": "input_image",
								"file_id": "file-abc123"
							}
						]
					}
				]
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if len(req.Input) != 1 {
					t.Errorf("Expected input length 1, got %d", len(req.Input))
				}
				msg, ok := req.Input[0].(map[string]any)
				if !ok {
					t.Fatal("Expected input to be a message object")
				}
				content, ok := msg["content"].([]any)
				if !ok {
					t.Fatal("Expected content to be an array")
				}
				// Verify image content with file_id
				imageContent, ok := content[1].(map[string]any)
				if !ok {
					t.Fatal("Expected second content item to be an object")
				}
				if imageContent["file_id"] != "file-abc123" {
					t.Errorf("Expected file_id 'file-abc123', got '%v'", imageContent["file_id"])
				}
			},
		},
		{
			name: "Tools with image generation",
			jsonData: `{
				"model": "gpt-4.1-mini",
				"input": "Generate an image of gray tabby cat hugging an otter with an orange scarf",
				"tools": [{"type": "image_generation"}]
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if len(req.Tools) != 1 {
					t.Errorf("Expected tools length 1, got %d", len(req.Tools))
				}
				if req.Tools[0].Type != "image_generation" {
					t.Errorf("Expected tool type 'image_generation', got '%s'", req.Tools[0].Type)
				}
			},
		},
		{
			name: "Function tools",
			jsonData: `{
				"model": "gpt-4.1",
				"input": "What's the weather like in Boston?",
				"tools": [
					{
						"type": "function",
						"name": "get_weather",
						"description": "Get current weather",
						"parameters": {
							"type": "object",
							"properties": {
								"location": {
									"type": "string",
									"description": "City name"
								}
							},
							"required": ["location"]
						}
					}
				],
				"tool_choice": "auto"
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if len(req.Tools) != 1 {
					t.Errorf("Expected tools length 1, got %d", len(req.Tools))
				}
				tool := req.Tools[0]
				if tool.Type != "function" {
					t.Errorf("Expected tool type 'function', got '%s'", tool.Type)
				}
				if tool.Name != "get_weather" {
					t.Errorf("Expected tool name 'get_weather', got '%s'", tool.Name)
				}
				if tool.Description != "Get current weather" {
					t.Errorf("Expected tool description 'Get current weather', got '%s'", tool.Description)
				}
				if req.ToolChoice != "auto" {
					t.Errorf("Expected tool_choice 'auto', got '%v'", req.ToolChoice)
				}
			},
		},
		{
			name: "Structured outputs with JSON schema",
			jsonData: `{
				"model": "gpt-4o-2024-08-06",
				"input": [
					{"role": "system", "content": "You are a helpful math tutor. Guide the user through the solution step by step."},
					{"role": "user", "content": "how can I solve 8x + 7 = -23"}
				],
				"text": {
					"format": {
						"type": "json_schema",
						"name": "math_reasoning",
						"schema": {
							"type": "object",
							"properties": {
								"steps": {
									"type": "array",
									"items": {
										"type": "object",
										"properties": {
											"explanation": {"type": "string"},
											"output": {"type": "string"}
										},
										"required": ["explanation", "output"],
										"additionalProperties": false
									}
								},
								"final_answer": {"type": "string"}
							},
							"required": ["steps", "final_answer"],
							"additionalProperties": false
						},
						"strict": true
					}
				}
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if req.Text == nil {
					t.Fatal("Expected text config to be set")
				}
				if req.Text.Format == nil {
					t.Fatal("Expected text format to be set")
				}
				if req.Text.Format.Type != "json_schema" {
					t.Errorf("Expected format type 'json_schema', got '%s'", req.Text.Format.Type)
				}
				if req.Text.Format.Name != "math_reasoning" {
					t.Errorf("Expected format name 'math_reasoning', got '%s'", req.Text.Format.Name)
				}
				if req.Text.Format.Strict == nil || !*req.Text.Format.Strict {
					t.Errorf("Expected strict mode to be true")
				}
				if req.Text.Format.Schema == nil {
					t.Fatal("Expected schema to be set")
				}
			},
		},
		{
			name: "Reasoning model with effort",
			jsonData: `{
				"model": "o3-2025-04-16",
				"input": "Solve this complex math problem: 3x + 7 = 22",
				"reasoning": {
					"effort": "high",
					"summary": "detailed"
				}
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if req.Reasoning == nil {
					t.Fatal("Expected reasoning config to be set")
				}
				if req.Reasoning.Effort == nil || *req.Reasoning.Effort != "high" {
					t.Errorf("Expected reasoning effort 'high', got %v", req.Reasoning.Effort)
				}
				if req.Reasoning.Summary == nil || *req.Reasoning.Summary != "detailed" {
					t.Errorf("Expected reasoning summary 'detailed', got %v", req.Reasoning.Summary)
				}
			},
		},
		{
			name: "Complete request with all optional parameters",
			jsonData: `{
				"model": "gpt-4.1",
				"input": "Tell me a story",
				"background": true,
				"include": ["usage", "metadata"],
				"instructions": "Be creative and engaging",
				"max_output_tokens": 1000,
				"metadata": {"user_id": "123", "session": "abc"},
				"parallel_tool_calls": true,
				"previous_response_id": "resp_456",
				"service_tier": "default",
				"store": true,
				"stream": false,
				"temperature": 0.7,
				"top_p": 0.9,
				"truncation": "auto",
				"user": "test_user"
			}`,
			expectError: false,
			validate: func(t *testing.T, req *ResponseAPIRequest) {
				if req.Background == nil || !*req.Background {
					t.Errorf("Expected background to be true")
				}
				if len(req.Include) != 2 || req.Include[0] != "usage" || req.Include[1] != "metadata" {
					t.Errorf("Expected include to be ['usage', 'metadata'], got %v", req.Include)
				}
				if req.MaxOutputTokens == nil || *req.MaxOutputTokens != 1000 {
					t.Errorf("Expected max_output_tokens 1000, got %v", req.MaxOutputTokens)
				}
				if req.ParallelToolCalls == nil || !*req.ParallelToolCalls {
					t.Errorf("Expected parallel_tool_calls to be true")
				}
				if req.PreviousResponseId == nil || *req.PreviousResponseId != "resp_456" {
					t.Errorf("Expected previous_response_id 'resp_456', got %v", req.PreviousResponseId)
				}
				if req.ServiceTier == nil || *req.ServiceTier != "default" {
					t.Errorf("Expected service_tier 'default', got %v", req.ServiceTier)
				}
				if req.Store == nil || !*req.Store {
					t.Errorf("Expected store to be true")
				}
				if req.Stream == nil || *req.Stream {
					t.Errorf("Expected stream to be false")
				}
				if req.Temperature == nil || *req.Temperature != 0.7 {
					t.Errorf("Expected temperature 0.7, got %v", req.Temperature)
				}
				if req.TopP == nil || *req.TopP != 0.9 {
					t.Errorf("Expected top_p 0.9, got %v", req.TopP)
				}
				if req.Truncation == nil || *req.Truncation != "auto" {
					t.Errorf("Expected truncation 'auto', got %v", req.Truncation)
				}
				if req.User == nil || *req.User != "test_user" {
					t.Errorf("Expected user 'test_user', got %v", req.User)
				}
			},
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

			if err == nil && tt.validate != nil {
				tt.validate(t, &request)
			}
		})
	}
}

// TestResponseAPIRequestEdgeCases tests edge cases and error handling
func TestResponseAPIRequestEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Missing required model field",
			jsonData: `{
				"input": "Hello world"
			}`,
			expectError: false, // JSON unmarshaling doesn't enforce required fields
		},
		{
			name: "Empty input and no prompt",
			jsonData: `{
				"model": "gpt-4.1"
			}`,
			expectError: false, // Valid - input and prompt are optional
		},
		{
			name: "Both input and prompt provided",
			jsonData: `{
				"model": "gpt-4.1",
				"input": "Hello",
				"prompt": {
					"id": "pmpt_123"
				}
			}`,
			expectError: false, // JSON parsing allows both, validation would be at business logic level
		},
		{
			name: "Invalid JSON",
			jsonData: `{
				"model": "gpt-4.1",
				"input": "Hello"
				"missing_comma": true
			}`,
			expectError: true,
		},
		{
			name: "Null values for optional fields",
			jsonData: `{
				"model": "gpt-4.1",
				"input": "Hello",
				"temperature": null,
				"max_output_tokens": null,
				"stream": null
			}`,
			expectError: false,
		},
		{
			name: "Empty arrays and objects",
			jsonData: `{
				"model": "gpt-4.1",
				"input": [],
				"tools": [],
				"include": [],
				"metadata": {}
			}`,
			expectError: false,
		},
		{
			name: "Complex nested structures",
			jsonData: `{
				"model": "gpt-4.1",
				"input": [
					{
						"role": "user",
						"content": [
							{
								"type": "input_text",
								"text": "Analyze this complex data"
							},
							{
								"type": "input_image",
								"image_url": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
								"detail": "high"
							}
						]
					}
				],
				"tools": [
					{
						"type": "function",
						"name": "complex_analysis",
						"description": "Perform complex data analysis",
						"parameters": {
							"type": "object",
							"properties": {
								"data": {
									"type": "array",
									"items": {
										"type": "object",
										"properties": {
											"id": {"type": "string"},
											"value": {"type": "number"},
											"metadata": {
												"type": "object",
												"additionalProperties": true
											}
										}
									}
								}
							}
						}
					}
				],
				"text": {
					"format": {
						"type": "json_schema",
						"name": "analysis_result",
						"schema": {
							"type": "object",
							"properties": {
								"summary": {"type": "string"},
								"details": {
									"type": "array",
									"items": {"type": "string"}
								},
								"confidence": {
									"type": "number",
									"minimum": 0,
									"maximum": 1
								}
							}
						}
					}
				}
			}`,
			expectError: false,
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

			// Basic validation that the struct was populated correctly for successful cases
			if err == nil {
				// Verify that the request can be marshaled back to JSON
				_, marshalErr := json.Marshal(request)
				if marshalErr != nil {
					t.Errorf("Failed to marshal request back to JSON: %v", marshalErr)
				}
			}
		})
	}
}

// TestResponseAPIInputMarshalUnmarshal tests bidirectional conversion of ResponseAPIInput
func TestResponseAPIInputMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input ResponseAPIInput
	}{
		{
			name:  "Single string",
			input: ResponseAPIInput{"Hello world"},
		},
		{
			name:  "Multiple strings",
			input: ResponseAPIInput{"Hello", "world"},
		},
		{
			name: "Mixed content",
			input: ResponseAPIInput{
				"Hello",
				map[string]any{
					"role":    "user",
					"content": "world",
				},
			},
		},
		{
			name: "Complex message structure",
			input: ResponseAPIInput{
				map[string]any{
					"role": "user",
					"content": []any{
						map[string]any{
							"type": "input_text",
							"text": "What's in this image?",
						},
						map[string]any{
							"type":      "input_image",
							"image_url": "https://example.com/image.jpg",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Failed to marshal input: %v", err)
			}

			// Unmarshal back
			var result ResponseAPIInput
			err = json.Unmarshal(jsonData, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal input: %v", err)
			}

			// Verify length matches
			if len(result) != len(tt.input) {
				t.Errorf("Expected length %d, got %d", len(tt.input), len(result))
			}

			// For single string case, verify it marshals as string not array
			if len(tt.input) == 1 {
				if str, ok := tt.input[0].(string); ok {
					var directString string
					err = json.Unmarshal(jsonData, &directString)
					if err != nil {
						t.Errorf("Single string should marshal as string, not array")
					} else if directString != str {
						t.Errorf("Expected string '%s', got '%s'", str, directString)
					}
				}
			}
		})
	}
}
