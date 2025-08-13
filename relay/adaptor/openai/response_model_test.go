package openai

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

func TestConvertChatCompletionToResponseAPI(t *testing.T) {
	// Test basic conversion
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "Hello, world!"},
		},
		MaxTokens:   100,
		Temperature: floatPtr(0.7),
		TopP:        floatPtr(0.9),
		Stream:      true,
		User:        "test-user",
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify basic fields
	if responseAPI.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", responseAPI.Model)
	}

	if *responseAPI.MaxOutputTokens != 100 {
		t.Errorf("Expected max_output_tokens 100, got %d", *responseAPI.MaxOutputTokens)
	}

	if *responseAPI.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", *responseAPI.Temperature)
	}

	if *responseAPI.TopP != 0.9 {
		t.Errorf("Expected top_p 0.9, got %f", *responseAPI.TopP)
	}

	if !*responseAPI.Stream {
		t.Error("Expected stream to be true")
	}

	if *responseAPI.User != "test-user" {
		t.Errorf("Expected user 'test-user', got '%s'", *responseAPI.User)
	}

	// Verify input conversion
	if len(responseAPI.Input) != 1 {
		t.Errorf("Expected 1 input item, got %d", len(responseAPI.Input))
	}

	inputMessage, ok := responseAPI.Input[0].(map[string]interface{})
	if !ok {
		t.Error("Expected input item to be map[string]interface{} type")
	}

	if inputMessage["role"] != "user" {
		t.Errorf("Expected message role 'user', got '%v'", inputMessage["role"])
	}

	// Check content structure
	content, ok := inputMessage["content"].([]map[string]interface{})
	if !ok {
		t.Error("Expected content to be []map[string]interface{}")
	}
	if len(content) != 1 {
		t.Errorf("Expected content length 1, got %d", len(content))
	}
	if content[0]["type"] != "input_text" {
		t.Errorf("Expected content type 'input_text', got '%v'", content[0]["type"])
	}
	if content[0]["text"] != "Hello, world!" {
		t.Errorf("Expected message content 'Hello, world!', got '%v'", content[0]["text"])
	}
}

func TestConvertWithSystemMessage(t *testing.T) {
	// Test system message conversion to instructions
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello!"},
		},
		MaxTokens: 50,
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify system message is converted to instructions
	if responseAPI.Instructions == nil {
		t.Error("Expected instructions to be set")
	} else if *responseAPI.Instructions != "You are a helpful assistant." {
		t.Errorf("Expected instructions 'You are a helpful assistant.', got '%s'", *responseAPI.Instructions)
	}

	// Verify system message is removed from input
	if len(responseAPI.Input) != 1 {
		t.Errorf("Expected 1 input item after system message removal, got %d", len(responseAPI.Input))
	}

	inputMessage, ok := responseAPI.Input[0].(map[string]interface{})
	if !ok {
		t.Error("Expected input item to be map[string]interface{} type")
	}

	if inputMessage["role"] != "user" {
		t.Errorf("Expected remaining message to be user role, got '%v'", inputMessage["role"])
	}
}

func TestConvertWithTools(t *testing.T) {
	// Test tools conversion
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "What's the weather?"},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: &model.Function{
					Name:        "get_weather",
					Description: "Get current weather",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
		},
		ToolChoice: "auto",
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify tools are preserved
	if len(responseAPI.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(responseAPI.Tools))
	}

	if responseAPI.Tools[0].Name != "get_weather" {
		t.Errorf("Expected tool name 'get_weather', got '%s'", responseAPI.Tools[0].Name)
	}

	if responseAPI.ToolChoice != "auto" {
		t.Errorf("Expected tool_choice 'auto', got '%v'", responseAPI.ToolChoice)
	}
}

func TestConvertWithResponseFormat(t *testing.T) {
	// Test response format conversion
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "Generate JSON"},
		},
		ResponseFormat: &model.ResponseFormat{
			Type: "json_object",
			JsonSchema: &model.JSONSchema{
				Name:        "response_schema",
				Description: "Test schema",
				Schema: map[string]interface{}{
					"type": "object",
				},
			},
		},
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify response format conversion
	if responseAPI.Text == nil {
		t.Error("Expected text config to be set")
	}

	if responseAPI.Text.Format == nil {
		t.Error("Expected text format to be set")
	}

	if responseAPI.Text.Format.Type != "json_object" {
		t.Errorf("Expected text format type to be 'json_object', got '%s'", responseAPI.Text.Format.Type)
	}

	if responseAPI.Text.Format.Name != "response_schema" {
		t.Errorf("Expected schema name 'response_schema', got '%s'", responseAPI.Text.Format.Name)
	}

	if responseAPI.Text.Format.Description != "Test schema" {
		t.Errorf("Expected schema description 'Test schema', got '%s'", responseAPI.Text.Format.Description)
	}

	if responseAPI.Text.Format.Schema == nil {
		t.Error("Expected JSON schema to be set")
	}
}

// TestConvertResponseAPIToChatCompletion tests the conversion from Response API format back to ChatCompletion format
func TestConvertResponseAPIToChatCompletion(t *testing.T) {
	// Create a Response API response
	responseAPI := &ResponseAPIResponse{
		Id:        "resp_123",
		Object:    "response",
		CreatedAt: 1234567890,
		Status:    "completed",
		Model:     "gpt-4",
		Output: []OutputItem{
			{
				Type:   "message",
				Id:     "msg_123",
				Status: "completed",
				Role:   "assistant",
				Content: []OutputContent{
					{
						Type: "output_text",
						Text: "Hello! How can I help you today?",
					},
				},
			},
		},
		Usage: &ResponseAPIUsage{
			InputTokens:  10,
			OutputTokens: 8,
			TotalTokens:  18,
		},
	}

	// Convert to ChatCompletion format
	chatCompletion := ConvertResponseAPIToChatCompletion(responseAPI)

	// Verify basic fields
	if chatCompletion.Id != "resp_123" {
		t.Errorf("Expected id 'resp_123', got '%s'", chatCompletion.Id)
	}

	if chatCompletion.Object != "chat.completion" {
		t.Errorf("Expected object 'chat.completion', got '%s'", chatCompletion.Object)
	}

	if chatCompletion.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", chatCompletion.Model)
	}

	if chatCompletion.Created != 1234567890 {
		t.Errorf("Expected created 1234567890, got %d", chatCompletion.Created)
	}

	// Verify choices
	if len(chatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(chatCompletion.Choices))
	}

	choice := chatCompletion.Choices[0]
	if choice.Index != 0 {
		t.Errorf("Expected choice index 0, got %d", choice.Index)
	}

	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	if choice.Message.Reasoning != nil {
		t.Errorf("Expected reasoning to be nil, got '%s'", *choice.Message.Reasoning)
	}

	if choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got '%s'", choice.FinishReason)
	}

	// Verify usage
	if chatCompletion.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt_tokens 10, got %d", chatCompletion.Usage.PromptTokens)
	}

	if chatCompletion.Usage.CompletionTokens != 8 {
		t.Errorf("Expected completion_tokens 8, got %d", chatCompletion.Usage.CompletionTokens)
	}

	if chatCompletion.Usage.TotalTokens != 18 {
		t.Errorf("Expected total_tokens 18, got %d", chatCompletion.Usage.TotalTokens)
	}
}

