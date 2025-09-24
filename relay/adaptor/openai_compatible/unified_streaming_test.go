package openai_compatible

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Laisky/go-utils/v5/log"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/model"
)

// Test data constants
const (
	testModelName    = "gpt-4-turbo"
	testContentType  = "application/json"
	testPromptTokens = 100
)

// Helper function to create a test logger
func createTestLogger() *log.LoggerT {
	logger, _ := log.New(log.WithLevel("debug"))
	return logger
}

// Helper function to create test streaming response
func createTestStreamResponse(id string, content string, usage *model.Usage) *ChatCompletionsStreamResponse {
	return &ChatCompletionsStreamResponse{
		Id:      id,
		Object:  "chat.completion.chunk",
		Created: 1234567890,
		Model:   testModelName,
		Choices: []ChatCompletionsStreamResponseChoice{
			{
				Index: 0,
				Delta: model.Message{Content: content},
			},
		},
		Usage: usage,
	}
}

// Helper function to create test streaming response with tool calls
func createTestStreamResponseWithToolCalls(id string, toolArgs interface{}) *ChatCompletionsStreamResponse {
	toolCall := model.Tool{
		Function: &model.Function{
			Arguments: toolArgs,
		},
	}

	return &ChatCompletionsStreamResponse{
		Id:      id,
		Object:  "chat.completion.chunk",
		Created: 1234567890,
		Model:   testModelName,
		Choices: []ChatCompletionsStreamResponseChoice{
			{
				Index: 0,
				Delta: model.Message{
					ToolCalls: []model.Tool{toolCall},
				},
			},
		},
	}
}

// TestThinkingProcessor_ProcessThinkingContent tests the ThinkingProcessor functionality
func TestThinkingProcessor_ProcessThinkingContent(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedContent   string
		expectedReasoning *string
		expectedModified  bool
		setupFunc         func(*ThinkingProcessor)
	}{
		{
			name:              "Empty input",
			input:             "",
			expectedContent:   "",
			expectedReasoning: nil,
			expectedModified:  false,
		},
		{
			name:              "No thinking tag",
			input:             "Hello world",
			expectedContent:   "Hello world",
			expectedReasoning: nil,
			expectedModified:  false,
		},
		{
			name:              "Complete thinking block in single chunk",
			input:             "Hello <think>reasoning here</think> world",
			expectedContent:   "Hello  world",
			expectedReasoning: stringPtr("reasoning here"),
			expectedModified:  true,
		},
		{
			name:              "Thinking block at start",
			input:             "<think>reasoning</think>content",
			expectedContent:   "content",
			expectedReasoning: stringPtr("reasoning"),
			expectedModified:  true,
		},
		{
			name:              "Thinking block at end",
			input:             "content<think>reasoning</think>",
			expectedContent:   "content",
			expectedReasoning: stringPtr("reasoning"),
			expectedModified:  true,
		},
		{
			name:              "Empty thinking block",
			input:             "Hello <think></think> world",
			expectedContent:   "Hello  world",
			expectedReasoning: nil,
			expectedModified:  true,
		},
		{
			name:              "Incomplete thinking block - opening tag only",
			input:             "Hello <think>partial reasoning",
			expectedContent:   "Hello ",
			expectedReasoning: stringPtr("partial reasoning"),
			expectedModified:  true,
		},
		{
			name:              "Continue inside thinking block",
			input:             " more reasoning",
			expectedContent:   "",
			expectedReasoning: stringPtr(" more reasoning"),
			expectedModified:  true,
			setupFunc: func(tp *ThinkingProcessor) {
				tp.isInThinkingBlock = true
			},
		},
		{
			name:              "Close thinking block",
			input:             " final reasoning</think> world",
			expectedContent:   " world",
			expectedReasoning: stringPtr(" final reasoning"),
			expectedModified:  true,
			setupFunc: func(tp *ThinkingProcessor) {
				tp.isInThinkingBlock = true
			},
		},
		{
			name:              "Already processed thinking tag",
			input:             "Hello <think>should be ignored</think> world",
			expectedContent:   "Hello <think>should be ignored</think> world",
			expectedReasoning: nil,
			expectedModified:  false,
			setupFunc: func(tp *ThinkingProcessor) {
				tp.hasProcessedThinkTag = true
			},
		},
		{
			name:              "Multiple thinking blocks - only first processed",
			input:             "Hello <think>first</think> middle <think>second</think> world",
			expectedContent:   "Hello  middle <think>second</think> world",
			expectedReasoning: stringPtr("first"),
			expectedModified:  true,
		},
		{
			name:              "JSON-escaped Unicode thinking block (complete)",
			input:             "Hello \\u003cthink\\u003ereasoning\\u003c/think\\u003e world",
			expectedContent:   "Hello  world",
			expectedReasoning: stringPtr("reasoning"),
			expectedModified:  true,
		},
		{
			name:              "JSON-escaped Unicode thinking block (opening only)",
			input:             "Hello \\u003cthink\\u003epartial reasoning",
			expectedContent:   "Hello ",
			expectedReasoning: stringPtr("partial reasoning"),
			expectedModified:  true,
		},
		{
			name:              "Mixed normal and Unicode thinking blocks",
			input:             "Hello \\u003cthink\\u003eunicode first\\u003c/think\\u003e <think>normal second</think>",
			expectedContent:   "Hello  <think>normal second</think>",
			expectedReasoning: stringPtr("unicode first"),
			expectedModified:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &ThinkingProcessor{}
			if tt.setupFunc != nil {
				tt.setupFunc(tp)
			}

			content, reasoning, modified := tp.ProcessThinkingContent(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content %q, got %q", tt.expectedContent, content)
			}

			if tt.expectedReasoning == nil {
				if reasoning != nil {
					t.Errorf("Expected nil reasoning, got %q", *reasoning)
				}
			} else {
				if reasoning == nil {
					t.Errorf("Expected reasoning %q, got nil", *tt.expectedReasoning)
				} else if *reasoning != *tt.expectedReasoning {
					t.Errorf("Expected reasoning %q, got %q", *tt.expectedReasoning, *reasoning)
				}
			}

			if modified != tt.expectedModified {
				t.Errorf("Expected modified %v, got %v", tt.expectedModified, modified)
			}
		})
	}
}

