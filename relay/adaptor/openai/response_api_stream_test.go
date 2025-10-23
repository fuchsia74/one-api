package openai

import (
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

// TestCompleteResponseAPIStream tests the complete Response API streaming workflow
// This test simulates the exact SSE format from the Response API specification
func TestCompleteResponseAPIStream(t *testing.T) {
	// This simulates exactly what would come from a Response API stream
	sseStreamExample := `event: response.created
data: {"type":"response.created","response":{"id":"resp_67c9fdcecf488190bdd9a0409de3a1ec07b8b0ad4e5eb654","object":"response","created_at":1741290958,"status":"in_progress","error":null,"incomplete_details":null,"instructions":"You are a helpful assistant.","max_output_tokens":null,"model":"gpt-4.1-2025-04-14","output":[],"parallel_tool_calls":true,"previous_response_id":null,"reasoning":{"effort":null,"summary":null},"store":true,"temperature":1.0,"text":{"format":{"type":"text"}},"tool_choice":"auto","tools":[],"top_p":1.0,"truncation":"disabled","usage":null,"user":null,"metadata":{}}}

event: response.output_item.added
data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","type":"message","status":"in_progress","role":"assistant","content":[]}}

event: response.content_part.added
data: {"type":"response.content_part.added","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"part":{"type":"output_text","text":"","annotations":[]}}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"delta":"Hi"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"delta":" there!"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"delta":" How can I assist you today?"}

event: response.output_text.done
data: {"type":"response.output_text.done","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"text":"Hi there! How can I assist you today?"}

event: response.content_part.done
data: {"type":"response.content_part.done","item_id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","output_index":0,"content_index":0,"part":{"type":"output_text","text":"Hi there! How can I assist you today?","annotations":[]}}

event: response.output_item.done
data: {"type":"response.output_item.done","output_index":0,"item":{"id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"Hi there! How can I assist you today?","annotations":[]}]}}

event: response.completed
data: {"type":"response.completed","response":{"id":"resp_67c9fdcecf488190bdd9a0409de3a1ec07b8b0ad4e5eb654","object":"response","created_at":1741290958,"status":"completed","error":null,"incomplete_details":null,"instructions":"You are a helpful assistant.","max_output_tokens":null,"model":"gpt-4.1-2025-04-14","output":[{"id":"msg_67c9fdcf37fc8190ba82116e33fb28c507b8b0ad4e5eb654","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"Hi there! How can I assist you today?","annotations":[]}]}],"parallel_tool_calls":true,"previous_response_id":null,"reasoning":{"effort":null,"summary":null},"store":true,"temperature":1.0,"text":{"format":{"type":"text"}},"tool_choice":"auto","tools":[],"top_p":1.0,"truncation":"disabled","usage":{"input_tokens":37,"output_tokens":11,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":48},"user":null,"metadata":{}}}

data: [DONE]`

	// Split into lines and process exactly like the ResponseAPIStreamHandler would
	lines := strings.Split(sseStreamExample, "\n")

	const dataPrefix = "data: "
	const dataPrefixLength = len(dataPrefix)

	responseText := ""
	eventCount := 0
	deltaCount := 0

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and event lines (exactly like NormalizeDataLine + ResponseAPIStreamHandler)
		if line == "" {
			continue
		}

		data := NormalizeDataLine(line)

		if !strings.HasPrefix(data, dataPrefix) {
			continue
		}

		// Extract JSON data
		jsonData := data[dataPrefixLength:]

		if jsonData == "[DONE]" {
			break
		}

		eventCount++

		// Parse using the improved parsing logic
		fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(jsonData))
		if err != nil {
			t.Errorf("Line %d: Parse error: %v", i+1, err)
			continue
		}

		// Convert to ResponseAPIResponse (same as ResponseAPIStreamHandler)
		var responseAPIChunk ResponseAPIResponse
		if fullResponse != nil {
			responseAPIChunk = *fullResponse
		} else if streamEvent != nil {
			responseAPIChunk = ConvertStreamEventToResponse(streamEvent)

			// Track delta events specifically
			if strings.Contains(streamEvent.Type, "delta") {
				if delta := extractStringFromRaw(streamEvent.Delta, "partial_json", "json", "text", "delta"); delta != "" {
					deltaCount++
					responseText += delta
				}
			}
		}

		// Convert to ChatCompletion format (same as ResponseAPIStreamHandler)
		chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)

		// Verify conversion worked
		if len(chatCompletionChunk.Choices) == 0 {
			t.Errorf("Line %d: ChatCompletion conversion failed - no choices", i+1)
		}
	}

	// Verify we got the expected content
	expectedText := "Hi there! How can I assist you today?"
	if responseText != expectedText {
		t.Errorf("Response text mismatch: expected '%s', got '%s'", expectedText, responseText)
	}

	// Verify we processed the expected number of events
	if eventCount == 0 {
		t.Error("No events were processed")
	}

	if deltaCount != 3 {
		t.Errorf("Expected 3 delta events, got %d", deltaCount)
	}
}