// TestConvertResponseAPIToChatCompletionWithFunctionCall tests the conversion with function calls
func TestConvertResponseAPIToChatCompletionWithFunctionCall(t *testing.T) {
	// Create a Response API response with function call (based on the real example)
	responseAPI := &ResponseAPIResponse{
		Id:        "resp_67ca09c5efe0819096d0511c92b8c890096610f474011cc0",
		Object:    "response",
		CreatedAt: 1741294021,
		Status:    "completed",
		Model:     "gpt-4.1-2025-04-14",
		Output: []OutputItem{
			{
				Type:      "function_call",
				Id:        "fc_67ca09c6bedc8190a7abfec07b1a1332096610f474011cc0",
				CallId:    "call_unLAR8MvFNptuiZK6K6HCy5k",
				Name:      "get_current_weather",
				Arguments: "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}",
				Status:    "completed",
			},
		},
		Usage: &ResponseAPIUsage{
			InputTokens:  291,
			OutputTokens: 23,
			TotalTokens:  314,
		},
	}

	// Convert to ChatCompletion format
	chatCompletion := ConvertResponseAPIToChatCompletion(responseAPI)

	// Verify basic fields
	if chatCompletion.Id != "resp_67ca09c5efe0819096d0511c92b8c890096610f474011cc0" {
		t.Errorf("Expected id 'resp_67ca09c5efe0819096d0511c92b8c890096610f474011cc0', got '%s'", chatCompletion.Id)
	}

	if chatCompletion.Model != "gpt-4.1-2025-04-14" {
		t.Errorf("Expected model 'gpt-4.1-2025-04-14', got '%s'", chatCompletion.Model)
	}

	// Verify choices
	if len(chatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(chatCompletion.Choices))
	}

	choice := chatCompletion.Choices[0]
	if choice.Index != 0 {
		t.Errorf("Expected choice index 0, got %d", choice.Index)
	}

	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	// Verify tool calls
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Message.ToolCalls))
	}

	toolCall := choice.Message.ToolCalls[0]
	if toolCall.Id != "call_unLAR8MvFNptuiZK6K6HCy5k" {
		t.Errorf("Expected tool call id 'call_unLAR8MvFNptuiZK6K6HCy5k', got '%s'", toolCall.Id)
	}

	if toolCall.Type != "function" {
		t.Errorf("Expected tool call type 'function', got '%s'", toolCall.Type)
	}

	if toolCall.Function.Name != "get_current_weather" {
		t.Errorf("Expected function name 'get_current_weather', got '%s'", toolCall.Function.Name)
	}

	expectedArgs := "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}"
	if toolCall.Function.Arguments != expectedArgs {
		t.Errorf("Expected arguments '%s', got '%s'", expectedArgs, toolCall.Function.Arguments)
	}

	if choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got '%s'", choice.FinishReason)
	}

	// Verify usage
	if chatCompletion.Usage.PromptTokens != 291 {
		t.Errorf("Expected prompt_tokens 291, got %d", chatCompletion.Usage.PromptTokens)
	}

	if chatCompletion.Usage.CompletionTokens != 23 {
		t.Errorf("Expected completion_tokens 23, got %d", chatCompletion.Usage.CompletionTokens)
	}

	if chatCompletion.Usage.TotalTokens != 314 {
		t.Errorf("Expected total_tokens 314, got %d", chatCompletion.Usage.TotalTokens)
	}
}

// TestConvertResponseAPIStreamToChatCompletion tests the conversion from Response API streaming format to ChatCompletion streaming format
func TestConvertResponseAPIStreamToChatCompletion(t *testing.T) {
	// Create a Response API streaming chunk
	responseAPIChunk := &ResponseAPIResponse{
		Id:        "resp_123",
		Object:    "response",
		CreatedAt: 1234567890,
		Status:    "in_progress",
		Model:     "gpt-4",
		Output: []OutputItem{
			{
				Type:   "message",
				Id:     "msg_123",
				Status: "in_progress",
				Role:   "assistant",
				Content: []OutputContent{
					{
						Type: "output_text",
						Text: "Hello",
					},
				},
			},
		},
	}

	// Convert to ChatCompletion streaming format
	streamChunk := ConvertResponseAPIStreamToChatCompletion(responseAPIChunk)

	// Verify basic fields
	if streamChunk.Id != "resp_123" {
		t.Errorf("Expected id 'resp_123', got '%s'", streamChunk.Id)
	}

	if streamChunk.Object != "chat.completion.chunk" {
		t.Errorf("Expected object 'chat.completion.chunk', got '%s'", streamChunk.Object)
	}

	if streamChunk.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", streamChunk.Model)
	}

	if streamChunk.Created != 1234567890 {
		t.Errorf("Expected created 1234567890, got %d", streamChunk.Created)
	}

	// Verify choices
	if len(streamChunk.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(streamChunk.Choices))
	}

	choice := streamChunk.Choices[0]
	if choice.Index != 0 {
		t.Errorf("Expected choice index 0, got %d", choice.Index)
	}

	if choice.Delta.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Delta.Role)
	}

	if choice.Delta.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", choice.Delta.Content)
	}

	// For in_progress status, finish_reason should be nil
	if choice.FinishReason != nil {
		t.Errorf("Expected finish_reason to be nil for in_progress status, got '%s'", *choice.FinishReason)
	}

	// Test completed status
	responseAPIChunk.Status = "completed"
	streamChunk = ConvertResponseAPIStreamToChatCompletion(responseAPIChunk)
	choice = streamChunk.Choices[0]

	if choice.FinishReason == nil || *choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop' for completed status, got %v", choice.FinishReason)
	}
}

// TestConvertResponseAPIStreamToChatCompletionWithFunctionCall tests streaming conversion with function calls
func TestConvertResponseAPIStreamToChatCompletionWithFunctionCall(t *testing.T) {
	// Create a Response API streaming chunk with function call
	responseAPIChunk := &ResponseAPIResponse{
		Id:        "resp_123",
		Object:    "response",
		CreatedAt: 1234567890,
		Status:    "completed",
		Model:     "gpt-4",
		Output: []OutputItem{
			{
				Type:      "function_call",
				Id:        "fc_123",
				CallId:    "call_456",
				Name:      "get_weather",
				Arguments: "{\"location\":\"Boston\"}",
				Status:    "completed",
			},
		},
	}

	// Convert to ChatCompletion streaming format
	streamChunk := ConvertResponseAPIStreamToChatCompletion(responseAPIChunk)

	// Verify basic fields
	if streamChunk.Id != "resp_123" {
		t.Errorf("Expected id 'resp_123', got '%s'", streamChunk.Id)
	}

	if streamChunk.Object != "chat.completion.chunk" {
		t.Errorf("Expected object 'chat.completion.chunk', got '%s'", streamChunk.Object)
	}

	// Verify choices
	if len(streamChunk.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(streamChunk.Choices))
	}

	choice := streamChunk.Choices[0]
	if choice.Index != 0 {
		t.Errorf("Expected choice index 0, got %d", choice.Index)
	}

	if choice.Delta.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Delta.Role)
	}

	// Verify tool calls
	if len(choice.Delta.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Delta.ToolCalls))
	}

	toolCall := choice.Delta.ToolCalls[0]
	if toolCall.Id != "call_456" {
		t.Errorf("Expected tool call id 'call_456', got '%s'", toolCall.Id)
	}

	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got '%s'", toolCall.Function.Name)
	}

	if toolCall.Function.Arguments != "{\"location\":\"Boston\"}" {
		t.Errorf("Expected arguments '{\"location\":\"Boston\"}', got '%s'", toolCall.Function.Arguments)
	}

	// For completed status, finish_reason should be "stop"
	if choice.FinishReason == nil || *choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop' for completed status, got %v", choice.FinishReason)
	}
}