// TestStreamingContext_NewStreamingContext tests context initialization
func TestStreamingContext_NewStreamingContext(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		enableThinking bool
		expectThinking bool
	}{
		{
			name:           "With thinking processor",
			enableThinking: true,
			expectThinking: true,
		},
		{
			name:           "Without thinking processor",
			enableThinking: false,
			expectThinking: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewStreamingContext(logger, tt.enableThinking)

			if ctx == nil {
				t.Fatal("Expected non-nil context")
			}

			if ctx.logger != logger {
				t.Error("Logger not properly set")
			}

			if tt.expectThinking {
				if ctx.thinkingProcessor == nil {
					t.Error("Expected thinking processor to be created")
				}
			} else {
				if ctx.thinkingProcessor != nil {
					t.Error("Expected thinking processor to be nil")
				}
			}

			// Check buffer pre-allocation
			if ctx.responseTextBuilder.Cap() == 0 {
				t.Error("Expected responseTextBuilder to be pre-allocated")
			}
			if ctx.toolArgsTextBuilder.Cap() == 0 {
				t.Error("Expected toolArgsTextBuilder to be pre-allocated")
			}

			// Verify initial state
			if ctx.chunksProcessed != 0 {
				t.Errorf("Expected chunksProcessed to be 0, got %d", ctx.chunksProcessed)
			}
			if ctx.doneRendered {
				t.Error("Expected doneRendered to be false")
			}
		})
	}
}