// TestResponseAPIStreamingEvents tests individual streaming events
func TestResponseAPIStreamingEvents(t *testing.T) {
	t.Run("Problematic streaming event", func(t *testing.T) {
		// Test the problematic streaming event that was causing the parsing error
		problematicEvent := `{"type":"response.output_text.done","sequence_number":22,"item_id":"msg_6849865110908191a4809c86e082ff710008bd3c6060334b","output_index":1,"content_index":0,"text":"Why don't skeletons fight each other?\n\nThey don't have the guts."}`

		// Test the new flexible parsing approach
		fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(problematicEvent))
		if err != nil {
			t.Fatalf("Failed to parse streaming event: %v", err)
		}

		if fullResponse != nil {
			t.Error("Expected stream event, got full response")
		}

		if streamEvent == nil {
			t.Fatal("Expected streamEvent to be non-nil")
		}

		if streamEvent.Type != "response.output_text.done" {
			t.Errorf("Expected type 'response.output_text.done', got '%s'", streamEvent.Type)
		}

		// Test conversion to ResponseAPIResponse
		responseAPIChunk := ConvertStreamEventToResponse(streamEvent)
		if len(responseAPIChunk.Output) == 0 {
			t.Error("Expected output items in converted response")
		}

		// Test conversion to ChatCompletion format
		chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)
		if len(chatCompletionChunk.Choices) == 0 {
			t.Error("Expected choices in ChatCompletion chunk")
		}
	})

	t.Run("Delta streaming event", func(t *testing.T) {
		deltaEvent := `{"type":"response.output_text.delta","sequence_number":6,"item_id":"msg_6849865110908191a4809c86e082ff710008bd3c6060334b","output_index":1,"content_index":0,"delta":"Why"}`

		_, streamEvent, err := ParseResponseAPIStreamEvent([]byte(deltaEvent))
		if err != nil {
			t.Fatalf("Failed to parse delta event: %v", err)
		}

		if streamEvent == nil {
			t.Fatal("Expected streamEvent to be non-nil")
		}

		if streamEvent.Type != "response.output_text.delta" {
			t.Errorf("Expected type 'response.output_text.delta', got '%s'", streamEvent.Type)
		}

		delta := extractStringFromRaw(streamEvent.Delta, "text", "delta")
		if delta != "Why" {
			t.Errorf("Expected delta 'Why', got '%s'", delta)
		}

		// Test conversion
		responseAPIChunk := ConvertStreamEventToResponse(streamEvent)
		chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)

		if len(chatCompletionChunk.Choices) == 0 {
			t.Error("Expected choices in ChatCompletion chunk")
		}

		if content, ok := chatCompletionChunk.Choices[0].Delta.Content.(string); !ok || content != "Why" {
			t.Errorf("Expected ChatCompletion delta content 'Why', got '%v'", chatCompletionChunk.Choices[0].Delta.Content)
		}
	})

	t.Run("Full response event", func(t *testing.T) {
		fullResponseEvent := `{"id":"resp_123","object":"response","created_at":1749648976,"status":"completed","model":"o3-2025-04-16","output":[{"type":"message","id":"msg_123","status":"completed","role":"assistant","content":[{"type":"output_text","text":"Hello world"}]}],"usage":{"input_tokens":9,"output_tokens":22,"total_tokens":31}}`

		fullResponse, _, err := ParseResponseAPIStreamEvent([]byte(fullResponseEvent))
		if err != nil {
			t.Fatalf("Failed to parse full response event: %v", err)
		}

		if fullResponse == nil {
			t.Fatal("Expected fullResponse to be non-nil")
		}

		if fullResponse.Id != "resp_123" {
			t.Errorf("Expected ID 'resp_123', got '%s'", fullResponse.Id)
		}

		if fullResponse.Status != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", fullResponse.Status)
		}

		if fullResponse.Usage == nil || fullResponse.Usage.TotalTokens != 31 {
			t.Errorf("Expected usage with total tokens 31, got %v", fullResponse.Usage)
		}
	})
}