// TestConvertResponseAPIToChatCompletionWithReasoning tests the conversion with reasoning content
func TestConvertResponseAPIToChatCompletionWithReasoning(t *testing.T) {
	// Create a Response API response with reasoning content (based on the real example)
	responseAPI := &ResponseAPIResponse{
		Id:        "resp_6848f7a7ac94819cba6af50194a156e7050d57f0136932b5",
		Object:    "response",
		CreatedAt: 1749612455,
		Status:    "completed",
		Model:     "o3-2025-04-16",
		Output: []OutputItem{
			{
				Id:   "rs_6848f7a7f800819ca52a87ae9a6a59ef050d57f0136932b5",
				Type: "reasoning",
				Summary: []OutputContent{
					{
						Type: "summary_text",
						Text: "**Telling a joke**\n\nThe user asked for a joke, which is a straightforward request. There's no conflict with the guidelines, so I can definitely comply.",
					},
				},
			},
			{
				Id:     "msg_6848f7abc86c819c877542f4a72a3f1d050d57f0136932b5",
				Type:   "message",
				Status: "completed",
				Role:   "assistant",
				Content: []OutputContent{
					{
						Type: "output_text",
						Text: "Why don't scientists trust atoms?\n\nBecause they make up everything!",
					},
				},
			},
		},
		Usage: &ResponseAPIUsage{
			InputTokens:  9,
			OutputTokens: 83,
			TotalTokens:  92,
		},
	}

	// Convert to ChatCompletion format
	chatCompletion := ConvertResponseAPIToChatCompletion(responseAPI)

	// Verify basic fields
	if chatCompletion.Id != "resp_6848f7a7ac94819cba6af50194a156e7050d57f0136932b5" {
		t.Errorf("Expected id 'resp_6848f7a7ac94819cba6af50194a156e7050d57f0136932b5', got '%s'", chatCompletion.Id)
	}

	if chatCompletion.Model != "o3-2025-04-16" {
		t.Errorf("Expected model 'o3-2025-04-16', got '%s'", chatCompletion.Model)
	}

	// Verify choices
	if len(chatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(chatCompletion.Choices))
	}

	choice := chatCompletion.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	expectedContent := "Why don't scientists trust atoms?\n\nBecause they make up everything!"
	if choice.Message.Content != expectedContent {
		t.Errorf("Expected content '%s', got '%s'", expectedContent, choice.Message.Content)
	}

	// Verify reasoning content is properly extracted
	if choice.Message.Reasoning == nil {
		t.Fatal("Expected reasoning content to be present, got nil")
	}

	expectedReasoning := "**Telling a joke**\n\nThe user asked for a joke, which is a straightforward request. There's no conflict with the guidelines, so I can definitely comply."
	if *choice.Message.Reasoning != expectedReasoning {
		t.Errorf("Expected reasoning '%s', got '%s'", expectedReasoning, *choice.Message.Reasoning)
	}

	if choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got '%s'", choice.FinishReason)
	}

	// Verify usage
	if chatCompletion.Usage.PromptTokens != 9 {
		t.Errorf("Expected prompt_tokens 9, got %d", chatCompletion.Usage.PromptTokens)
	}

	if chatCompletion.Usage.CompletionTokens != 83 {
		t.Errorf("Expected completion_tokens 83, got %d", chatCompletion.Usage.CompletionTokens)
	}

	if chatCompletion.Usage.TotalTokens != 92 {
		t.Errorf("Expected total_tokens 92, got %d", chatCompletion.Usage.TotalTokens)
	}
}

// TestFunctionCallWorkflow tests the complete function calling workflow:
// ChatCompletion -> ResponseAPI -> ResponseAPI Response -> ChatCompletion
func TestFunctionCallWorkflow(t *testing.T) {
	// Step 1: Create original ChatCompletion request with tools
	originalRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "What's the weather like in Boston today?"},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: &model.Function{
					Name:        "get_current_weather",
					Description: "Get the current weather in a given location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state, e.g. San Francisco, CA",
							},
							"unit": map[string]interface{}{
								"type": "string",
								"enum": []string{"celsius", "fahrenheit"},
							},
						},
						"required": []string{"location", "unit"},
					},
				},
			},
		},
		ToolChoice: "auto",
	}

	// Step 2: Convert ChatCompletion to Response API format
	responseAPIRequest := ConvertChatCompletionToResponseAPI(originalRequest)

	// Verify tools are preserved in request
	if len(responseAPIRequest.Tools) != 1 {
		t.Fatalf("Expected 1 tool in request, got %d", len(responseAPIRequest.Tools))
	}

	if responseAPIRequest.Tools[0].Name != "get_current_weather" {
		t.Errorf("Expected tool name 'get_current_weather', got '%s'", responseAPIRequest.Tools[0].Name)
	}

	if responseAPIRequest.ToolChoice != "auto" {
		t.Errorf("Expected tool_choice 'auto', got '%v'", responseAPIRequest.ToolChoice)
	}

	// Step 3: Create a Response API response with function call (simulates upstream response)
	responseAPIResponse := &ResponseAPIResponse{
		Id:        "resp_67ca09c5efe0819096d0511c92b8c890096610f474011cc0",
		Object:    "response",
		CreatedAt: 1741294021,
		Status:    "completed",
		Model:     "gpt-4.1-2025-04-14",
		Output: []OutputItem{
			{
				Type:      "function_call",
				Id:        "fc_67ca09c6bedc8190a7abfec07b1a1332096610f474011cc0",
				CallId:    "call_unLAR8MvFNptuiZK6K6HCy5k",
				Name:      "get_current_weather",
				Arguments: "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}",
				Status:    "completed",
			},
		},
		Usage: &ResponseAPIUsage{
			InputTokens:  291,
			OutputTokens: 23,
			TotalTokens:  314,
		},
	}

	// Step 4: Convert Response API response back to ChatCompletion format
	finalChatCompletion := ConvertResponseAPIToChatCompletion(responseAPIResponse)

	// Step 5: Verify the final ChatCompletion response preserves all function call information
	if len(finalChatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(finalChatCompletion.Choices))
	}

	choice := finalChatCompletion.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	// Verify tool calls are preserved
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Message.ToolCalls))
	}

	toolCall := choice.Message.ToolCalls[0]
	if toolCall.Id != "call_unLAR8MvFNptuiZK6K6HCy5k" {
		t.Errorf("Expected tool call id 'call_unLAR8MvFNptuiZK6K6HCy5k', got '%s'", toolCall.Id)
	}

	if toolCall.Type != "function" {
		t.Errorf("Expected tool call type 'function', got '%s'", toolCall.Type)
	}

	if toolCall.Function.Name != "get_current_weather" {
		t.Errorf("Expected function name 'get_current_weather', got '%s'", toolCall.Function.Name)
	}

	expectedArgs := "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}"
	if toolCall.Function.Arguments != expectedArgs {
		t.Errorf("Expected arguments '%s', got '%s'", expectedArgs, toolCall.Function.Arguments)
	}

	// Verify usage is preserved
	if finalChatCompletion.Usage.PromptTokens != 291 {
		t.Errorf("Expected prompt_tokens 291, got %d", finalChatCompletion.Usage.PromptTokens)
	}

	if finalChatCompletion.Usage.CompletionTokens != 23 {
		t.Errorf("Expected completion_tokens 23, got %d", finalChatCompletion.Usage.CompletionTokens)
	}

	if finalChatCompletion.Usage.TotalTokens != 314 {
		t.Errorf("Expected total_tokens 314, got %d", finalChatCompletion.Usage.TotalTokens)
	}

	t.Log("Function call workflow test completed successfully!")
	t.Logf("Original request tools: %d", len(originalRequest.Tools))
	t.Logf("Response API request tools: %d", len(responseAPIRequest.Tools))
	t.Logf("Final response tool calls: %d", len(choice.Message.ToolCalls))
}