// TestStreamingContext_ProcessStreamChunk tests chunk processing
func TestStreamingContext_ProcessStreamChunk(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name           string
		enableThinking bool
		response       *ChatCompletionsStreamResponse
		expectModified bool
	}{
		{
			name:           "Basic content chunk",
			enableThinking: false,
			response:       createTestStreamResponse("test-1", "Hello world", nil),
			expectModified: true,
		},
		{
			name:           "Content with thinking block",
			enableThinking: true,
			response:       createTestStreamResponse("test-2", "Hello <think>reasoning</think> world", nil),
			expectModified: true,
		},
		{
			name:           "Chunk with usage",
			enableThinking: false,
			response: createTestStreamResponse("test-3", "Hello", &model.Usage{
				PromptTokens:     50,
				CompletionTokens: 25,
				TotalTokens:      75,
			}),
			expectModified: true,
		},
		{
			name:           "String tool call arguments",
			enableThinking: false,
			response:       createTestStreamResponseWithToolCalls("test-4", `{"arg": "value"}`),
			expectModified: true,
		},
		{
			name:           "Object tool call arguments",
			enableThinking: false,
			response:       createTestStreamResponseWithToolCalls("test-5", map[string]interface{}{"arg": "value"}),
			expectModified: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewStreamingContext(logger, tt.enableThinking)
			originalContent := ""
			if len(tt.response.Choices) > 0 {
				originalContent = tt.response.Choices[0].Delta.StringContent()
			}

			modified := ctx.ProcessStreamChunk(tt.response)

			if modified != tt.expectModified {
				t.Errorf("Expected modified %v, got %v", tt.expectModified, modified)
			}

			// Verify content accumulation
			responseText := ctx.responseTextBuilder.String()
			// Note: responseTextBuilder always contains original content (before thinking processing)
			// The thinking processing modifies the response.Choices[].Delta.Content but not the builder
			if originalContent != "" {
				if !strings.Contains(responseText, originalContent) {
					t.Error("Expected original content to be preserved in responseTextBuilder")
				}
			}

			// Verify usage accumulation
			if tt.response.Usage != nil {
				if ctx.usage != tt.response.Usage {
					t.Error("Expected usage to be accumulated")
				}
			}

			// Verify counter increment
			if ctx.chunksProcessed != 1 {
				t.Errorf("Expected chunksProcessed to be 1, got %d", ctx.chunksProcessed)
			}
		})
	}
}

// TestStreamingContext_ManageBufferCapacity tests buffer management
func TestStreamingContext_ManageBufferCapacity(t *testing.T) {
	logger := createTestLogger()
	ctx := NewStreamingContext(logger, false)

	// Test functional behavior rather than exact capacity values
	// Force buffer to significantly exceed max capacity
	largeContent := strings.Repeat("x", MaxBuilderCapacity+100000) // Much larger to ensure management kicks in
	ctx.responseTextBuilder.WriteString(largeContent)
	ctx.toolArgsTextBuilder.WriteString(largeContent)

	// Check capacity before management
	responseCapBefore := ctx.responseTextBuilder.Cap()

	if responseCapBefore <= MaxBuilderCapacity {
		t.Skip("Buffer didn't exceed MaxBuilderCapacity, skipping capacity management test")
	}

	// Trigger capacity management
	ctx.ManageBufferCapacity()

	// Verify content was preserved (this is the most important functional test)
	if ctx.responseTextBuilder.String() != largeContent {
		t.Error("Response content not preserved during capacity management")
	}
	if ctx.toolArgsTextBuilder.String() != largeContent {
		t.Error("Tool args content not preserved during capacity management")
	}

	// Check that ManageBufferCapacity method runs without panicking
	// Additional management calls should be safe
	ctx.ManageBufferCapacity()
	ctx.ManageBufferCapacity()

	// Verify content is still preserved after multiple management calls
	if ctx.responseTextBuilder.String() != largeContent {
		t.Error("Response content not preserved after multiple capacity management calls")
	}
	if ctx.toolArgsTextBuilder.String() != largeContent {
		t.Error("Tool args content not preserved after multiple capacity management calls")
	}
}

