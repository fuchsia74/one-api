package openai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func TestAdaptorIntegration(t *testing.T) {
	// Create a mock request
	chatRequest := &model.GeneralOpenAIRequest{
		Model: "gpt-4",
		Messages: []model.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "Hello, world!"},
		},
		MaxTokens:   150,
		Temperature: floatPtr(0.8),
		Stream:      false,
		Tools: []model.Tool{
			{
				Type: "function",
				Function: &model.Function{
					Name:        "get_weather",
					Description: "Get current weather",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"location": map[string]any{
								"type":        "string",
								"description": "City name",
							},
						},
						"required": []string{"location"},
					},
				},
			},
		},
		ToolChoice: "auto",
	}

	// Create Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}

	// Create meta context
	testMeta := &meta.Meta{
		Mode:        relaymode.ChatCompletions,
		ChannelType: 1, // OpenAI channel type
	}
	c.Set("meta", testMeta)

	// Create adaptor
	adaptor := &Adaptor{}
	adaptor.Init(testMeta)

	// Test URL generation
	url, err := adaptor.GetRequestURL(testMeta)
	if err != nil {
		t.Fatalf("GetRequestURL failed: %v", err)
	}

	expectedURL := "/v1/responses"
	if !contains(url, expectedURL) {
		t.Errorf("Expected URL to contain %s, got %s", expectedURL, url)
	}

	// Test request conversion
	convertedReq, err := adaptor.ConvertRequest(c, relaymode.ChatCompletions, chatRequest)
	if err != nil {
		t.Fatalf("ConvertRequest failed: %v", err)
	}

	// Verify it's converted to ResponseAPIRequest
	responseAPIReq, ok := convertedReq.(*ResponseAPIRequest)
	if !ok {
		t.Fatalf("Expected ResponseAPIRequest, got %T", convertedReq)
	}

	// Verify basic fields
	if responseAPIReq.Model != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %s", responseAPIReq.Model)
	}

	if responseAPIReq.MaxOutputTokens == nil || *responseAPIReq.MaxOutputTokens != 150 {
		t.Errorf("Expected MaxOutputTokens 150, got %v", responseAPIReq.MaxOutputTokens)
	}

	if responseAPIReq.Temperature == nil || *responseAPIReq.Temperature != 0.8 {
		t.Errorf("Expected Temperature 0.8, got %v", responseAPIReq.Temperature)
	}

	// Verify system message was moved to instructions
	if responseAPIReq.Instructions == nil || *responseAPIReq.Instructions != "You are a helpful assistant." {
		t.Errorf("Expected instructions to be set, got %v", responseAPIReq.Instructions)
	}

	// Verify input messages (should have 1 message after system is removed)
	if len(responseAPIReq.Input) != 1 {
		t.Errorf("Expected 1 input message, got %d", len(responseAPIReq.Input))
	}

	// Verify the input message is correctly structured
	inputMessage, ok := responseAPIReq.Input[0].(model.Message)
	if !ok {
		t.Errorf("Expected input[0] to be model.Message, got %T", responseAPIReq.Input[0])
	} else if inputMessage.Role != "user" {
		t.Errorf("Expected input message role 'user', got '%s'", inputMessage.Role)
	}

	// Verify tools are preserved
	if len(responseAPIReq.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(responseAPIReq.Tools))
	}

	// Verify the request can be marshaled to JSON
	jsonData, err := json.Marshal(responseAPIReq)
	if err != nil {
		t.Fatalf("Failed to marshal ResponseAPIRequest: %v", err)
	}

	// Verify it can be unmarshaled back
	var unmarshaled ResponseAPIRequest
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ResponseAPIRequest: %v", err)
	}

	t.Logf("Successfully converted ChatCompletion to Response API format")
	t.Logf("JSON: %s", string(jsonData))
}

func TestAdaptorNonChatCompletion(t *testing.T) {
	// Test that non-chat completion requests are not converted
	embeddingRequest := &model.GeneralOpenAIRequest{
		Model: "text-embedding-ada-002",
		Input: "Test text",
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{}

	testMeta := &meta.Meta{
		Mode:        relaymode.Embeddings,
		ChannelType: 1,
	}
	c.Set("meta", testMeta)

	adaptor := &Adaptor{}
	adaptor.Init(testMeta)

	// Test that non-chat completion requests are not converted
	convertedReq, err := adaptor.ConvertRequest(c, relaymode.Embeddings, embeddingRequest)
	if err != nil {
		t.Fatalf("ConvertRequest failed: %v", err)
	}

	// Should return the original request unchanged
	originalReq, ok := convertedReq.(*model.GeneralOpenAIRequest)
	if !ok {
		t.Fatalf("Expected GeneralOpenAIRequest, got %T", convertedReq)
	}

	if originalReq.Model != "text-embedding-ada-002" {
		t.Errorf("Expected model unchanged, got %s", originalReq.Model)
	}
}

// Helper functions
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