func TestConvertWithLegacyFunctions(t *testing.T) {
	// Test legacy functions conversion
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "What's the weather?"},
		},
		Functions: []model.Function{
			{
				Name:        "get_current_weather",
				Description: "Get current weather",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]interface{}{
							"type": "string",
							"enum": []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"location"},
				},
			},
		},
		FunctionCall: "auto",
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify functions are converted to tools
	if len(responseAPI.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(responseAPI.Tools))
	}

	if responseAPI.Tools[0].Type != "function" {
		t.Errorf("Expected tool type 'function', got '%s'", responseAPI.Tools[0].Type)
	}

	if responseAPI.Tools[0].Name != "get_current_weather" {
		t.Errorf("Expected function name 'get_current_weather', got '%s'", responseAPI.Tools[0].Name)
	}

	if responseAPI.ToolChoice != "auto" {
		t.Errorf("Expected tool_choice 'auto', got '%v'", responseAPI.ToolChoice)
	}

	// Verify the function parameters are preserved
	if responseAPI.Tools[0].Parameters == nil {
		t.Error("Expected function parameters to be preserved")
	}

	// Verify properties are preserved
	if props, ok := responseAPI.Tools[0].Parameters["properties"].(map[string]interface{}); ok {
		if location, ok := props["location"].(map[string]interface{}); ok {
			if location["type"] != "string" {
				t.Errorf("Expected location type 'string', got '%v'", location["type"])
			}
		} else {
			t.Error("Expected location property to be preserved")
		}
	} else {
		t.Error("Expected properties to be preserved")
	}
}

// TestLegacyFunctionCallWorkflow tests the complete legacy function calling workflow:
// ChatCompletion with Functions -> ResponseAPI -> ResponseAPI Response -> ChatCompletion
func TestLegacyFunctionCallWorkflow(t *testing.T) {
	// Step 1: Create original ChatCompletion request with legacy functions
	originalRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "user", Content: "What's the weather like in Boston today?"},
		},
		Functions: []model.Function{
			{
				Name:        "get_current_weather",
				Description: "Get the current weather in a given location",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
						"unit": map[string]interface{}{
							"type": "string",
							"enum": []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"location", "unit"},
				},
			},
		},
		FunctionCall: "auto",
	}

	// Step 2: Convert ChatCompletion to Response API format
	responseAPIRequest := ConvertChatCompletionToResponseAPI(originalRequest)

	// Verify functions are converted to tools in request
	if len(responseAPIRequest.Tools) != 1 {
		t.Fatalf("Expected 1 tool in request, got %d", len(responseAPIRequest.Tools))
	}

	if responseAPIRequest.Tools[0].Name != "get_current_weather" {
		t.Errorf("Expected tool name 'get_current_weather', got '%s'", responseAPIRequest.Tools[0].Name)
	}

	if responseAPIRequest.ToolChoice != "auto" {
		t.Errorf("Expected tool_choice 'auto', got '%v'", responseAPIRequest.ToolChoice)
	}

	// Step 3: Create mock Response API response (simulating what the API would return)
	responseAPIResponse := &ResponseAPIResponse{
		Id:        "resp_legacy_test",
		Object:    "response",
		CreatedAt: 1741294021,
		Status:    "completed",
		Model:     "gpt-4.1-2025-04-14",
		Output: []OutputItem{
			{
				Type:      "function_call",
				Id:        "fc_legacy_test",
				CallId:    "call_legacy_test_123",
				Name:      "get_current_weather",
				Arguments: "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}",
				Status:    "completed",
			},
		},
		ParallelToolCalls: true,
		ToolChoice:        "auto",
		Tools: []model.Tool{
			{
				Type: "function",
				Function: &model.Function{
					Name:        "get_current_weather",
					Description: "Get the current weather in a given location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type":        "string",
								"description": "The city and state, e.g. San Francisco, CA",
							},
							"unit": map[string]interface{}{
								"type": "string",
								"enum": []string{"celsius", "fahrenheit"},
							},
						},
						"required": []string{"location", "unit"},
					},
				},
			},
		},
		Usage: &ResponseAPIUsage{
			InputTokens:  291,
			OutputTokens: 23,
			TotalTokens:  314,
		},
	}

	// Step 4: Convert Response API response back to ChatCompletion format
	finalChatCompletion := ConvertResponseAPIToChatCompletion(responseAPIResponse)

	// Step 5: Verify the final ChatCompletion response preserves all function call information
	if len(finalChatCompletion.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(finalChatCompletion.Choices))
	}

	choice := finalChatCompletion.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
	}

	// Verify tool calls are preserved
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Message.ToolCalls))
	}

	toolCall := choice.Message.ToolCalls[0]
	if toolCall.Id != "call_legacy_test_123" {
		t.Errorf("Expected tool call id 'call_legacy_test_123', got '%s'", toolCall.Id)
	}

	if toolCall.Type != "function" {
		t.Errorf("Expected tool call type 'function', got '%s'", toolCall.Type)
	}

	if toolCall.Function.Name != "get_current_weather" {
		t.Errorf("Expected function name 'get_current_weather', got '%s'", toolCall.Function.Name)
	}

	expectedArgs := "{\"location\":\"Boston, MA\",\"unit\":\"celsius\"}"
	if toolCall.Function.Arguments != expectedArgs {
		t.Errorf("Expected arguments '%s', got '%s'", expectedArgs, toolCall.Function.Arguments)
	}

	// Verify usage is preserved
	if finalChatCompletion.Usage.PromptTokens != 291 {
		t.Errorf("Expected prompt_tokens 291, got %d", finalChatCompletion.Usage.PromptTokens)
	}

	if finalChatCompletion.Usage.CompletionTokens != 23 {
		t.Errorf("Expected completion_tokens 23, got %d", finalChatCompletion.Usage.CompletionTokens)
	}

	if finalChatCompletion.Usage.TotalTokens != 314 {
		t.Errorf("Expected total_tokens 314, got %d", finalChatCompletion.Usage.TotalTokens)
	}

	t.Log("Legacy function call workflow test completed successfully!")
	t.Logf("Original request functions: %d", len(originalRequest.Functions))
	t.Logf("Response API request tools: %d", len(responseAPIRequest.Tools))
	t.Logf("Final response tool calls: %d", len(choice.Message.ToolCalls))
}

