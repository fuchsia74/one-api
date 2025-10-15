package openai_test

import (
	"encoding/json"
	"testing"

	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

// TestConvertChatCompletionToResponseAPIWithMCP tests that MCP tools are properly converted
// from ChatCompletion format to Response API format, preserving all MCP-specific fields
func TestConvertChatCompletionToResponseAPIWithMCP(t *testing.T) {
	// Create a ChatCompletion request with MCP tool (matches the failing curl example)
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4o",
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "What transport protocols are supported in the 2025-03-26 version of the MCP spec?",
			},
		},
		Tools: []model.Tool{
			{
				Type:            "mcp",
				ServerLabel:     "deepwiki",
				ServerUrl:       "https://mcp.deepwiki.com/mcp",
				RequireApproval: "never",
			},
		},
	}

	// Convert to Response API format
	responseAPI := openai.ConvertChatCompletionToResponseAPI(chatRequest)

	// Verify the conversion succeeded
	if responseAPI == nil {
		t.Fatal("ConvertChatCompletionToResponseAPI returned nil")
	}

	// Verify tools were converted properly
	if len(responseAPI.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(responseAPI.Tools))
	}

	tool := responseAPI.Tools[0]

	// Verify MCP-specific fields are preserved
	if tool.Type != "mcp" {
		t.Errorf("Expected tool type 'mcp', got '%s'", tool.Type)
	}

	if tool.ServerLabel != "deepwiki" {
		t.Errorf("Expected server_label 'deepwiki', got '%s'", tool.ServerLabel)
	}

	if tool.ServerUrl != "https://mcp.deepwiki.com/mcp" {
		t.Errorf("Expected server_url 'https://mcp.deepwiki.com/mcp', got '%s'", tool.ServerUrl)
	}

	if tool.RequireApproval != "never" {
		t.Errorf("Expected require_approval 'never', got %v", tool.RequireApproval)
	}

	// Verify function-specific fields are empty for MCP tools
	if tool.Name != "" {
		t.Errorf("Expected empty name for MCP tool, got '%s'", tool.Name)
	}

	if tool.Description != "" {
		t.Errorf("Expected empty description for MCP tool, got '%s'", tool.Description)
	}

	if tool.Parameters != nil {
		t.Errorf("Expected nil parameters for MCP tool, got %v", tool.Parameters)
	}

	// Verify the tool can be marshaled to JSON (important for the actual API request)
	jsonData, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("Failed to marshal MCP tool to JSON: %v", err)
	}

	// Verify the JSON contains the required server_label field
	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if serverLabel, exists := result["server_label"]; !exists {
		t.Error("server_label field is missing from JSON - this would cause the original error")
	} else if serverLabel != "deepwiki" {
		t.Errorf("Expected server_label 'deepwiki' in JSON, got %v", serverLabel)
	}

	t.Logf("Successfully converted MCP tool to Response API format: %s", string(jsonData))
}

// TestConvertChatCompletionToResponseAPIWithMCPAndFunction tests mixed MCP and function tools
func TestConvertChatCompletionToResponseAPIWithMCPAndFunction(t *testing.T) {
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4o",
		Messages: []model.Message{
			{
				Role:    "user",
				Content: "Test mixed tools",
			},
		},
		Tools: []model.Tool{
			{
				Type: "function",
				Function: &model.Function{
					Name:        "get_weather",
					Description: "Get weather information",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"location": map[string]any{
								"type": "string",
							},
						},
					},
				},
			},
			{
				Type:        "mcp",
				ServerLabel: "stripe",
				ServerUrl:   "https://mcp.stripe.com",
				RequireApproval: map[string]any{
					"never": map[string]any{
						"tool_names": []string{"create_payment_link"},
					},
				},
				AllowedTools: []string{"create_payment_link", "get_balance"},
				Headers: map[string]string{
					"Authorization": "Bearer sk_test_123",
				},
			},
		},
	}

	responseAPI := openai.ConvertChatCompletionToResponseAPI(chatRequest)

	if len(responseAPI.Tools) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(responseAPI.Tools))
	}

	// Verify function tool
	functionTool := responseAPI.Tools[0]
	if functionTool.Type != "function" {
		t.Errorf("Expected first tool type 'function', got '%s'", functionTool.Type)
	}
	if functionTool.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got '%s'", functionTool.Name)
	}
	if functionTool.ServerLabel != "" {
		t.Error("Function tool should not have server_label")
	}

	// Verify MCP tool
	mcpTool := responseAPI.Tools[1]
	if mcpTool.Type != "mcp" {
		t.Errorf("Expected second tool type 'mcp', got '%s'", mcpTool.Type)
	}
	if mcpTool.ServerLabel != "stripe" {
		t.Errorf("Expected server_label 'stripe', got '%s'", mcpTool.ServerLabel)
	}
	if len(mcpTool.AllowedTools) != 2 {
		t.Errorf("Expected 2 allowed tools, got %d", len(mcpTool.AllowedTools))
	}
	if mcpTool.Headers["Authorization"] != "Bearer sk_test_123" {
		t.Errorf("Expected Authorization header, got '%s'", mcpTool.Headers["Authorization"])
	}
	if mcpTool.Name != "" {
		t.Error("MCP tool should not have function name")
	}
}

// TestMCPToolJSONSerialization tests that the converted MCP tool produces valid JSON
func TestMCPToolJSONSerialization(t *testing.T) {
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4o",
		Messages: []model.Message{
			{Role: "user", Content: "Test"},
		},
		Tools: []model.Tool{
			{
				Type:            "mcp",
				ServerLabel:     "deepwiki",
				ServerUrl:       "https://mcp.deepwiki.com/mcp",
				RequireApproval: "never",
			},
		},
	}

	responseAPI := openai.ConvertChatCompletionToResponseAPI(chatRequest)

	// Marshal the entire request to JSON
	jsonData, err := json.Marshal(responseAPI)
	if err != nil {
		t.Fatalf("Failed to marshal Response API request: %v", err)
	}

	// Verify it can be unmarshaled back
	var unmarshaled openai.ResponseAPIRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Response API request: %v", err)
	}

	// Verify MCP tool fields are preserved
	if len(unmarshaled.Tools) != 1 {
		t.Fatalf("Expected 1 tool after round-trip, got %d", len(unmarshaled.Tools))
	}

	tool := unmarshaled.Tools[0]
	if tool.ServerLabel != "deepwiki" {
		t.Errorf("server_label lost during JSON round-trip: got '%s'", tool.ServerLabel)
	}

	t.Logf("JSON serialization successful: %s", string(jsonData))
}