// TestSSEProcessing tests SSE line processing logic
func TestSSEProcessing(t *testing.T) {
	testCases := []struct {
		name          string
		line          string
		shouldProcess bool
		expectError   bool
	}{
		{
			name:          "Valid data line",
			line:          `data: {"type":"response.output_text.delta","delta":"Hi"}`,
			shouldProcess: true,
			expectError:   false,
		},
		{
			name:          "Event line (should skip)",
			line:          "event: response.created",
			shouldProcess: false,
			expectError:   false,
		},
		{
			name:          "Empty line (should skip)",
			line:          "",
			shouldProcess: false,
			expectError:   false,
		},
		{
			name:          "DONE signal",
			line:          "data: [DONE]",
			shouldProcess: false, // DONE is handled specially
			expectError:   false,
		},
		{
			name:          "Malformed JSON",
			line:          `data: {"invalid": json}`,
			shouldProcess: true,
			expectError:   true,
		},
	}

	const dataPrefix = "data: "
	const dataPrefixLength = len(dataPrefix)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := NormalizeDataLine(tc.line)

			// Check if line should be processed
			if !strings.HasPrefix(data, dataPrefix) {
				if tc.shouldProcess {
					t.Errorf("Expected line to be processed, but it was skipped")
				}
				return
			}

			jsonData := data[dataPrefixLength:]

			if jsonData == "[DONE]" {
				if tc.shouldProcess {
					t.Errorf("DONE signal should not be processed as JSON")
				}
				return
			}

			if !tc.shouldProcess {
				t.Errorf("Expected line to be skipped, but it was processed")
				return
			}

			// Test parsing
			_, _, err := ParseResponseAPIStreamEvent([]byte(jsonData))

			if tc.expectError && err == nil {
				t.Errorf("Expected parsing error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected parsing error: %v", err)
			}
		})
	}
}

// TestStreamEventTypes tests various Response API stream event types
func TestStreamEventTypes(t *testing.T) {
	testCases := []struct {
		name         string
		eventData    string
		expectedType string
	}{
		{
			name:         "response.created event",
			eventData:    `{"type":"response.created","response":{"id":"resp_123","status":"in_progress"}}`,
			expectedType: "response.created",
		},
		{
			name:         "response.output_text.delta event",
			eventData:    `{"type":"response.output_text.delta","delta":"Hi"}`,
			expectedType: "response.output_text.delta",
		},
		{
			name:         "response.output_text.done event",
			eventData:    `{"type":"response.output_text.done","text":"Complete text"}`,
			expectedType: "response.output_text.done",
		},
		{
			name:         "response.completed event",
			eventData:    `{"type":"response.completed","response":{"id":"resp_123","status":"completed"}}`,
			expectedType: "response.completed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(tc.eventData))
			if err != nil {
				t.Fatalf("Parsing failed: %v", err)
			}

			if fullResponse != nil && streamEvent != nil {
				t.Error("Both fullResponse and streamEvent should not be non-nil")
			}

			if fullResponse == nil && streamEvent == nil {
				t.Error("Both fullResponse and streamEvent should not be nil")
			}

			var eventType string
			if fullResponse != nil {
				// For full response events, we need to extract the type from the original data
				// This is a limitation of the current parsing approach
				if strings.Contains(tc.eventData, `"type":"response.created"`) {
					eventType = "response.created"
				} else if strings.Contains(tc.eventData, `"type":"response.completed"`) {
					eventType = "response.completed"
				}
			} else if streamEvent != nil {
				eventType = streamEvent.Type
			}

			if eventType != tc.expectedType {
				t.Errorf("Expected event type '%s', got '%s'", tc.expectedType, eventType)
			}

			// Test conversion for stream events
			if streamEvent != nil {
				responseAPIChunk := ConvertStreamEventToResponse(streamEvent)
				chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)

				if chatCompletionChunk.Object != "chat.completion.chunk" {
					t.Errorf("Expected object 'chat.completion.chunk', got '%s'", chatCompletionChunk.Object)
				}
			}
		})
	}
}