// TestParseResponseAPIStreamEvent tests the flexible parsing of Response API streaming events
func TestParseResponseAPIStreamEvent(t *testing.T) {
	t.Run("Parse response.output_text.done event", func(t *testing.T) {
		// This is the problematic event that was causing parsing failures
		eventData := `{"type":"response.output_text.done","sequence_number":22,"item_id":"msg_6849865110908191a4809c86e082ff710008bd3c6060334b","output_index":1,"content_index":0,"text":"Why don't skeletons fight each other?\n\nThey don't have the guts."}`

		fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(eventData))
		if err != nil {
			t.Fatalf("Failed to parse streaming event: %v", err)
		}

		// Should parse as streaming event, not full response
		if fullResponse != nil {
			t.Error("Expected fullResponse to be nil for streaming event")
		}

		if streamEvent == nil {
			t.Fatal("Expected streamEvent to be non-nil")
		}

		// Verify event fields
		if streamEvent.Type != "response.output_text.done" {
			t.Errorf("Expected type 'response.output_text.done', got '%s'", streamEvent.Type)
		}

		if streamEvent.SequenceNumber != 22 {
			t.Errorf("Expected sequence_number 22, got %d", streamEvent.SequenceNumber)
		}

		if streamEvent.ItemId != "msg_6849865110908191a4809c86e082ff710008bd3c6060334b" {
			t.Errorf("Expected item_id 'msg_6849865110908191a4809c86e082ff710008bd3c6060334b', got '%s'", streamEvent.ItemId)
		}

		expectedText := "Why don't skeletons fight each other?\n\nThey don't have the guts."
		if streamEvent.Text != expectedText {
			t.Errorf("Expected text '%s', got '%s'", expectedText, streamEvent.Text)
		}
	})

	t.Run("Parse response.output_text.delta event", func(t *testing.T) {
		eventData := `{"type":"response.output_text.delta","sequence_number":6,"item_id":"msg_6849865110908191a4809c86e082ff710008bd3c6060334b","output_index":1,"content_index":0,"delta":"Why"}`

		_, streamEvent, err := ParseResponseAPIStreamEvent([]byte(eventData))
		if err != nil {
			t.Fatalf("Failed to parse delta event: %v", err)
		}

		if streamEvent == nil {
			t.Fatal("Expected streamEvent to be non-nil")
		}

		// Verify event fields
		if streamEvent.Type != "response.output_text.delta" {
			t.Errorf("Expected type 'response.output_text.delta', got '%s'", streamEvent.Type)
		}

		if streamEvent.Delta != "Why" {
			t.Errorf("Expected delta 'Why', got '%s'", streamEvent.Delta)
		}
	})

	t.Run("Parse full response event", func(t *testing.T) {
		eventData := `{"id":"resp_123","object":"response","created_at":1749648976,"status":"completed","model":"o3-2025-04-16","output":[{"type":"message","id":"msg_123","status":"completed","role":"assistant","content":[{"type":"output_text","text":"Hello world"}]}],"usage":{"input_tokens":9,"output_tokens":22,"total_tokens":31}}`

		fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(eventData))
		if err != nil {
			t.Fatalf("Failed to parse full response event: %v", err)
		}

		// Should parse as full response, not streaming event
		if streamEvent != nil {
			t.Error("Expected streamEvent to be nil for full response")
		}

		if fullResponse == nil {
			t.Fatal("Expected fullResponse to be non-nil")
		}

		// Verify response fields
		if fullResponse.Id != "resp_123" {
			t.Errorf("Expected id 'resp_123', got '%s'", fullResponse.Id)
		}

		if fullResponse.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", fullResponse.Status)
		}

		if fullResponse.Usage == nil || fullResponse.Usage.TotalTokens != 31 {
			t.Errorf("Expected total_tokens 31, got %v", fullResponse.Usage)
		}
	})

	t.Run("Parse invalid JSON", func(t *testing.T) {
		eventData := `{"invalid": json}`

		_, _, err := ParseResponseAPIStreamEvent([]byte(eventData))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

// TestConvertStreamEventToResponse tests the conversion of streaming events to ResponseAPIResponse format
func TestConvertStreamEventToResponse(t *testing.T) {
	t.Run("Convert response.output_text.done event", func(t *testing.T) {
		streamEvent := &ResponseAPIStreamEvent{
			Type:           "response.output_text.done",
			SequenceNumber: 22,
			ItemId:         "msg_123",
			OutputIndex:    1,
			ContentIndex:   0,
			Text:           "Hello, world!",
		}

		response := ConvertStreamEventToResponse(streamEvent)

		// Verify basic fields
		if response.Object != "response" {
			t.Errorf("Expected object 'response', got '%s'", response.Object)
		}

		if response.Status != "in_progress" {
			t.Errorf("Expected status 'in_progress', got '%s'", response.Status)
		}

		// Verify output
		if len(response.Output) != 1 {
			t.Fatalf("Expected 1 output item, got %d", len(response.Output))
		}

		output := response.Output[0]
		if output.Type != "message" {
			t.Errorf("Expected output type 'message', got '%s'", output.Type)
		}

		if output.Role != "assistant" {
			t.Errorf("Expected output role 'assistant', got '%s'", output.Role)
		}

		if len(output.Content) != 1 {
			t.Fatalf("Expected 1 content item, got %d", len(output.Content))
		}

		content := output.Content[0]
		if content.Type != "output_text" {
			t.Errorf("Expected content type 'output_text', got '%s'", content.Type)
		}

		if content.Text != "Hello, world!" {
			t.Errorf("Expected content text 'Hello, world!', got '%s'", content.Text)
		}
	})

	t.Run("Convert response.output_text.delta event", func(t *testing.T) {
		streamEvent := &ResponseAPIStreamEvent{
			Type:           "response.output_text.delta",
			SequenceNumber: 6,
			ItemId:         "msg_123",
			OutputIndex:    1,
			ContentIndex:   0,
			Delta:          "Hello",
		}

		response := ConvertStreamEventToResponse(streamEvent)

		// Verify basic fields
		if response.Object != "response" {
			t.Errorf("Expected object 'response', got '%s'", response.Object)
		}

		if response.Status != "in_progress" {
			t.Errorf("Expected status 'in_progress', got '%s'", response.Status)
		}

		// Verify output
		if len(response.Output) != 1 {
			t.Fatalf("Expected 1 output item, got %d", len(response.Output))
		}

		output := response.Output[0]
		if output.Type != "message" {
			t.Errorf("Expected output type 'message', got '%s'", output.Type)
		}

		if output.Role != "assistant" {
			t.Errorf("Expected output role 'assistant', got '%s'", output.Role)
		}

		if len(output.Content) != 1 {
			t.Fatalf("Expected 1 content item, got %d", len(output.Content))
		}

		content := output.Content[0]
		if content.Type != "output_text" {
			t.Errorf("Expected content type 'output_text', got '%s'", content.Type)
		}

		if content.Text != "Hello" {
			t.Errorf("Expected content text 'Hello', got '%s'", content.Text)
		}
	})

	t.Run("Convert unknown event type", func(t *testing.T) {
		streamEvent := &ResponseAPIStreamEvent{
			Type:           "response.unknown.event",
			SequenceNumber: 1,
			ItemId:         "msg_123",
		}

		response := ConvertStreamEventToResponse(streamEvent)

		// Should still create a basic response structure
		if response.Object != "response" {
			t.Errorf("Expected object 'response', got '%s'", response.Object)
		}

		if response.Status != "in_progress" {
			t.Errorf("Expected status 'in_progress', got '%s'", response.Status)
		}

		// Output should be empty for unknown event types
		if len(response.Output) != 0 {
			t.Errorf("Expected 0 output items for unknown event, got %d", len(response.Output))
		}
	})
}

// TestStreamEventIntegration tests the complete integration of streaming event parsing with ChatCompletion conversion
func TestStreamEventIntegration(t *testing.T) {
	t.Run("End-to-end streaming event processing", func(t *testing.T) {
		// Test the problematic event that was causing the original bug
		eventData := `{"type":"response.output_text.done","sequence_number":22,"item_id":"msg_6849865110908191a4809c86e082ff710008bd3c6060334b","output_index":1,"content_index":0,"text":"Why don't skeletons fight each other?\n\nThey don't have the guts."}`

		// Step 1: Parse the streaming event
		_, streamEvent, err := ParseResponseAPIStreamEvent([]byte(eventData))
		if err != nil {
			t.Fatalf("Failed to parse streaming event: %v", err)
		}

		if streamEvent == nil {
			t.Fatal("Expected streamEvent to be non-nil")
		}

		// Step 2: Convert to ResponseAPIResponse format
		responseAPIChunk := ConvertStreamEventToResponse(streamEvent)

		// Step 3: Convert to ChatCompletion streaming format
		chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)

		// Verify the final result
		if len(chatCompletionChunk.Choices) != 1 {
			t.Fatalf("Expected 1 choice, got %d", len(chatCompletionChunk.Choices))
		}

		choice := chatCompletionChunk.Choices[0]
		expectedContent := "Why don't skeletons fight each other?\n\nThey don't have the guts."
		if content, ok := choice.Delta.Content.(string); !ok || content != expectedContent {
			t.Errorf("Expected delta content '%s', got '%v'", expectedContent, choice.Delta.Content)
		}
	})

	t.Run("Delta event processing", func(t *testing.T) {
		eventData := `{"type":"response.output_text.delta","sequence_number":6,"item_id":"msg_6849865110908191a4809c86e082ff710008bd3c6060334b","output_index":1,"content_index":0,"delta":"Why"}`

		// Step 1: Parse the streaming event
		_, streamEvent, err := ParseResponseAPIStreamEvent([]byte(eventData))
		if err != nil {
			t.Fatalf("Failed to parse delta event: %v", err)
		}

		if streamEvent == nil {
			t.Fatal("Expected streamEvent to be non-nil")
		}

		// Step 2: Convert to ResponseAPIResponse format
		responseAPIChunk := ConvertStreamEventToResponse(streamEvent)

		// Step 3: Convert to ChatCompletion streaming format
		chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)

		// Verify the final result
		if len(chatCompletionChunk.Choices) != 1 {
			t.Fatalf("Expected 1 choice, got %d", len(chatCompletionChunk.Choices))
		}

		choice := chatCompletionChunk.Choices[0]
		if content, ok := choice.Delta.Content.(string); !ok || content != "Why" {
			t.Errorf("Expected delta content 'Why', got '%v'", choice.Delta.Content)
		}
	})
}

