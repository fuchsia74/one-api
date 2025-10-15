package openai_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

// TestMCPOutputItemSerialization tests that MCP-specific OutputItem fields are properly serialized
func TestMCPOutputItemSerialization(t *testing.T) {
	// Test mcp_list_tools output item
	mcpListTools := openai.OutputItem{
		Type:        "mcp_list_tools",
		Id:          "mcpl_682d4379df088191886b70f4ec39f90403937d5f622d7a90",
		ServerLabel: "deepwiki",
		Tools: []model.Tool{
			{
				Type: "function",
				Function: &model.Function{
					Name:        "read_wiki_structure",
					Description: "Read repository structure",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"repoName": map[string]any{
								"type":        "string",
								"description": "GitHub repository: owner/repo (e.g. \"facebook/react\")",
							},
						},
						"required": []string{"repoName"},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(mcpListTools)
	if err != nil {
		t.Fatalf("Failed to marshal mcp_list_tools: %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify MCP-specific fields
	if result["type"] != "mcp_list_tools" {
		t.Errorf("Expected type 'mcp_list_tools', got %v", result["type"])
	}
	if result["server_label"] != "deepwiki" {
		t.Errorf("Expected server_label 'deepwiki', got %v", result["server_label"])
	}
	if tools, ok := result["tools"].([]any); !ok || len(tools) != 1 {
		t.Errorf("Expected tools array with 1 item, got %v", result["tools"])
	}
}

// TestMCPCallOutputItem tests mcp_call output item serialization
func TestMCPCallOutputItem(t *testing.T) {
	// Test successful mcp_call
	mcpCall := openai.OutputItem{
		Type:        "mcp_call",
		Id:          "mcp_682d437d90a88191bf88cd03aae0c3e503937d5f622d7a90",
		ServerLabel: "deepwiki",
		Name:        "ask_question",
		Arguments:   "{\"repoName\":\"modelcontextprotocol/modelcontextprotocol\",\"question\":\"What transport protocols does the 2025-03-26 version of the MCP spec support?\"}",
		Output:      "The 2025-03-26 version of the Model Context Protocol (MCP) specification supports two standard transport mechanisms: `stdio` and `Streamable HTTP`",
	}

	jsonData, err := json.Marshal(mcpCall)
	if err != nil {
		t.Fatalf("Failed to marshal mcp_call: %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify fields
	if result["type"] != "mcp_call" {
		t.Errorf("Expected type 'mcp_call', got %v", result["type"])
	}
	if result["name"] != "ask_question" {
		t.Errorf("Expected name 'ask_question', got %v", result["name"])
	}
	if result["server_label"] != "deepwiki" {
		t.Errorf("Expected server_label 'deepwiki', got %v", result["server_label"])
	}
	if result["output"] == "" {
		t.Error("Expected output to be present")
	}
}

// TestMCPCallWithError tests mcp_call output item with error
func TestMCPCallWithError(t *testing.T) {
	errorMsg := "Connection failed"
	mcpCallError := openai.OutputItem{
		Type:        "mcp_call",
		Id:          "mcp_error_123",
		ServerLabel: "stripe",
		Name:        "create_payment_link",
		Arguments:   "{\"amount\":2000}",
		Error:       &errorMsg,
	}

	jsonData, err := json.Marshal(mcpCallError)
	if err != nil {
		t.Fatalf("Failed to marshal mcp_call with error: %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["error"] != "Connection failed" {
		t.Errorf("Expected error 'Connection failed', got %v", result["error"])
	}
}

// TestMCPApprovalRequest tests mcp_approval_request output item
func TestMCPApprovalRequest(t *testing.T) {
	approvalRequest := openai.OutputItem{
		Type:        "mcp_approval_request",
		Id:          "mcpr_682d498e3bd4819196a0ce1664f8e77b04ad1e533afccbfa",
		ServerLabel: "deepwiki",
		Name:        "ask_question",
		Arguments:   "{\"repoName\":\"modelcontextprotocol/modelcontextprotocol\",\"question\":\"What transport protocols are supported in the 2025-03-26 version of the MCP spec?\"}",
	}

	jsonData, err := json.Marshal(approvalRequest)
	if err != nil {
		t.Fatalf("Failed to marshal mcp_approval_request: %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify fields
	if result["type"] != "mcp_approval_request" {
		t.Errorf("Expected type 'mcp_approval_request', got %v", result["type"])
	}
	if result["server_label"] != "deepwiki" {
		t.Errorf("Expected server_label 'deepwiki', got %v", result["server_label"])
	}
}

// TestMCPApprovalResponseInput tests the MCP approval response input structure
func TestMCPApprovalResponseInput(t *testing.T) {
	approvalResponse := openai.MCPApprovalResponseInput{
		Type:              "mcp_approval_response",
		Approve:           true,
		ApprovalRequestId: "mcpr_682d498e3bd4819196a0ce1664f8e77b04ad1e533afccbfa",
	}

	jsonData, err := json.Marshal(approvalResponse)
	if err != nil {
		t.Fatalf("Failed to marshal MCPApprovalResponseInput: %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify fields
	if result["type"] != "mcp_approval_response" {
		t.Errorf("Expected type 'mcp_approval_response', got %v", result["type"])
	}
	if result["approve"] != true {
		t.Errorf("Expected approve true, got %v", result["approve"])
	}
	if result["approval_request_id"] != "mcpr_682d498e3bd4819196a0ce1664f8e77b04ad1e533afccbfa" {
		t.Errorf("Expected correct approval_request_id, got %v", result["approval_request_id"])
	}

	// Test deserialization
	var deserializedResponse openai.MCPApprovalResponseInput
	err = json.Unmarshal(jsonData, &deserializedResponse)
	if err != nil {
		t.Fatalf("Failed to deserialize MCPApprovalResponseInput: %v", err)
	}

	if deserializedResponse.Type != "mcp_approval_response" {
		t.Errorf("Expected type 'mcp_approval_response', got %s", deserializedResponse.Type)
	}
	if !deserializedResponse.Approve {
		t.Error("Expected approve to be true")
	}
	if deserializedResponse.ApprovalRequestId != "mcpr_682d498e3bd4819196a0ce1664f8e77b04ad1e533afccbfa" {
		t.Errorf("Expected correct approval_request_id, got %s", deserializedResponse.ApprovalRequestId)
	}
}

// TestResponseAPIResponseWithMCPOutput tests complete ResponseAPIResponse with MCP output items
func TestResponseAPIResponseWithMCPOutput(t *testing.T) {
	response := openai.ResponseAPIResponse{
		Id:        "resp_123",
		Object:    "response",
		Model:     "gpt-4.1",
		Status:    "completed",
		CreatedAt: 1234567890,
		Output: []openai.OutputItem{
			{
				Type:        "mcp_list_tools",
				Id:          "mcpl_456",
				ServerLabel: "deepwiki",
				Tools: []model.Tool{
					{
						Type: "function",
						Function: &model.Function{
							Name:        "ask_question",
							Description: "Ask a question",
						},
					},
				},
			},
			{
				Type:        "mcp_call",
				Id:          "mcp_789",
				ServerLabel: "deepwiki",
				Name:        "ask_question",
				Arguments:   "{\"question\":\"test\"}",
				Output:      "Test response",
			},
		},
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal ResponseAPIResponse with MCP output: %v", err)
	}

	var deserializedResponse openai.ResponseAPIResponse
	err = json.Unmarshal(jsonData, &deserializedResponse)
	if err != nil {
		t.Fatalf("Failed to deserialize ResponseAPIResponse: %v", err)
	}

	if len(deserializedResponse.Output) != 2 {
		t.Errorf("Expected 2 output items, got %d", len(deserializedResponse.Output))
	}

	// Verify first output item (mcp_list_tools)
	firstOutput := deserializedResponse.Output[0]
	if firstOutput.Type != "mcp_list_tools" {
		t.Errorf("Expected first output type 'mcp_list_tools', got %s", firstOutput.Type)
	}
	if len(firstOutput.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(firstOutput.Tools))
	}

	// Verify second output item (mcp_call)
	secondOutput := deserializedResponse.Output[1]
	if secondOutput.Type != "mcp_call" {
		t.Errorf("Expected second output type 'mcp_call', got %s", secondOutput.Type)
	}
	if secondOutput.Output != "Test response" {
		t.Errorf("Expected output 'Test response', got %s", secondOutput.Output)
	}
}

// TestConvertResponseAPIToChatCompletionWithMCP tests MCP output conversion to ChatCompletion
func TestConvertResponseAPIToChatCompletionWithMCP(t *testing.T) {
	responseAPIResp := &openai.ResponseAPIResponse{
		Id:        "resp_mcp_test",
		Model:     "gpt-4.1",
		Status:    "completed",
		CreatedAt: 1234567890,
		Output: []openai.OutputItem{
			{
				Type: "message",
				Role: "assistant",
				Content: []openai.OutputContent{
					{
						Type: "output_text",
						Text: "Hello! I'll help you with that.",
					},
				},
			},
			{
				Type:        "mcp_list_tools",
				ServerLabel: "deepwiki",
				Tools: []model.Tool{
					{Type: "function", Function: &model.Function{Name: "ask_question"}},
				},
			},
			{
				Type:        "mcp_call",
				ServerLabel: "deepwiki",
				Name:        "ask_question",
				Arguments:   "{\"question\":\"test\"}",
				Output:      "The answer is 42",
			},
		},
	}

	chatResponse := openai.ConvertResponseAPIToChatCompletion(responseAPIResp)

	if chatResponse == nil {
		t.Fatal("Expected non-nil chat response")
	}

	if chatResponse.Id != "resp_mcp_test" {
		t.Errorf("Expected ID 'resp_mcp_test', got %s", chatResponse.Id)
	}

	if len(chatResponse.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(chatResponse.Choices))
	}

	choice := chatResponse.Choices[0]
	content := choice.Message.Content.(string)

	// Verify that MCP content was included in the response text
	if !containsSubstring(content, "Hello! I'll help you with that.") {
		t.Error("Expected original message content to be preserved")
	}
	if !containsSubstring(content, "MCP Server 'deepwiki' tools imported: 1 tools available") {
		t.Error("Expected MCP list tools info to be included")
	}
	if !containsSubstring(content, "MCP Tool 'ask_question' result: The answer is 42") {
		t.Error("Expected MCP call result to be included")
	}
}

// Helper function to check if a string contains a substring.
// It simplifies the use of the standard library.
func containsSubstring(s, substr string) bool { return strings.Contains(s, substr) }