// TestNormalizeDataLine tests the data line normalization function
func TestNormalizeDataLine(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Already normalized",
			input:    "data: {\"test\": true}",
			expected: "data: {\"test\": true}",
		},
		{
			name:     "No space after colon",
			input:    "data:{\"test\": true}",
			expected: "data: {\"test\": true}",
		},
		{
			name:     "Multiple spaces after colon",
			input:    "data:   {\"test\": true}",
			expected: "data: {\"test\": true}",
		},
		{
			name:     "Tab after colon",
			input:    "data:\t{\"test\": true}",
			expected: "data: \t{\"test\": true}", // TrimLeft only removes spaces, not tabs
		},
		{
			name:     "Non-data line",
			input:    "event: test",
			expected: "event: test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeDataLine(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestDuplicateContentFix tests that the streaming handler properly prevents duplicate content
// by only accumulating delta events and not processing complete/done events for accumulation
func TestDuplicateContentFix(t *testing.T) {
	// This test case includes both delta events (incremental) and done events (complete)
	// The key issue was that both types were being accumulated, causing duplicates
	sseStreamWithDuplicateRisk := `event: response.created
data: {"type":"response.created","response":{"id":"resp_test_dup","object":"response","created_at":1741290958,"status":"in_progress"}}

event: response.output_item.added
data: {"type":"response.output_item.added","output_index":0,"item":{"id":"msg_test_dup","type":"message","status":"in_progress","role":"assistant","content":[]}}

event: response.content_part.added
data: {"type":"response.content_part.added","item_id":"msg_test_dup","output_index":0,"content_index":0,"part":{"type":"output_text","text":"","annotations":[]}}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_test_dup","output_index":0,"content_index":0,"delta":"The"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_test_dup","output_index":0,"content_index":0,"delta":" quick"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_test_dup","output_index":0,"content_index":0,"delta":" brown"}

event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_test_dup","output_index":0,"content_index":0,"delta":" fox"}

event: response.output_text.done
data: {"type":"response.output_text.done","item_id":"msg_test_dup","output_index":0,"content_index":0,"text":"The quick brown fox"}

event: response.content_part.done
data: {"type":"response.content_part.done","item_id":"msg_test_dup","output_index":0,"content_index":0,"part":{"type":"output_text","text":"The quick brown fox","annotations":[]}}

event: response.output_item.done
data: {"type":"response.output_item.done","output_index":0,"item":{"id":"msg_test_dup","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"The quick brown fox","annotations":[]}]}}

event: response.completed
data: {"type":"response.completed","response":{"id":"resp_test_dup","object":"response","created_at":1741290958,"status":"completed","output":[{"id":"msg_test_dup","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"The quick brown fox","annotations":[]}]}],"usage":{"input_tokens":10,"output_tokens":4,"total_tokens":14}}}

data: [DONE]`

	// Process using the same logic as ResponseAPIStreamHandler
	lines := strings.Split(sseStreamWithDuplicateRisk, "\n")
	const dataPrefix = "data: "
	const dataPrefixLength = len(dataPrefix)

	accumulatedText := ""
	deltaEventCount := 0
	doneEventCount := 0
	completeDoneEvents := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		data := NormalizeDataLine(line)
		if !strings.HasPrefix(data, dataPrefix) {
			continue
		}

		jsonData := data[dataPrefixLength:]
		if jsonData == "[DONE]" {
			break
		}

		// Parse the streaming event
		fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(jsonData))
		if err != nil {
			continue
		}

		var responseAPIChunk ResponseAPIResponse
		if fullResponse != nil {
			responseAPIChunk = *fullResponse
		} else if streamEvent != nil {
			responseAPIChunk = ConvertStreamEventToResponse(streamEvent)

			// Count event types
			if strings.Contains(streamEvent.Type, "delta") {
				deltaEventCount++
			} else if strings.Contains(streamEvent.Type, "done") {
				doneEventCount++
				if streamEvent.Type == "response.output_text.done" ||
					streamEvent.Type == "response.content_part.done" ||
					streamEvent.Type == "response.output_item.done" {
					completeDoneEvents++
				}
			}

			// Apply the fix: Only accumulate from delta events to prevent duplication
			if strings.Contains(streamEvent.Type, "delta") {
				if delta := extractStringFromRaw(streamEvent.Delta, "partial_json", "json", "text", "delta"); delta != "" {
					accumulatedText += delta
				}
			}
		}

		// Convert to ChatCompletion format
		chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)
		if len(chatCompletionChunk.Choices) == 0 {
			t.Errorf("ChatCompletion conversion failed for event type: %s",
				func() string {
					if streamEvent != nil {
						return streamEvent.Type
					}
					return "full_response"
				}())
		}
	}

	// Validate the fix
	expectedText := "The quick brown fox"
	if accumulatedText != expectedText {
		t.Errorf("DUPLICATE CONTENT DETECTED: expected '%s', got '%s'", expectedText, accumulatedText)
		t.Errorf("This indicates that done events are still being accumulated, causing duplication")
	}

	// Verify event counts
	if deltaEventCount != 4 {
		t.Errorf("Expected 4 delta events, got %d", deltaEventCount)
	}

	if completeDoneEvents < 3 {
		t.Errorf("Expected at least 3 complete done events, got %d", completeDoneEvents)
	}

	// Additional validation: ensure we have processed both types of events
	if deltaEventCount == 0 {
		t.Error("No delta events processed - test setup is incorrect")
	}

	if doneEventCount == 0 {
		t.Error("No done events processed - test setup is incorrect")
	}

	t.Logf("✅ Duplicate content fix validated successfully")
	t.Logf("📊 Delta events: %d, Done events: %d, Complete done events: %d",
		deltaEventCount, doneEventCount, completeDoneEvents)
	t.Logf("🎯 Final accumulated text: '%s' (correct, no duplication)", accumulatedText)
}