// TestConvertChatCompletionToResponseAPIWithToolResults tests that tool result messages
// are properly converted to function_call_output format for Response API
func TestContentTypeBasedOnRole(t *testing.T) {
	// Test that user messages use "input_text" and assistant messages use "output_text"
	userMessage := model.Message{
		Role:    "user",
		Content: "Hello, how are you?",
	}

	assistantMessage := model.Message{
		Role:    "assistant",
		Content: "I'm doing well, thank you!",
	}

	// Convert user message
	userResult := convertMessageToResponseAPIFormat(userMessage)
	userContent := userResult["content"].([]map[string]interface{})
	if userContent[0]["type"] != "input_text" {
		t.Errorf("Expected user message to use 'input_text' type, got '%s'", userContent[0]["type"])
	}
	if userContent[0]["text"] != "Hello, how are you?" {
		t.Errorf("Expected user message text to be preserved, got '%s'", userContent[0]["text"])
	}

	// Convert assistant message
	assistantResult := convertMessageToResponseAPIFormat(assistantMessage)
	assistantContent := assistantResult["content"].([]map[string]interface{})
	if assistantContent[0]["type"] != "output_text" {
		t.Errorf("Expected assistant message to use 'output_text' type, got '%s'", assistantContent[0]["type"])
	}
	if assistantContent[0]["text"] != "I'm doing well, thank you!" {
		t.Errorf("Expected assistant message text to be preserved, got '%s'", assistantContent[0]["text"])
	}
}

func TestConversationWithMultipleRoles(t *testing.T) {
	// Test a conversation similar to the error log scenario
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []model.Message{
			{Role: "system", Content: "you are an experienced english translator"},
			{Role: "user", Content: "我认为后端为 invalid 返回 200 是很荒谬的"},
			{Role: "assistant", Content: "I think it's absurd for the backend to return 200 for an invalid response"},
			{Role: "user", Content: "用户发送的 openai 请求，应该被转换为 ResponseAPI"},
			{Role: "assistant", Content: "The OpenAI request sent by the user should be converted into a ResponseAPI"},
			{Role: "user", Content: "halo"},
		},
		MaxTokens:   5000,
		Temperature: floatPtr(1.0),
		Stream:      true,
	}

	// Convert to Response API format
	responseAPIRequest := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify the conversion
	if responseAPIRequest.Model != "gpt-4o-mini" {
		t.Errorf("Expected model to be 'gpt-4o-mini', got '%s'", responseAPIRequest.Model)
	}

	// Check that input array has correct content types
	inputArray := []any(responseAPIRequest.Input)
	for i, item := range inputArray {
		if itemMap, ok := item.(map[string]interface{}); ok {
			role := itemMap["role"].(string)
			content := itemMap["content"].([]map[string]interface{})

			expectedType := "input_text"
			if role == "assistant" {
				expectedType = "output_text"
			}

			if content[0]["type"] != expectedType {
				t.Errorf("Message %d with role '%s' should use '%s' type, got '%s'",
					i, role, expectedType, content[0]["type"])
			}
		}
	}
}