// TestStreamingContext_CalculateUsage tests usage calculation
func TestStreamingContext_CalculateUsage(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name                     string
		setupFunc                func(*StreamingContext)
		promptTokens             int
		expectedPromptTokens     int
		expectedCompletionTokens int
		expectedTotalTokens      int
		expectFallback           bool
	}{
		{
			name: "No upstream usage - fallback calculation",
			setupFunc: func(ctx *StreamingContext) {
				ctx.responseTextBuilder.WriteString("Hello world")
				ctx.toolArgsTextBuilder.WriteString(`{"arg": "value"}`)
			},
			promptTokens:             testPromptTokens,
			expectedPromptTokens:     testPromptTokens,
			expectedCompletionTokens: (len("Hello world") + len(`{"arg": "value"}`)) / 4, // CountTokenText estimation
			expectedTotalTokens:      testPromptTokens + (len("Hello world")+len(`{"arg": "value"}`))/4,
			expectFallback:           true,
		},
		{
			name: "Complete upstream usage",
			setupFunc: func(ctx *StreamingContext) {
				ctx.usage = &model.Usage{
					PromptTokens:     testPromptTokens,
					CompletionTokens: 50,
					TotalTokens:      testPromptTokens + 50,
				}
			},
			promptTokens:             testPromptTokens,
			expectedPromptTokens:     testPromptTokens,
			expectedCompletionTokens: 50,
			expectedTotalTokens:      testPromptTokens + 50,
			expectFallback:           false,
		},
		{
			name: "Partial upstream usage - missing completion tokens",
			setupFunc: func(ctx *StreamingContext) {
				ctx.responseTextBuilder.WriteString("Hello world")
				ctx.usage = &model.Usage{
					PromptTokens: testPromptTokens,
					TotalTokens:  0, // Will be calculated
				}
			},
			promptTokens:             testPromptTokens,
			expectedPromptTokens:     testPromptTokens,
			expectedCompletionTokens: len("Hello world") / 4,
			expectedTotalTokens:      testPromptTokens + len("Hello world")/4,
			expectFallback:           false,
		},
		{
			name: "Partial upstream usage - missing prompt tokens",
			setupFunc: func(ctx *StreamingContext) {
				ctx.usage = &model.Usage{
					CompletionTokens: 50,
					TotalTokens:      0, // Will be calculated
				}
			},
			promptTokens:             testPromptTokens,
			expectedPromptTokens:     testPromptTokens,
			expectedCompletionTokens: 50,
			expectedTotalTokens:      testPromptTokens + 50,
			expectFallback:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewStreamingContext(logger, false)
			if tt.setupFunc != nil {
				tt.setupFunc(ctx)
			}

			usage := ctx.CalculateUsage(tt.promptTokens, testModelName)

			if usage == nil {
				t.Fatal("Expected non-nil usage")
			}

			if usage.PromptTokens != tt.expectedPromptTokens {
				t.Errorf("Expected prompt tokens %d, got %d", tt.expectedPromptTokens, usage.PromptTokens)
			}
			if usage.CompletionTokens != tt.expectedCompletionTokens {
				t.Errorf("Expected completion tokens %d, got %d", tt.expectedCompletionTokens, usage.CompletionTokens)
			}
			if usage.TotalTokens != tt.expectedTotalTokens {
				t.Errorf("Expected total tokens %d, got %d", tt.expectedTotalTokens, usage.TotalTokens)
			}
		})
	}
}