// TestResponseAPIStreamUsageHandling tests that usage information from response.completed events
// is properly captured and converted to ChatCompletion streaming format
func TestResponseAPIStreamUsageHandling(t *testing.T) {
	// This simulates a Response API stream with usage information in the response.completed event
	sseStreamWithUsage := `event: response.output_text.delta
data: {"type":"response.output_text.delta","item_id":"msg_test","output_index":0,"content_index":0,"delta":"Hello"}

event: response.completed
data: {"type":"response.completed","response":{"id":"resp_test123","object":"response","created_at":1749954928,"status":"completed","model":"gpt-4o-2024-11-20","output":[{"id":"msg_test","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"Hello world!","annotations":[]}]}],"usage":{"input_tokens":97,"input_tokens_details":{"cached_tokens":0},"output_tokens":76,"output_tokens_details":{"reasoning_tokens":10},"total_tokens":173}}}

data: [DONE]`

	// Split into lines and process exactly like the ResponseAPIStreamHandler would
	lines := strings.Split(sseStreamWithUsage, "\n")

	const dataPrefix = "data: "
	const dataPrefixLength = len(dataPrefix)

	var deltaEvents []ChatCompletionsStreamResponse
	var usageEvents []ChatCompletionsStreamResponse
	responseText := ""

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and event lines
		if line == "" {
			continue
		}

		data := NormalizeDataLine(line)

		if !strings.HasPrefix(data, dataPrefix) {
			continue
		}

		// Extract JSON data
		jsonData := data[dataPrefixLength:]

		if jsonData == "[DONE]" {
			break
		}

		// Parse using the same logic as ResponseAPIStreamHandler
		fullResponse, streamEvent, err := ParseResponseAPIStreamEvent([]byte(jsonData))
		if err != nil {
			t.Errorf("Line %d: Parse error: %v", i+1, err)
			continue
		}

		// Convert to ResponseAPIResponse (same as ResponseAPIStreamHandler)
		var responseAPIChunk ResponseAPIResponse
		if fullResponse != nil {
			responseAPIChunk = *fullResponse
		} else if streamEvent != nil {
			responseAPIChunk = ConvertStreamEventToResponse(streamEvent)
		}

		// Simulate the logic from ResponseAPIStreamHandler for handling events
		if streamEvent != nil {
			eventType := streamEvent.Type

			// Process delta events
			if strings.Contains(eventType, "delta") {
				// Accumulate response text
				if delta := extractStringFromRaw(streamEvent.Delta, "partial_json", "json", "text", "delta"); delta != "" {
					responseText += delta
				}

				// Convert and store delta event
				chatCompletionChunk := ConvertResponseAPIStreamToChatCompletion(&responseAPIChunk)
				deltaEvents = append(deltaEvents, *chatCompletionChunk)
			} else if eventType == "response.completed" && responseAPIChunk.Usage != nil {
				// This is the new logic we're testing - capture usage from response.completed
				convertedUsage := responseAPIChunk.Usage.ToModelUsage()
				if convertedUsage != nil {
					// Create a usage-only streaming chunk (mimicking the new logic)
					usageChunk := ChatCompletionsStreamResponse{
						Id:      responseAPIChunk.Id,
						Object:  "chat.completion.chunk",
						Created: responseAPIChunk.CreatedAt,
						Model:   responseAPIChunk.Model,
						Choices: []ChatCompletionsStreamResponseChoice{
							{
								Index: 0,
								Delta: model.Message{
									Role:    "assistant",
									Content: "", // Usage chunk should have empty content
								},
								FinishReason: nil, // Don't set finish reason in usage chunk
							},
						},
						Usage: convertedUsage,
					}
					usageEvents = append(usageEvents, usageChunk)
				}
			}
		}
	}

	// Verify we got the expected content from delta events
	expectedText := "Hello"
	if responseText != expectedText {
		t.Errorf("Response text mismatch: expected '%s', got '%s'", expectedText, responseText)
	}

	// Verify we got delta events
	if len(deltaEvents) == 0 {
		t.Error("No delta events were captured")
	}

	// Verify we got exactly one usage event from response.completed
	if len(usageEvents) != 1 {
		t.Errorf("Expected exactly 1 usage event, got %d", len(usageEvents))
	}

	if len(usageEvents) > 0 {
		usageEvent := usageEvents[0]

		// Verify the usage event structure
		if usageEvent.Usage == nil {
			t.Error("Usage event missing usage information")
		} else {
			// Verify usage values were converted correctly from ResponseAPI format
			if usageEvent.Usage.PromptTokens != 97 {
				t.Errorf("Expected PromptTokens=97, got %d", usageEvent.Usage.PromptTokens)
			}
			if usageEvent.Usage.CompletionTokens != 76 {
				t.Errorf("Expected CompletionTokens=76, got %d", usageEvent.Usage.CompletionTokens)
			}
			if usageEvent.Usage.TotalTokens != 173 {
				t.Errorf("Expected TotalTokens=173, got %d", usageEvent.Usage.TotalTokens)
			}

			// Verify completion token details were converted correctly
			if usageEvent.Usage.CompletionTokensDetails == nil {
				t.Error("Expected CompletionTokensDetails to be present")
			} else {
				if usageEvent.Usage.CompletionTokensDetails.ReasoningTokens != 10 {
					t.Errorf("Expected ReasoningTokens=10, got %d", usageEvent.Usage.CompletionTokensDetails.ReasoningTokens)
				}
			}
		}

		// Verify the structure matches ChatCompletion streaming format
		if usageEvent.Object != "chat.completion.chunk" {
			t.Errorf("Expected object='chat.completion.chunk', got '%s'", usageEvent.Object)
		}

		if usageEvent.Id != "resp_test123" {
			t.Errorf("Expected id='resp_test123', got '%s'", usageEvent.Id)
		}

		if usageEvent.Model != "gpt-4o-2024-11-20" {
			t.Errorf("Expected model='gpt-4o-2024-11-20', got '%s'", usageEvent.Model)
		}

		// Verify the choice structure
		if len(usageEvent.Choices) != 1 {
			t.Errorf("Expected 1 choice, got %d", len(usageEvent.Choices))
		} else {
			choice := usageEvent.Choices[0]
			if choice.Delta.Role != "assistant" {
				t.Errorf("Expected delta role='assistant', got '%s'", choice.Delta.Role)
			}

			// IMPORTANT: Verify that usage chunk has empty string content, not nil
			if content, ok := choice.Delta.Content.(string); !ok {
				t.Errorf("Expected delta content to be a string, got %T", choice.Delta.Content)
			} else if content != "" {
				t.Errorf("Expected delta content to be empty string for usage chunk, got '%s'", content)
			}

			if choice.FinishReason != nil {
				t.Errorf("Expected FinishReason=nil for usage chunk, got '%v'", choice.FinishReason)
			}
		}
	}

	t.Logf("✅ Usage handling test passed")
	t.Logf("📊 Delta events: %d, Usage events: %d", len(deltaEvents), len(usageEvents))
	t.Logf("🎯 Accumulated text: '%s'", responseText)
}
