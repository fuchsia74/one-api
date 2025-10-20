package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestToolIndexField tests that the Index field is properly serialized in streaming tool calls
func TestToolIndexField(t *testing.T) {
	// Test streaming tool call with Index field set
	index := 0
	streamingTool := Tool{
		Id:   "call_123",
		Type: "function",
		Function: &Function{
			Name:      "get_weather",
			Arguments: `{"location": "Paris"}`,
		},
		Index: &index,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(streamingTool)
	if err != nil {
		t.Fatalf("Failed to marshal streaming tool: %v", err)
	}

	// Verify that the index field is present in JSON
	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check that index field exists and has correct value
	if indexValue, exists := result["index"]; !exists {
		t.Error("Index field is missing from JSON output")
	} else if indexValue != float64(0) { // JSON numbers are float64
		t.Errorf("Expected index to be 0, got %v", indexValue)
	}

	// Test non-streaming tool call without Index field
	nonStreamingTool := Tool{
		Id:   "call_456",
		Type: "function",
		Function: &Function{
			Name:      "send_email",
			Arguments: `{"to": "test@example.com"}`,
		},
		// Index is nil for non-streaming responses
	}

	// Serialize to JSON
	jsonData2, err := json.Marshal(nonStreamingTool)
	if err != nil {
		t.Fatalf("Failed to marshal non-streaming tool: %v", err)
	}

	// Verify that the index field is omitted in JSON (due to omitempty)
	var result2 map[string]any
	err = json.Unmarshal(jsonData2, &result2)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check that index field does not exist
	if _, exists := result2["index"]; exists {
		t.Error("Index field should be omitted for non-streaming tool calls")
	}
}

// TestStreamingToolCallAccumulation tests the complete streaming tool call accumulation workflow
func TestStreamingToolCallAccumulation(t *testing.T) {
	// Simulate streaming tool call deltas as they would come from the API
	streamingDeltas := []Tool{
		{
			Id:    "call_123",
			Type:  "function",
			Index: intPtr(0),
			Function: &Function{
				Name:      "get_weather",
				Arguments: "",
			},
		},
		{
			Index: intPtr(0),
			Function: &Function{
				Arguments: `{"location":`,
			},
		},
		{
			Index: intPtr(0),
			Function: &Function{
				Arguments: ` "Paris"}`,
			},
		},
	}

	// Accumulate the deltas (simulating client-side accumulation)
	finalToolCalls := make(map[int]Tool)

	for _, delta := range streamingDeltas {
		if delta.Index == nil {
			t.Error("Index field should be present in streaming tool call deltas")
			continue
		}

		index := *delta.Index

		if _, exists := finalToolCalls[index]; !exists {
			// First delta for this tool call
			finalToolCalls[index] = delta
		} else {
			// Subsequent delta - accumulate arguments
			existing := finalToolCalls[index]
			existingArgs, _ := existing.Function.Arguments.(string)
			deltaArgs, _ := delta.Function.Arguments.(string)
			existing.Function.Arguments = existingArgs + deltaArgs
			finalToolCalls[index] = existing
		}
	}

	// Verify the final accumulated tool call
	if len(finalToolCalls) != 1 {
		t.Fatalf("Expected 1 final tool call, got %d", len(finalToolCalls))
	}

	finalTool := finalToolCalls[0]
	expectedArgs := `{"location": "Paris"}`
	actualArgs, _ := finalTool.Function.Arguments.(string)
	if actualArgs != expectedArgs {
		t.Errorf("Expected accumulated arguments '%s', got '%s'", expectedArgs, actualArgs)
	}

	if finalTool.Id != "call_123" {
		t.Errorf("Expected tool call id 'call_123', got '%s'", finalTool.Id)
	}

	if finalTool.Function == nil || finalTool.Function.Name != "get_weather" {
		t.Errorf("Expected function name 'get_weather', got '%v'", finalTool.Function)
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}

// TestToolIndexFieldDeserialization tests that the Index field can be properly deserialized
func TestToolIndexFieldDeserialization(t *testing.T) {
	// JSON with index field (streaming response)
	streamingJSON := `{
		"id": "call_789",
		"type": "function",
		"function": {
			"name": "calculate",
			"arguments": "{\"x\": 5, \"y\": 3}"
		},
		"index": 1
	}`

	var streamingTool Tool
	err := json.Unmarshal([]byte(streamingJSON), &streamingTool)
	if err != nil {
		t.Fatalf("Failed to unmarshal streaming tool JSON: %v", err)
	}

	// Verify index field is properly set
	if streamingTool.Index == nil {
		t.Error("Index field should not be nil for streaming tool")
	} else if *streamingTool.Index != 1 {
		t.Errorf("Expected index to be 1, got %d", *streamingTool.Index)
	}

	// JSON without index field (non-streaming response)
	nonStreamingJSON := `{
		"id": "call_101",
		"type": "function",
		"function": {
			"name": "search",
			"arguments": "{\"query\": \"test\"}"
		}
	}`

	var nonStreamingTool Tool
	err = json.Unmarshal([]byte(nonStreamingJSON), &nonStreamingTool)
	if err != nil {
		t.Fatalf("Failed to unmarshal non-streaming tool JSON: %v", err)
	}

	// Verify index field is nil
	if nonStreamingTool.Index != nil {
		t.Error("Index field should be nil for non-streaming tool")
	}
}

// TestMCPToolSerialization tests that MCP tools are properly serialized with all MCP fields
func TestMCPToolSerialization(t *testing.T) {
	// Test MCP tool with all fields populated
	mcpTool := Tool{
		Id:              "mcp_001",
		Type:            "mcp",
		ServerLabel:     "deepwiki",
		ServerUrl:       "https://mcp.deepwiki.com/mcp",
		RequireApproval: "never",
		AllowedTools:    []string{"ask_question", "read_wiki_structure"},
		Headers: map[string]string{
			"Authorization":   "Bearer token123",
			"X-Custom-Header": "custom_value",
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(mcpTool)
	if err != nil {
		t.Fatalf("Failed to marshal MCP tool: %v", err)
	}

	// Verify all MCP fields are present
	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check all MCP-specific fields
	if serverLabel, exists := result["server_label"]; !exists || serverLabel != "deepwiki" {
		t.Errorf("Expected server_label to be 'deepwiki', got %v", serverLabel)
	}

	if serverUrl, exists := result["server_url"]; !exists || serverUrl != "https://mcp.deepwiki.com/mcp" {
		t.Errorf("Expected server_url to be 'https://mcp.deepwiki.com/mcp', got %v", serverUrl)
	}

	if requireApproval, exists := result["require_approval"]; !exists || requireApproval != "never" {
		t.Errorf("Expected require_approval to be 'never', got %v", requireApproval)
	}

	// Verify function field is NOT present for MCP tools (since Function is a pointer and nil)
	if _, exists := result["function"]; exists {
		t.Error("Function field should not be present for MCP tools")
	}
}

// TestMCPToolDeserialization tests that MCP tools can be properly deserialized
func TestMCPToolDeserialization(t *testing.T) {
	// JSON for MCP tool
	mcpJSON := `{
		"id": "mcp_002",
		"type": "mcp",
		"server_label": "stripe",
		"server_url": "https://mcp.stripe.com",
		"require_approval": {
			"never": {
				"tool_names": ["create_payment_link", "get_balance"]
			}
		},
		"allowed_tools": ["create_payment_link", "get_balance", "list_customers"],
		"headers": {
			"Authorization": "Bearer sk_test_123",
			"Content-Type": "application/json"
		}
	}`

	var mcpTool Tool
	err := json.Unmarshal([]byte(mcpJSON), &mcpTool)
	if err != nil {
		t.Fatalf("Failed to unmarshal MCP tool JSON: %v", err)
	}

	// Verify all fields are properly set
	if mcpTool.Id != "mcp_002" {
		t.Errorf("Expected id to be 'mcp_002', got '%s'", mcpTool.Id)
	}

	if mcpTool.Type != "mcp" {
		t.Errorf("Expected type to be 'mcp', got '%s'", mcpTool.Type)
	}

	if mcpTool.ServerLabel != "stripe" {
		t.Errorf("Expected server_label to be 'stripe', got '%s'", mcpTool.ServerLabel)
	}

	if mcpTool.ServerUrl != "https://mcp.stripe.com" {
		t.Errorf("Expected server_url to be 'https://mcp.stripe.com', got '%s'", mcpTool.ServerUrl)
	}

	// Check allowed_tools slice
	expectedTools := []string{"create_payment_link", "get_balance", "list_customers"}
	if len(mcpTool.AllowedTools) != len(expectedTools) {
		t.Errorf("Expected %d allowed tools, got %d", len(expectedTools), len(mcpTool.AllowedTools))
	}

	// Check headers map
	if mcpTool.Headers["Authorization"] != "Bearer sk_test_123" {
		t.Errorf("Expected Authorization header to be 'Bearer sk_test_123', got '%s'", mcpTool.Headers["Authorization"])
	}
}

// TestMCPRequireApprovalVariations tests different RequireApproval configurations
func TestMCPRequireApprovalVariations(t *testing.T) {
	testCases := []struct {
		name     string
		approval any
		jsonStr  string
	}{
		{
			name:     "String never",
			approval: "never",
			jsonStr:  `"never"`,
		},
		{
			name: "Object with tool names",
			approval: map[string]any{
				"never": map[string]any{
					"tool_names": []string{"tool1", "tool2"},
				},
			},
			jsonStr: `{"never":{"tool_names":["tool1","tool2"]}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mcpTool := Tool{
				Type:            "mcp",
				ServerLabel:     "test",
				RequireApproval: tc.approval,
			}

			jsonData, err := json.Marshal(mcpTool)
			if err != nil {
				t.Fatalf("Failed to marshal MCP tool: %v", err)
			}

			// Verify the require_approval field is serialized correctly
			var result map[string]any
			err = json.Unmarshal(jsonData, &result)
			if err != nil {
				t.Fatalf("Failed to unmarshal JSON: %v", err)
			}

			// Convert require_approval back to JSON to compare
			approvalBytes, err := json.Marshal(result["require_approval"])
			if err != nil {
				t.Fatalf("Failed to marshal require_approval: %v", err)
			}

			if string(approvalBytes) != tc.jsonStr {
				t.Errorf("Expected require_approval JSON to be %s, got %s", tc.jsonStr, string(approvalBytes))
			}
		})
	}
}

// TestMixedToolArray tests arrays containing both function and MCP tools
func TestMixedToolArray(t *testing.T) {
	tools := []Tool{
		{
			Id:   "func_001",
			Type: "function",
			Function: &Function{
				Name:        "get_weather",
				Description: "Get weather information",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The location to get weather for",
						},
					},
					"required": []string{"location"},
				},
			},
		},
		{
			Id:              "mcp_001",
			Type:            "mcp",
			ServerLabel:     "deepwiki",
			ServerUrl:       "https://mcp.deepwiki.com/mcp",
			RequireApproval: "never",
			AllowedTools:    []string{"ask_question"},
			Headers: map[string]string{
				"Authorization": "Bearer token123",
			},
		},
	}

	// Serialize the mixed array
	jsonData, err := json.Marshal(tools)
	if err != nil {
		t.Fatalf("Failed to marshal mixed tool array: %v", err)
	}

	// Deserialize back
	var deserializedTools []Tool
	err = json.Unmarshal(jsonData, &deserializedTools)
	if err != nil {
		t.Fatalf("Failed to unmarshal mixed tool array: %v", err)
	}

	// Verify we have 2 tools
	if len(deserializedTools) != 2 {
		t.Fatalf("Expected 2 tools, got %d", len(deserializedTools))
	}

	// Verify function tool
	funcTool := deserializedTools[0]
	if funcTool.Type != "function" {
		t.Errorf("Expected first tool type to be 'function', got '%s'", funcTool.Type)
	}
	if funcTool.Function.Name != "get_weather" {
		t.Errorf("Expected function name to be 'get_weather', got '%s'", funcTool.Function.Name)
	}
	if funcTool.ServerLabel != "" {
		t.Error("Function tool should not have server_label")
	}

	// Verify MCP tool
	mcpTool := deserializedTools[1]
	if mcpTool.Type != "mcp" {
		t.Errorf("Expected second tool type to be 'mcp', got '%s'", mcpTool.Type)
	}
	if mcpTool.ServerLabel != "deepwiki" {
		t.Errorf("Expected server_label to be 'deepwiki', got '%s'", mcpTool.ServerLabel)
	}
	if mcpTool.Function != nil {
		t.Error("MCP tool should not have function definition")
	}
}

// TestMCPToolEdgeCases tests edge cases and validation scenarios for MCP tools
func TestMCPToolEdgeCases(t *testing.T) {
	// Test MCP tool with minimal fields
	minimalMCP := Tool{
		Type:        "mcp",
		ServerLabel: "minimal",
		ServerUrl:   "https://minimal.example.com",
	}

	jsonData, err := json.Marshal(minimalMCP)
	if err != nil {
		t.Fatalf("Failed to marshal minimal MCP tool: %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check that optional fields are omitted (except function which is always present as empty object)
	optionalFields := []string{"require_approval", "allowed_tools", "headers", "id"}
	for _, field := range optionalFields {
		if _, exists := result[field]; exists {
			t.Errorf("Field '%s' should be omitted for minimal MCP tool", field)
		}
	}

	// Function field should NOT be present for MCP tools
	if _, exists := result["function"]; exists {
		t.Error("Function field should not be present for MCP tools")
	}

	// Test empty headers map is omitted
	emptyHeadersMCP := Tool{
		Type:        "mcp",
		ServerLabel: "test",
		Headers:     map[string]string{},
	}

	jsonData, err = json.Marshal(emptyHeadersMCP)
	if err != nil {
		t.Fatalf("Failed to marshal MCP tool with empty headers: %v", err)
	}

	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if _, exists := result["headers"]; exists {
		t.Error("Empty headers map should be omitted")
	}

	// Test empty allowed_tools slice is omitted
	emptyToolsMCP := Tool{
		Type:         "mcp",
		ServerLabel:  "test",
		AllowedTools: []string{},
	}

	jsonData, err = json.Marshal(emptyToolsMCP)
	if err != nil {
		t.Fatalf("Failed to marshal MCP tool with empty allowed_tools: %v", err)
	}

	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if _, exists := result["allowed_tools"]; exists {
		t.Error("Empty allowed_tools slice should be omitted")
	}
}

func TestToolUnmarshalFlattenedFunction(t *testing.T) {
	jsonStr := `{
		"type": "function",
		"name": "get_weather",
		"description": "Get current temperature for a given location.",
		"parameters": {
			"type": "object",
			"properties": {
				"location": {
					"type": "string"
				}
			},
			"required": ["location"],
			"additionalProperties": false
		},
		"strict": true
	}`

	var tool Tool
	require.NoError(t, json.Unmarshal([]byte(jsonStr), &tool))
	require.NotNil(t, tool.Function)
	require.Equal(t, "function", tool.Type)
	require.Equal(t, "get_weather", tool.Function.Name)
	require.Equal(t, "Get current temperature for a given location.", tool.Function.Description)
	require.NotNil(t, tool.Function.Strict)
	require.True(t, *tool.Function.Strict)
	require.NotNil(t, tool.Function.Parameters)

	encoded, err := json.Marshal(tool)
	require.NoError(t, err)

	var serialized map[string]any
	require.NoError(t, json.Unmarshal(encoded, &serialized))
	require.Equal(t, "function", serialized["type"])

	fn, ok := serialized["function"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "get_weather", fn["name"])
	require.Equal(t, true, fn["strict"])

	_, hasName := serialized["name"]
	require.False(t, hasName)
}