// TestStreamingContext_ValidateStreamCompletion tests stream validation
func TestStreamingContext_ValidateStreamCompletion(t *testing.T) {
	logger := createTestLogger()

	tests := []struct {
		name        string
		setupFunc   func(*StreamingContext)
		expectValid bool
		expectError bool
	}{
		{
			name: "Valid stream - chunks processed",
			setupFunc: func(ctx *StreamingContext) {
				ctx.chunksProcessed = 5
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "Valid stream - content present",
			setupFunc: func(ctx *StreamingContext) {
				ctx.responseTextBuilder.WriteString("Hello world")
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "Valid stream - both chunks and content",
			setupFunc: func(ctx *StreamingContext) {
				ctx.chunksProcessed = 3
				ctx.responseTextBuilder.WriteString("Hello world")
			},
			expectValid: true,
			expectError: false,
		},
		{
			name:        "Invalid stream - no chunks or content",
			setupFunc:   nil,
			expectValid: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewStreamingContext(logger, false)
			if tt.setupFunc != nil {
				tt.setupFunc(ctx)
			}

			err, valid := ctx.ValidateStreamCompletion(testModelName, testContentType)

			if valid != tt.expectValid {
				t.Errorf("Expected valid %v, got %v", tt.expectValid, valid)
			}

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestStreamingContext_Integration tests complete streaming workflow
func TestStreamingContext_Integration(t *testing.T) {
	logger := createTestLogger()
	ctx := NewStreamingContext(logger, true) // Enable thinking

	// Simulate streaming workflow
	responses := []*ChatCompletionsStreamResponse{
		createTestStreamResponse("test-1", "Hello <think>", nil),
		createTestStreamResponse("test-2", "I need to think about this", nil),
		createTestStreamResponse("test-3", "</think> world!", nil),
		createTestStreamResponse("test-4", " How are you?", &model.Usage{
			PromptTokens:     testPromptTokens,
			CompletionTokens: 25,
			TotalTokens:      testPromptTokens + 25,
		}),
	}

	// Process all chunks
	for _, response := range responses {
		modified := ctx.ProcessStreamChunk(response)
		if !modified {
			t.Error("Expected all chunks to be modified")
		}
	}

	// Verify final state
	if ctx.chunksProcessed != 4 {
		t.Errorf("Expected 4 chunks processed, got %d", ctx.chunksProcessed)
	}

	// Calculate final usage
	finalUsage := ctx.CalculateUsage(testPromptTokens, testModelName)
	if finalUsage.PromptTokens != testPromptTokens {
		t.Errorf("Expected prompt tokens %d, got %d", testPromptTokens, finalUsage.PromptTokens)
	}

	// Validate completion
	err, valid := ctx.ValidateStreamCompletion(testModelName, testContentType)
	if !valid {
		t.Error("Expected valid completion")
	}
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify content accumulation in responseTextBuilder
	// Note: responseTextBuilder contains original content before thinking processing
	finalContent := ctx.responseTextBuilder.String()
	if !strings.Contains(finalContent, "Hello") || !strings.Contains(finalContent, "world!") {
		t.Error("Expected content to be preserved in responseTextBuilder")
	}
	// The thinking tags remain in responseTextBuilder as it stores original content
	// The actual thinking processing modifies the response.Choices[].Delta.Content fields
}

// TestBufferCapacityConstants tests the capacity constants
func TestBufferCapacityConstants(t *testing.T) {
	if DefaultBuilderCapacity != 4096 {
		t.Errorf("Expected DefaultBuilderCapacity to be 4096, got %d", DefaultBuilderCapacity)
	}
	if LargeBuilderCapacity != 65536 {
		t.Errorf("Expected LargeBuilderCapacity to be 65536, got %d", LargeBuilderCapacity)
	}
	if MaxBuilderCapacity != 1048576 {
		t.Errorf("Expected MaxBuilderCapacity to be 1048576, got %d", MaxBuilderCapacity)
	}

	// Verify logical ordering
	if DefaultBuilderCapacity >= LargeBuilderCapacity {
		t.Error("DefaultBuilderCapacity should be less than LargeBuilderCapacity")
	}
	if LargeBuilderCapacity >= MaxBuilderCapacity {
		t.Error("LargeBuilderCapacity should be less than MaxBuilderCapacity")
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// TestUnifiedStreamProcessing_ThinkingMapping verifies that when thinking is enabled and reasoning_format
// is set, the streamed chunks contain the reasoning in the requested field and avoid duplicates.
func TestUnifiedStreamProcessing_ThinkingMapping(t *testing.T) {
	// Prepare a gin test context with query params
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest("POST", "/v1/chat/completions?thinking=true&reasoning_format=thinking", nil)
	c.Request = req

	// Build a single SSE chunk where delta content includes a think block
	chunk := ChatCompletionsStreamResponse{
		Id:      "test-id",
		Object:  "chat.completion.chunk",
		Created: 123,
		Model:   testModelName,
		Choices: []ChatCompletionsStreamResponseChoice{
			{
				Index: 0,
				Delta: model.Message{Content: "hello <think>abc</think> world"},
			},
		},
	}
	b, _ := json.Marshal(chunk)
	sse := "data: " + string(b) + "\n\n" + "data: [DONE]\n"

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(sse)),
	}

	// Run unified processing with thinking enabled
	if err, _ := UnifiedStreamProcessing(c, resp, 0, testModelName, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse the first emitted chunk from the recorder
	body := w.Body.String()
	// Extract the JSON after "data: "
	lines := strings.Split(body, "\n")
	var jsonLine string
	for _, ln := range lines {
		if strings.HasPrefix(ln, "data: {") {
			jsonLine = strings.TrimPrefix(ln, "data: ")
			break
		}
	}
	if jsonLine == "" {
		t.Fatalf("no JSON chunk found in response body: %q", body)
	}

	var out ChatCompletionsStreamResponse
	if err := json.Unmarshal([]byte(jsonLine), &out); err != nil {
		t.Fatalf("failed to unmarshal emitted chunk: %v", err)
	}
	if len(out.Choices) == 0 {
		t.Fatalf("no choices in emitted chunk: %v", out)
	}
	got := out.Choices[0].Delta
	// Expect content without think tags
	if got.StringContent() != "hello  world" {
		t.Fatalf("unexpected content: %q", got.StringContent())
	}
	// Expect thinking field to be set as per reasoning_format=thinking
	if got.Thinking == nil || *got.Thinking != "abc" {
		t.Fatalf("expected thinking=abc, got %#v", got.Thinking)
	}
	// And ReasoningContent should be cleared to avoid duplicates
	if got.ReasoningContent != nil {
		t.Fatalf("expected ReasoningContent to be nil when reasoning_format=thinking; got %#v", *got.ReasoningContent)
	}
}