func TestConvertChatCompletionToResponseAPIWithToolResults(t *testing.T) {
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4o",
		Messages: []model.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What's the current time?"},
			{
				Role:    "assistant",
				Content: "",
				ToolCalls: []model.Tool{
					{
						Id:   "initial_datetime_call",
						Type: "function",
						Function: &model.Function{
							Name:      "get_current_datetime",
							Arguments: `{}`,
						},
					},
				},
			},
			{
				Role:       "tool",
				ToolCallId: "initial_datetime_call",
				Content:    `{"year":2025,"month":6,"day":12,"hour":11,"minute":43,"second":7}`,
			},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: &model.Function{
					Name:        "get_current_datetime",
					Description: "Get current date and time",
					Parameters: map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
				},
			},
		},
	}

	responseAPI := ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify system message was moved to instructions
	if responseAPI.Instructions == nil || *responseAPI.Instructions != "You are a helpful assistant." {
		t.Errorf("Expected system message to be moved to instructions, got %v", responseAPI.Instructions)
	}

	// Verify input array structure
	expectedInputs := 2 // user message, assistant summary
	if len(responseAPI.Input) != expectedInputs {
		t.Fatalf("Expected %d inputs, got %d", expectedInputs, len(responseAPI.Input))
	}

	// Verify first message (user)
	if msg, ok := responseAPI.Input[0].(map[string]interface{}); !ok || msg["role"] != "user" {
		t.Errorf("Expected first input to be user message, got %v", responseAPI.Input[0])
	} else {
		// Check content structure
		if content, ok := msg["content"].([]map[string]interface{}); ok && len(content) > 0 {
			if content[0]["text"] != "What's the current time?" {
				t.Errorf("Expected user message content 'What's the current time?', got '%v'", content[0]["text"])
			}
		}
	}

	// Verify second message (assistant summary of function calls)
	if msg, ok := responseAPI.Input[1].(map[string]interface{}); !ok || msg["role"] != "assistant" {
		t.Fatalf("Expected second input to be assistant summary message, got %T", responseAPI.Input[1])
	} else {
		// Check content structure for function call summary
		if content, ok := msg["content"].([]map[string]interface{}); ok && len(content) > 0 {
			if textContent, ok := content[0]["text"].(string); ok {
				if !strings.Contains(textContent, "Previous function calls") {
					t.Errorf("Expected assistant message to contain function call summary, got '%s'", textContent)
				}
				if !strings.Contains(textContent, "get_current_datetime") {
					t.Errorf("Expected assistant message to mention get_current_datetime, got '%s'", textContent)
				}
				if !strings.Contains(textContent, "year\":2025") {
					t.Errorf("Expected assistant message to contain function result, got '%s'", textContent)
				}
			}
		}
	}

	// Verify tools were converted properly
	if len(responseAPI.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(responseAPI.Tools))
	}

	tool := responseAPI.Tools[0]
	if tool.Name != "get_current_datetime" {
		t.Errorf("Expected tool name 'get_current_datetime', got '%s'", tool.Name)
	}

	if tool.Type != "function" {
		t.Errorf("Expected tool type 'function', got '%s'", tool.Type)
	}
}

// TestStreamingToolCallsIndexField tests that the Index field is properly set in streaming tool calls
func TestStreamingToolCallsIndexField(t *testing.T) {
	// Create a Response API streaming chunk with function call
	responseAPIChunk := &ResponseAPIResponse{
		Id:        "resp_123",
		Object:    "response",
		CreatedAt: 1234567890,
		Status:    "in_progress",
		Output: []OutputItem{
			{
				Type:      "function_call",
				CallId:    "call_abc123",
				Name:      "get_weather",
				Arguments: `{"location": "Paris"}`,
			},
		},
	}

	// Convert to ChatCompletion streaming format
	chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(responseAPIChunk)

	// Verify the response structure
	if len(chatCompletionChunk.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(chatCompletionChunk.Choices))
	}

	choice := chatCompletionChunk.Choices[0]
	if len(choice.Delta.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Delta.ToolCalls))
	}

	toolCall := choice.Delta.ToolCalls[0]

	// Verify that the Index field is set
	if toolCall.Index == nil {
		t.Error("Index field should be set for streaming tool calls")
	} else if *toolCall.Index != 0 {
		t.Errorf("Expected index to be 0, got %d", *toolCall.Index)
	}

	// Verify other tool call fields
	if toolCall.Id != "call_abc123" {
		t.Errorf("Expected tool call id 'call_abc123', got '%s'", toolCall.Id)
	}

	if toolCall.Type != "function" {
		t.Errorf("Expected tool call type 'function', got '%s'", toolCall.Type)
	}

	if toolCall.Function.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got '%s'", toolCall.Function.Name)
	}

	expectedArgs := `{"location": "Paris"}`
	if toolCall.Function.Arguments != expectedArgs {
		t.Errorf("Expected arguments '%s', got '%s'", expectedArgs, toolCall.Function.Arguments)
	}
}

// TestStreamingToolCallsWithOutputIndex tests that the Index field is properly set using output_index from streaming events
func TestStreamingToolCallsWithOutputIndex(t *testing.T) {
	// Test with explicit output_index from streaming event
	responseAPIChunk := &ResponseAPIResponse{
		Id:        "resp_456",
		Object:    "response",
		CreatedAt: 1234567890,
		Status:    "in_progress",
		Output: []OutputItem{
			{
				Type:      "function_call",
				CallId:    "call_def456",
				Name:      "send_email",
				Arguments: `{"to": "test@example.com"}`,
			},
		},
	}

	// Simulate output_index = 2 from a streaming event (e.g., this is the 3rd tool call)
	outputIndex := 2
	chatCompletionChunk := ConvertResponseAPIStreamToChatCompletionWithIndex(responseAPIChunk, &outputIndex)

	// Verify the response structure
	if len(chatCompletionChunk.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(chatCompletionChunk.Choices))
	}

	choice := chatCompletionChunk.Choices[0]
	if len(choice.Delta.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(choice.Delta.ToolCalls))
	}

	toolCall := choice.Delta.ToolCalls[0]

	// Verify that the Index field is set to the provided output_index
	if toolCall.Index == nil {
		t.Error("Index field should be set for streaming tool calls")
	} else if *toolCall.Index != 2 {
		t.Errorf("Expected index to be 2 (from output_index), got %d", *toolCall.Index)
	}

	// Verify other tool call fields
	if toolCall.Id != "call_def456" {
		t.Errorf("Expected tool call id 'call_def456', got '%s'", toolCall.Id)
	}

	if toolCall.Type != "function" {
		t.Errorf("Expected tool call type 'function', got '%s'", toolCall.Type)
	}

	if toolCall.Function.Name != "send_email" {
		t.Errorf("Expected function name 'send_email', got '%s'", toolCall.Function.Name)
	}
}

