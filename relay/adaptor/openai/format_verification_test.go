package openai

import (
	"encoding/json"
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

func TestResponseAPIFormat(t *testing.T) {
	// Create a request similar to the one that caused the error
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "o3",
		Messages: []model.Message{
			{Role: "user", Content: "What is the weather like in Boston?"},
		},
		Stream:      false,
		Temperature: floatPtr(1.0),
		User:        "",
	}

	// Convert to Response API format
	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Marshal to JSON to see the exact format
	jsonData, err := json.Marshal(responseAPI)
	if err != nil {
		t.Fatalf("Failed to marshal ResponseAPIRequest: %v", err)
	}

	t.Logf("Generated Response API request: %s", string(jsonData))

	// Verify the input structure
	if len(responseAPI.Input) != 1 {
		t.Errorf("Expected 1 input item, got %d", len(responseAPI.Input))
	}

	// Verify that input[0] is a direct message, not wrapped
	inputMessage, ok := responseAPI.Input[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected input[0] to be map[string]interface{}, got %T", responseAPI.Input[0])
	}

	// Verify the message has the correct role
	if inputMessage["role"] != "user" {
		t.Errorf("Expected role 'user', got '%v'", inputMessage["role"])
	}

	// Verify the message has the correct content
	expectedContent := "What is the weather like in Boston?"
	if content, ok := inputMessage["content"].([]map[string]interface{}); ok && len(content) > 0 {
		if content[0]["text"] != expectedContent {
			t.Errorf("Expected content '%s', got '%v'", expectedContent, content[0]["text"])
		}
	} else {
		t.Error("Expected content to be []map[string]interface{}")
	}

	// Parse the JSON back to verify it's valid
	var unmarshaled ResponseAPIRequest
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ResponseAPIRequest: %v", err)
	}

	// Verify the unmarshaled data matches expectations
	if len(unmarshaled.Input) != 1 {
		t.Errorf("After unmarshal: Expected 1 input item, got %d", len(unmarshaled.Input))
	}

	// The unmarshaled input will be map[string]interface{} due to JSON unmarshaling
	inputMap, ok := unmarshaled.Input[0].(map[string]interface{})
	if !ok {
		t.Fatalf("After unmarshal: Expected input[0] to be map[string]interface{}, got %T", unmarshaled.Input[0])
	}

	// Verify the role in the map
	if role, exists := inputMap["role"]; !exists || role != "user" {
		t.Errorf("After unmarshal: Expected role 'user', got %v", role)
	}

	// Verify the content in the map (should be array format after unmarshaling)
	if content, exists := inputMap["content"]; !exists {
		t.Error("After unmarshal: Expected content field to exist")
	} else if contentArray, ok := content.([]interface{}); !ok {
		t.Errorf("After unmarshal: Expected content to be []interface{}, got %T", content)
	} else if len(contentArray) != 1 {
		t.Errorf("After unmarshal: Expected content array length 1, got %d", len(contentArray))
	} else if contentItem, ok := contentArray[0].(map[string]interface{}); !ok {
		t.Errorf("After unmarshal: Expected content[0] to be map[string]interface{}, got %T", contentArray[0])
	} else if contentItem["text"] != expectedContent {
		t.Errorf("After unmarshal: Expected content text '%s', got %v", expectedContent, contentItem["text"])
	}
}

func TestResponseAPIWithSystemMessage(t *testing.T) {
	// Test the exact scenario from the error log with a system message
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "o3",
		Messages: []model.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is the weather like in Boston?"},
		},
		Stream:      false,
		Temperature: floatPtr(1.0),
		User:        "",
	}

	// Convert to Response API format
	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Marshal to JSON
	jsonData, err := json.Marshal(responseAPI)
	if err != nil {
		t.Fatalf("Failed to marshal ResponseAPIRequest: %v", err)
	}

	t.Logf("Generated Response API request with system message: %s", string(jsonData))

	// Verify system message is moved to instructions
	if responseAPI.Instructions == nil || *responseAPI.Instructions != "You are a helpful assistant." {
		t.Errorf("Expected instructions to be 'You are a helpful assistant.', got %v", responseAPI.Instructions)
	}

	// Verify only user message remains in input
	if len(responseAPI.Input) != 1 {
		t.Errorf("Expected 1 input item after system message removal, got %d", len(responseAPI.Input))
	}

	inputMessage, ok := responseAPI.Input[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected input[0] to be map[string]interface{}, got %T", responseAPI.Input[0])
	}

	if inputMessage["role"] != "user" {
		t.Errorf("Expected remaining message role 'user', got '%v'", inputMessage["role"])
	}
}