// TestMultipleStreamingToolCallsIndexConsistency tests that multiple tool calls get consistent indices
func TestMultipleStreamingToolCallsIndexConsistency(t *testing.T) {
	// Test multiple tool calls with different output_index values
	testCases := []struct {
		name        string
		outputIndex *int
		expectedIdx int
	}{
		{
			name:        "First tool call with output_index 0",
			outputIndex: func() *int { i := 0; return &i }(),
			expectedIdx: 0,
		},
		{
			name:        "Second tool call with output_index 1",
			outputIndex: func() *int { i := 1; return &i }(),
			expectedIdx: 1,
		},
		{
			name:        "Third tool call with output_index 2",
			outputIndex: func() *int { i := 2; return &i }(),
			expectedIdx: 2,
		},
		{
			name:        "Tool call without output_index (fallback to position)",
			outputIndex: nil,
			expectedIdx: 0, // Should fallback to position in slice (0)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			responseAPIChunk := &ResponseAPIResponse{
				Id:        "resp_multi",
				Object:    "response",
				CreatedAt: 1234567890,
				Status:    "in_progress",
				Output: []OutputItem{
					{
						Type:      "function_call",
						CallId:    "call_multi_123",
						Name:      "test_function",
						Arguments: `{"param": "value"}`,
					},
				},
			}

			chatCompletionChunk := ConvertResponseAPIStreamToChatCompletionWithIndex(responseAPIChunk, tc.outputIndex)

			// Verify the index is set correctly
			if len(chatCompletionChunk.Choices) != 1 {
				t.Fatalf("Expected 1 choice, got %d", len(chatCompletionChunk.Choices))
			}

			choice := chatCompletionChunk.Choices[0]
			if len(choice.Delta.ToolCalls) != 1 {
				t.Fatalf("Expected 1 tool call, got %d", len(choice.Delta.ToolCalls))
			}

			toolCall := choice.Delta.ToolCalls[0]
			if toolCall.Index == nil {
				t.Error("Index field should be set for streaming tool calls")
			} else if *toolCall.Index != tc.expectedIdx {
				t.Errorf("Expected index to be %d, got %d", tc.expectedIdx, *toolCall.Index)
			}
		})
	}
}

func TestResponseAPIUsageConversion(t *testing.T) {
	// Test JSON containing OpenAI Response API usage format
	responseJSON := `{
		"id": "resp_test",
		"object": "response",
		"created_at": 1749860991,
		"status": "completed",
		"model": "gpt-4o-2024-11-20",
		"output": [],
		"usage": {
			"input_tokens": 97,
			"output_tokens": 165,
			"total_tokens": 262
		}
	}`

	var responseAPI ResponseAPIResponse
	err := json.Unmarshal([]byte(responseJSON), &responseAPI)
	if err != nil {
		t.Fatalf("Failed to unmarshal ResponseAPI: %v", err)
	}

	// Verify the ResponseAPIUsage fields are correctly parsed
	if responseAPI.Usage == nil {
		t.Fatal("Usage should not be nil")
	}

	if responseAPI.Usage.InputTokens != 97 {
		t.Errorf("Expected InputTokens to be 97, got %d", responseAPI.Usage.InputTokens)
	}

	if responseAPI.Usage.OutputTokens != 165 {
		t.Errorf("Expected OutputTokens to be 165, got %d", responseAPI.Usage.OutputTokens)
	}

	if responseAPI.Usage.TotalTokens != 262 {
		t.Errorf("Expected TotalTokens to be 262, got %d", responseAPI.Usage.TotalTokens)
	}

	// Test conversion to model.Usage
	modelUsage := responseAPI.Usage.ToModelUsage()
	if modelUsage == nil {
		t.Fatal("Converted usage should not be nil")
	}

	if modelUsage.PromptTokens != 97 {
		t.Errorf("Expected PromptTokens to be 97, got %d", modelUsage.PromptTokens)
	}

	if modelUsage.CompletionTokens != 165 {
		t.Errorf("Expected CompletionTokens to be 165, got %d", modelUsage.CompletionTokens)
	}

	if modelUsage.TotalTokens != 262 {
		t.Errorf("Expected TotalTokens to be 262, got %d", modelUsage.TotalTokens)
	}

	// Test conversion to ChatCompletion format
	chatCompletion := ConvertResponseAPIToChatCompletion(&responseAPI)
	if chatCompletion == nil {
		t.Fatal("Converted chat completion should not be nil")
	}

	if chatCompletion.Usage.PromptTokens != 97 {
		t.Errorf("Expected PromptTokens to be 97, got %d", chatCompletion.Usage.PromptTokens)
	}

	if chatCompletion.Usage.CompletionTokens != 165 {
		t.Errorf("Expected CompletionTokens to be 165, got %d", chatCompletion.Usage.CompletionTokens)
	}

	if chatCompletion.Usage.TotalTokens != 262 {
		t.Errorf("Expected TotalTokens to be 262, got %d", chatCompletion.Usage.TotalTokens)
	}
}

func TestResponseAPIUsageWithFallback(t *testing.T) {
	// Test case 1: No usage provided by OpenAI
	responseWithoutUsage := `{
		"id": "resp_no_usage",
		"object": "response",
		"created_at": 1749860991,
		"status": "completed",
		"model": "gpt-4o-2024-11-20",
		"output": [
			{
				"type": "message",
				"role": "assistant",
				"content": [
					{
						"type": "output_text",
						"text": "Hello! How can I help you today?"
					}
				]
			}
		]
	}`

	var responseAPI ResponseAPIResponse
	err := json.Unmarshal([]byte(responseWithoutUsage), &responseAPI)
	if err != nil {
		t.Fatalf("Failed to unmarshal ResponseAPI: %v", err)
	}

	// Convert to ChatCompletion format
	chatCompletion := ConvertResponseAPIToChatCompletion(&responseAPI)

	// Usage should be zero/empty since no usage was provided and no fallback calculation is done in the conversion function
	if chatCompletion.Usage.PromptTokens != 0 || chatCompletion.Usage.CompletionTokens != 0 {
		t.Errorf("Expected zero usage when no usage provided, got prompt=%d, completion=%d",
			chatCompletion.Usage.PromptTokens, chatCompletion.Usage.CompletionTokens)
	}

	// Test case 2: Zero usage provided by OpenAI
	responseWithZeroUsage := `{
		"id": "resp_zero_usage",
		"object": "response",
		"created_at": 1749860991,
		"status": "completed",
		"model": "gpt-4o-2024-11-20",
		"output": [
			{
				"type": "message",
				"role": "assistant",
				"content": [
					{
						"type": "output_text",
						"text": "This is a test response"
					}
				]
			}
		],
		"usage": {
			"input_tokens": 0,
			"output_tokens": 0,
			"total_tokens": 0
		}
	}`

	err = json.Unmarshal([]byte(responseWithZeroUsage), &responseAPI)
	if err != nil {
		t.Fatalf("Failed to unmarshal ResponseAPI with zero usage: %v", err)
	}

	// Convert to ChatCompletion format
	chatCompletion = ConvertResponseAPIToChatCompletion(&responseAPI)

	// Usage should still be zero since the conversion function doesn't set zero usage
	if chatCompletion.Usage.PromptTokens != 0 || chatCompletion.Usage.CompletionTokens != 0 {
		t.Errorf("Expected zero usage when zero usage provided, got prompt=%d, completion=%d",
			chatCompletion.Usage.PromptTokens, chatCompletion.Usage.CompletionTokens)
	}
}
