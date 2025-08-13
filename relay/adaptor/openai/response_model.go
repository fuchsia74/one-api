package openai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/songquanpeng/one-api/relay/model"
)

// ResponseAPIInput represents the input field that can be either a string or an array
type ResponseAPIInput []any

// UnmarshalJSON implements custom unmarshaling for ResponseAPIInput
// to handle both string and array inputs as per OpenAI Response API specification
func (r *ResponseAPIInput) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*r = ResponseAPIInput{str}
		return nil
	}

	// If string unmarshaling fails, try as array
	var arr []any
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*r = ResponseAPIInput(arr)
	return nil
}

// MarshalJSON implements custom marshaling for ResponseAPIInput
// If the input contains only one string element, marshal as string
// Otherwise, marshal as array
func (r ResponseAPIInput) MarshalJSON() ([]byte, error) {
	// If there's exactly one element and it's a string, marshal as string
	if len(r) == 1 {
		if str, ok := r[0].(string); ok {
			return json.Marshal(str)
		}
	}
	// Otherwise, marshal as array
	return json.Marshal([]any(r))
}

// IsModelsOnlySupportedByChatCompletionAPI determines if a model only supports ChatCompletion API
// and should not be converted to Response API format.
// Currently returns false for all models (allowing conversion), but can be implemented later
// to return true for specific models that only support ChatCompletion API.
func IsModelsOnlySupportedByChatCompletionAPI(actualModel string) bool {
	switch {
	case strings.Contains(actualModel, "gpt") && strings.Contains(actualModel, "-search-"),
		strings.Contains(actualModel, "gpt") && strings.Contains(actualModel, "-audio-"):
		return true
	default:
		return false
	}
}

// ResponseAPIRequest represents the OpenAI Response API request structure
// https://platform.openai.com/docs/api-reference/responses
type ResponseAPIRequest struct {
	Input              ResponseAPIInput               `json:"input,omitempty"`                // Optional: Text, image, or file inputs to the model (string or array) - mutually exclusive with prompt
	Model              string                         `json:"model"`                          // Required: Model ID used to generate the response
	Background         *bool                          `json:"background,omitempty"`           // Optional: Whether to run the model response in the background
	Include            []string                       `json:"include,omitempty"`              // Optional: Additional output data to include
	Instructions       *string                        `json:"instructions,omitempty"`         // Optional: System message as the first item in the model's context
	MaxOutputTokens    *int                           `json:"max_output_tokens,omitempty"`    // Optional: Upper bound for the number of tokens
	Metadata           any                            `json:"metadata,omitempty"`             // Optional: Set of 16 key-value pairs
	ParallelToolCalls  *bool                          `json:"parallel_tool_calls,omitempty"`  // Optional: Whether to allow the model to run tool calls in parallel
	PreviousResponseId *string                        `json:"previous_response_id,omitempty"` // Optional: The unique ID of the previous response
	Prompt             *ResponseAPIPrompt             `json:"prompt,omitempty"`               // Optional: Prompt template configuration - mutually exclusive with input
	Reasoning          *model.OpenAIResponseReasoning `json:"reasoning,omitempty"`            // Optional: Configuration options for reasoning models
	ServiceTier        *string                        `json:"service_tier,omitempty"`         // Optional: Latency tier to use for processing
	Store              *bool                          `json:"store,omitempty"`                // Optional: Whether to store the generated model response
	Stream             *bool                          `json:"stream,omitempty"`               // Optional: If set to true, model response data will be streamed
	Temperature        *float64                       `json:"temperature,omitempty"`          // Optional: Sampling temperature
	Text               *ResponseTextConfig            `json:"text,omitempty"`                 // Optional: Configuration options for a text response
	ToolChoice         any                            `json:"tool_choice,omitempty"`          // Optional: How the model should select tools
	Tools              []ResponseAPITool              `json:"tools,omitempty"`                // Optional: Array of tools the model may call
	TopP               *float64                       `json:"top_p,omitempty"`                // Optional: Alternative to sampling with temperature
	Truncation         *string                        `json:"truncation,omitempty"`           // Optional: Truncation strategy
	User               *string                        `json:"user,omitempty"`                 // Optional: Stable identifier for end-users
}

// ResponseAPIPrompt represents the prompt template configuration for Response API requests
type ResponseAPIPrompt struct {
	Id        string                 `json:"id"`                  // Required: Unique identifier of the prompt template
	Version   *string                `json:"version,omitempty"`   // Optional: Specific version of the prompt (defaults to "current")
	Variables map[string]interface{} `json:"variables,omitempty"` // Optional: Map of values to substitute in for variables in the prompt
}

// ResponseAPITool represents the tool format for Response API requests
// This differs from the ChatCompletion tool format where function properties are nested
type ResponseAPITool struct {
	Type        string                 `json:"type"`                  // Required: "function"
	Name        string                 `json:"name"`                  // Required: Function name
	Description string                 `json:"description,omitempty"` // Optional: Function description
	Parameters  map[string]interface{} `json:"parameters,omitempty"`  // Optional: Function parameters schema
}

// ResponseTextConfig represents the text configuration for Response API
type ResponseTextConfig struct {
	Format *ResponseTextFormat `json:"format,omitempty"` // Optional: Format configuration for structured outputs
}

// ResponseTextFormat represents the format configuration for Response API structured outputs
type ResponseTextFormat struct {
	Type        string                 `json:"type"`                  // Required: Format type (e.g., "text", "json_schema")
	Name        string                 `json:"name,omitempty"`        // Optional: Schema name for json_schema type
	Description string                 `json:"description,omitempty"` // Optional: Schema description
	Schema      map[string]interface{} `json:"schema,omitempty"`      // Optional: JSON schema definition
	Strict      *bool                  `json:"strict,omitempty"`      // Optional: Whether to use strict mode
}

// convertResponseAPIIDToToolCall converts Response API function call IDs back to ChatCompletion format
// Removes the "fc_" and "call_" prefixes to get the original ID
func convertResponseAPIIDToToolCall(fcID, callID string) string {
	if fcID != "" && strings.HasPrefix(fcID, "fc_") {
		return strings.TrimPrefix(fcID, "fc_")
	}
	if callID != "" && strings.HasPrefix(callID, "call_") {
		return strings.TrimPrefix(callID, "call_")
	}
	// Fallback to using the ID as-is
	if fcID != "" {
		return fcID
	}
	return callID
}

// convertToolCallIDToResponseAPI converts a ChatCompletion tool call ID to Response API format
// The Response API expects IDs with "fc_" prefix for function calls and "call_" prefix for call_id
func convertToolCallIDToResponseAPI(originalID string) (fcID, callID string) {
	if originalID == "" {
		return "", ""
	}

	// If the ID already has the correct prefix, use it as-is
	if strings.HasPrefix(originalID, "fc_") {
		return originalID, strings.Replace(originalID, "fc_", "call_", 1)
	}
	if strings.HasPrefix(originalID, "call_") {
		return strings.Replace(originalID, "call_", "fc_", 1), originalID
	}

	// Otherwise, generate appropriate prefixes
	return "fc_" + originalID, "call_" + originalID
}

// findToolCallName finds the function name for a given tool call ID
func findToolCallName(toolCalls []model.Tool, toolCallId string) string {
	for _, toolCall := range toolCalls {
		if toolCall.Id == toolCallId {
			return toolCall.Function.Name
		}
	}
	return "unknown_function"
}

// convertMessageToResponseAPIFormat converts a ChatCompletion message to Response API format
// This function handles the content type conversion from ChatCompletion format to Response API format
func convertMessageToResponseAPIFormat(message model.Message) map[string]interface{} {
	responseMsg := map[string]interface{}{
		"role": message.Role,
	}

	// Determine the appropriate content type based on message role
	// For Response API: user messages use "input_text", assistant messages use "output_text"
	textContentType := "input_text"
	if message.Role == "assistant" {
		textContentType = "output_text"
	}

	// Handle different content types
	switch content := message.Content.(type) {
	case string:
		// Simple string content - convert to appropriate text format based on role
		if content != "" {
			responseMsg["content"] = []map[string]interface{}{
				{
					"type": textContentType,
					"text": content,
				},
			}
		}
	case []model.MessageContent:
		// Structured content - convert each part to Response API format
		var convertedContent []map[string]interface{}
		for _, part := range content {
			switch part.Type {
			case model.ContentTypeText:
				if part.Text != nil && *part.Text != "" {
					convertedContent = append(convertedContent, map[string]interface{}{
						"type": textContentType,
						"text": *part.Text,
					})
				}
			case model.ContentTypeImageURL:
				if part.ImageURL != nil && part.ImageURL.Url != "" {
					convertedContent = append(convertedContent, map[string]interface{}{
						"type":      "input_image",
						"image_url": part.ImageURL.Url,
					})
				}
			case model.ContentTypeInputAudio:
				if part.InputAudio != nil {
					convertedContent = append(convertedContent, map[string]interface{}{
						"type":        "input_audio",
						"input_audio": part.InputAudio,
					})
				}
			default:
				// For unknown types, try to preserve as much as possible
				partMap := map[string]interface{}{
					"type": textContentType, // Use appropriate text type based on role
				}
				if part.Text != nil {
					partMap["text"] = *part.Text
				}
				convertedContent = append(convertedContent, partMap)
			}
		}
		if len(convertedContent) > 0 {
			responseMsg["content"] = convertedContent
		}
	case []interface{}:
		// Handle generic interface array (from JSON unmarshaling)
		var convertedContent []map[string]interface{}
		for _, item := range content {
			if itemMap, ok := item.(map[string]interface{}); ok {
				convertedItem := make(map[string]interface{})
				for k, v := range itemMap {
					convertedItem[k] = v
				}
				// Convert content types to Response API format based on message role
				if itemType, exists := itemMap["type"]; exists {
					switch itemType {
					case "text":
						convertedItem["type"] = textContentType
					case "image_url":
						convertedItem["type"] = "input_image"
					}
				}
				convertedContent = append(convertedContent, convertedItem)
			}
		}
		if len(convertedContent) > 0 {
			responseMsg["content"] = convertedContent
		}
	default:
		// Fallback: convert to string and treat as appropriate text type based on role
		if contentStr := fmt.Sprintf("%v", content); contentStr != "" && contentStr != "<nil>" {
			responseMsg["content"] = []map[string]interface{}{
				{
					"type": textContentType,
					"text": contentStr,
				},
			}
		}
	}

	// Add other message fields if present
	if message.Name != nil {
		responseMsg["name"] = *message.Name
	}

	return responseMsg
}

// ConvertChatCompletionToResponseAPI converts a ChatCompletion request to Response API format
func ConvertChatCompletionToResponseAPI(request *model.GeneralOpenAIRequest) *ResponseAPIRequest {
	responseReq := &ResponseAPIRequest{
		Model: request.Model,
		Input: make(ResponseAPIInput, 0, len(request.Messages)),
	}

	// Convert messages to input - Response API expects messages directly in the input array
	// IMPORTANT: Response API doesn't support ChatCompletion function call history format
	// We'll convert function call history to text summaries to preserve context
	var pendingToolCalls []model.Tool
	var pendingToolResults []string

	for _, message := range request.Messages {
		if message.Role == "tool" {
			// Collect tool results to summarize
			pendingToolResults = append(pendingToolResults, fmt.Sprintf("Function %s returned: %s",
				findToolCallName(pendingToolCalls, message.ToolCallId), message.StringContent()))
			continue
		} else if message.Role == "assistant" && len(message.ToolCalls) > 0 {
			// Collect tool calls for summarization
			pendingToolCalls = append(pendingToolCalls, message.ToolCalls...)

			// If assistant has text content, include it
			if message.Content != "" {
				convertedMsg := convertMessageToResponseAPIFormat(message)
				responseReq.Input = append(responseReq.Input, convertedMsg)
			}
		} else {
			// For regular messages, add any pending function call summary first
			if len(pendingToolCalls) > 0 && len(pendingToolResults) > 0 {
				// Create a summary message for the function call interactions
				summary := "Previous function calls:\n"
				for i, toolCall := range pendingToolCalls {
					summary += fmt.Sprintf("- Called %s(%s)", toolCall.Function.Name, toolCall.Function.Arguments)
					if i < len(pendingToolResults) {
						summary += fmt.Sprintf(" → %s", pendingToolResults[i])
					}
					summary += "\n"
				}

				summaryMsg := model.Message{
					Role:    "assistant",
					Content: summary,
				}
				convertedSummaryMsg := convertMessageToResponseAPIFormat(summaryMsg)
				responseReq.Input = append(responseReq.Input, convertedSummaryMsg)

				// Clear pending calls and results
				pendingToolCalls = nil
				pendingToolResults = nil
			}

			// Add the regular message - convert to Response API format
			convertedMsg := convertMessageToResponseAPIFormat(message)
			responseReq.Input = append(responseReq.Input, convertedMsg)
		}
	}

	// Add any remaining pending function call summary at the end
	if len(pendingToolCalls) > 0 && len(pendingToolResults) > 0 {
		summary := "Previous function calls:\n"
		for i, toolCall := range pendingToolCalls {
			summary += fmt.Sprintf("- Called %s(%s)", toolCall.Function.Name, toolCall.Function.Arguments)
			if i < len(pendingToolResults) {
				summary += fmt.Sprintf(" → %s", pendingToolResults[i])
			}
			summary += "\n"
		}

		summaryMsg := model.Message{
			Role:    "assistant",
			Content: summary,
		}
		convertedSummaryMsg := convertMessageToResponseAPIFormat(summaryMsg)
		responseReq.Input = append(responseReq.Input, convertedSummaryMsg)
	}

	// Map other fields
	if request.MaxTokens > 0 {
		responseReq.MaxOutputTokens = &request.MaxTokens
	}
	if request.MaxCompletionTokens != nil {
		responseReq.MaxOutputTokens = request.MaxCompletionTokens
	}

	responseReq.Temperature = request.Temperature
	responseReq.TopP = request.TopP
	responseReq.Stream = &request.Stream
	responseReq.User = &request.User
	responseReq.Store = request.Store
	responseReq.Metadata = request.Metadata

	if request.ServiceTier != nil {
		responseReq.ServiceTier = request.ServiceTier
	}

	if request.ParallelTooCalls != nil {
		responseReq.ParallelToolCalls = request.ParallelTooCalls
	}

	// Handle tools (modern format)
	if len(request.Tools) > 0 {
		// Convert ChatCompletion tools to Response API tool format
		responseAPITools := make([]ResponseAPITool, 0, len(request.Tools))
		for _, tool := range request.Tools {
			responseAPITool := ResponseAPITool{
				Type:        tool.Type,
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			}
			responseAPITools = append(responseAPITools, responseAPITool)
		}
		responseReq.Tools = responseAPITools
		responseReq.ToolChoice = request.ToolChoice
	} else if len(request.Functions) > 0 {
		// Handle legacy functions format by converting to Response API tool format
		responseAPITools := make([]ResponseAPITool, 0, len(request.Functions))
		for _, function := range request.Functions {
			responseAPITool := ResponseAPITool{
				Type:        "function",
				Name:        function.Name,
				Description: function.Description,
				Parameters:  function.Parameters,
			}
			responseAPITools = append(responseAPITools, responseAPITool)
		}
		responseReq.Tools = responseAPITools
		responseReq.ToolChoice = request.FunctionCall
	}

	// Handle thinking/reasoning
	if isModelSupportedReasoning(request.Model) {
		if request.ReasoningEffort != nil || request.Thinking != nil {
			if responseReq.Reasoning == nil {
				responseReq.Reasoning = &model.OpenAIResponseReasoning{
					Effort: request.ReasoningEffort,
				}
			}
		}

		// Initialize reasoning if not already set for reasoning-supported models
		if responseReq.Reasoning == nil {
			responseReq.Reasoning = &model.OpenAIResponseReasoning{}
		}

		if responseReq.Reasoning.Summary == nil {
			reasoningSummary := "detailed"
			responseReq.Reasoning.Summary = &reasoningSummary
		}

		if request.ReasoningEffort == nil {
			reasoningEffort := "high"
			responseReq.Reasoning.Effort = &reasoningEffort
		}
	} else {
		request.ReasoningEffort = nil
	}

	// Handle response format
	if request.ResponseFormat != nil {
		textConfig := &ResponseTextConfig{
			Format: &ResponseTextFormat{
				Type: request.ResponseFormat.Type,
			},
		}

		// Handle structured output with JSON schema
		if request.ResponseFormat.JsonSchema != nil {
			textConfig.Format.Name = request.ResponseFormat.JsonSchema.Name
			textConfig.Format.Description = request.ResponseFormat.JsonSchema.Description
			textConfig.Format.Schema = request.ResponseFormat.JsonSchema.Schema
			textConfig.Format.Strict = request.ResponseFormat.JsonSchema.Strict
		}

		responseReq.Text = textConfig
	}

	// Handle system message as instructions
	if len(request.Messages) > 0 && request.Messages[0].Role == "system" {
		systemContent := request.Messages[0].StringContent()
		responseReq.Instructions = &systemContent

		// Remove system message from input since it's now in instructions
		responseReq.Input = responseReq.Input[1:]
	}

	return responseReq
}

// ResponseAPIUsage represents the usage information structure for Response API
// Response API uses different field names than Chat Completions API
type ResponseAPIUsage struct {
	InputTokens         int                                 `json:"input_tokens"`                    // Tokens used in the input
	OutputTokens        int                                 `json:"output_tokens"`                   // Tokens used in the output
	TotalTokens         int                                 `json:"total_tokens"`                    // Total tokens used
	InputTokensDetails  *model.UsagePromptTokensDetails     `json:"input_tokens_details,omitempty"`  // Details about input tokens
	OutputTokensDetails *model.UsageCompletionTokensDetails `json:"output_tokens_details,omitempty"` // Details about output tokens
}

// ToModelUsage converts ResponseAPIUsage to model.Usage for compatibility
func (r *ResponseAPIUsage) ToModelUsage() *model.Usage {
	if r == nil {
		return nil
	}

	return &model.Usage{
		PromptTokens:            r.InputTokens,
		CompletionTokens:        r.OutputTokens,
		TotalTokens:             r.TotalTokens,
		PromptTokensDetails:     r.InputTokensDetails,
		CompletionTokensDetails: r.OutputTokensDetails,
	}
}

// FromModelUsage converts model.Usage to ResponseAPIUsage for compatibility
func (r *ResponseAPIUsage) FromModelUsage(usage *model.Usage) *ResponseAPIUsage {
	if usage == nil {
		return nil
	}

	return &ResponseAPIUsage{
		InputTokens:         usage.PromptTokens,
		OutputTokens:        usage.CompletionTokens,
		TotalTokens:         usage.TotalTokens,
		InputTokensDetails:  usage.PromptTokensDetails,
		OutputTokensDetails: usage.CompletionTokensDetails,
	}
}

// ResponseAPIResponse represents the OpenAI Response API response structure
// https://platform.openai.com/docs/api-reference/responses
type ResponseAPIResponse struct {
	Id                 string                         `json:"id"`                             // Unique identifier for this Response
	Object             string                         `json:"object"`                         // The object type of this resource - always set to "response"
	CreatedAt          int64                          `json:"created_at"`                     // Unix timestamp (in seconds) of when this Response was created
	Status             string                         `json:"status"`                         // The status of the response generation
	Model              string                         `json:"model"`                          // Model ID used to generate the response
	Output             []OutputItem                   `json:"output"`                         // An array of content items generated by the model
	Usage              *ResponseAPIUsage              `json:"usage,omitempty"`                // Token usage details (Response API format)
	Instructions       *string                        `json:"instructions,omitempty"`         // System message as the first item in the model's context
	MaxOutputTokens    *int                           `json:"max_output_tokens,omitempty"`    // Upper bound for the number of tokens
	Metadata           any                            `json:"metadata,omitempty"`             // Set of 16 key-value pairs
	ParallelToolCalls  bool                           `json:"parallel_tool_calls"`            // Whether to allow the model to run tool calls in parallel
	PreviousResponseId *string                        `json:"previous_response_id,omitempty"` // The unique ID of the previous response
	Reasoning          *model.OpenAIResponseReasoning `json:"reasoning,omitempty"`            // Configuration options for reasoning models
	ServiceTier        *string                        `json:"service_tier,omitempty"`         // Latency tier used for processing
	Temperature        *float64                       `json:"temperature,omitempty"`          // Sampling temperature used
	Text               *ResponseTextConfig            `json:"text,omitempty"`                 // Configuration options for text response
	ToolChoice         any                            `json:"tool_choice,omitempty"`          // How the model selected tools
	Tools              []model.Tool                   `json:"tools,omitempty"`                // Array of tools the model may call
	TopP               *float64                       `json:"top_p,omitempty"`                // Alternative to sampling with temperature
	Truncation         *string                        `json:"truncation,omitempty"`           // Truncation strategy
	User               *string                        `json:"user,omitempty"`                 // Stable identifier for end-users
	Error              *model.Error                   `json:"error,omitempty"`                // Error object if the response failed
	IncompleteDetails  *IncompleteDetails             `json:"incomplete_details,omitempty"`   // Details about why the response is incomplete
}

// OutputItem represents an item in the response output array
type OutputItem struct {
	Type    string          `json:"type"`              // Type of output item (e.g., "message", "reasoning", "function_call")
	Id      string          `json:"id,omitempty"`      // Unique identifier for this item
	Status  string          `json:"status,omitempty"`  // Status of this item (e.g., "completed")
	Role    string          `json:"role,omitempty"`    // Role of the message (e.g., "assistant")
	Content []OutputContent `json:"content,omitempty"` // Array of content items
	Summary []OutputContent `json:"summary,omitempty"` // Array of summary items (for reasoning)
	// Function call fields
	CallId    string `json:"call_id,omitempty"`   // Call ID for function calls
	Name      string `json:"name,omitempty"`      // Function name for function calls
	Arguments string `json:"arguments,omitempty"` // Function arguments for function calls
}

// OutputContent represents content within an output item
type OutputContent struct {
	Type        string `json:"type"`                  // Type of content (e.g., "output_text", "summary_text")
	Text        string `json:"text,omitempty"`        // Text content
	Annotations []any  `json:"annotations,omitempty"` // Annotations for the content
}

// IncompleteDetails provides details about why a response is incomplete
type IncompleteDetails struct {
	Reason string `json:"reason,omitempty"` // Reason why the response is incomplete
}

// ConvertResponseAPIToChatCompletion converts a Response API response back to ChatCompletion format
// This function follows the same pattern as ResponseClaude2OpenAI in the anthropic adaptor
func ConvertResponseAPIToChatCompletion(responseAPIResp *ResponseAPIResponse) *TextResponse {
	var responseText string
	var reasoningText string
	tools := make([]model.Tool, 0)

	// Extract content from output array
	for _, outputItem := range responseAPIResp.Output {
		switch outputItem.Type {
		case "message":
			if outputItem.Role == "assistant" {
				for _, content := range outputItem.Content {
					switch content.Type {
					case "output_text":
						responseText += content.Text
					case "reasoning":
						reasoningText += content.Text
					default:
						// Handle other content types if needed
					}
				}
			}
		case "reasoning":
			// Handle reasoning items separately
			for _, summaryContent := range outputItem.Summary {
				if summaryContent.Type == "summary_text" {
					reasoningText += summaryContent.Text
				}
			}
		case "function_call":
			// Handle function call items
			if outputItem.CallId != "" && outputItem.Name != "" {
				tool := model.Tool{
					Id:   outputItem.CallId,
					Type: "function",
					Function: model.Function{
						Name:      outputItem.Name,
						Arguments: outputItem.Arguments,
					},
				}
				tools = append(tools, tool)
			}
		}
	}

	// Handle reasoning content from reasoning field if present
	if responseAPIResp.Reasoning != nil {
		// Reasoning content would be handled here if needed
	}

	// Convert status to finish reason
	finishReason := "stop"
	switch responseAPIResp.Status {
	case "completed":
		finishReason = "stop"
	case "failed":
		finishReason = "stop"
	case "incomplete":
		finishReason = "length"
	case "cancelled":
		finishReason = "stop"
	default:
		finishReason = "stop"
	}

	choice := TextResponseChoice{
		Index: 0,
		Message: model.Message{
			Role:      "assistant",
			Content:   responseText,
			Name:      nil,
			ToolCalls: tools,
		},
		FinishReason: finishReason,
	}

	if reasoningText != "" {
		choice.Message.Reasoning = &reasoningText
	}

	// Create the chat completion response
	fullTextResponse := TextResponse{
		Id:      responseAPIResp.Id,
		Model:   responseAPIResp.Model,
		Object:  "chat.completion",
		Created: responseAPIResp.CreatedAt,
		Choices: []TextResponseChoice{choice},
	}

	// Set usage if available and valid - convert Response API usage fields to Chat Completion format
	if responseAPIResp.Usage != nil {
		if convertedUsage := responseAPIResp.Usage.ToModelUsage(); convertedUsage != nil {
			// Only set usage if it contains meaningful data
			if convertedUsage.PromptTokens > 0 || convertedUsage.CompletionTokens > 0 || convertedUsage.TotalTokens > 0 {
				fullTextResponse.Usage = *convertedUsage
			}
		}
	}
	// Note: If usage is nil or contains no meaningful data, the caller should calculate tokens

	return &fullTextResponse
}

// ConvertResponseAPIStreamToChatCompletion converts a Response API streaming response chunk back to ChatCompletion streaming format
// This function handles individual streaming chunks from the Response API
func ConvertResponseAPIStreamToChatCompletion(responseAPIChunk *ResponseAPIResponse) *ChatCompletionsStreamResponse {
	return ConvertResponseAPIStreamToChatCompletionWithIndex(responseAPIChunk, nil)
}

// ConvertResponseAPIStreamToChatCompletionWithIndex converts a Response API streaming response chunk back to ChatCompletion streaming format
// with optional output_index from streaming events for proper tool call index assignment
func ConvertResponseAPIStreamToChatCompletionWithIndex(responseAPIChunk *ResponseAPIResponse, outputIndex *int) *ChatCompletionsStreamResponse {
	var deltaContent string
	var reasoningText string
	var finishReason *string
	var toolCalls []model.Tool

	// Extract content from output array
	for _, outputItem := range responseAPIChunk.Output {
		switch outputItem.Type {
		case "message":
			if outputItem.Role == "assistant" {
				for _, content := range outputItem.Content {
					switch content.Type {
					case "output_text":
						deltaContent += content.Text
					case "reasoning":
						reasoningText += content.Text
					default:
						// Handle other content types if needed
					}
				}
			}
		case "reasoning":
			// Handle reasoning items separately - extract from summary content
			for _, summaryContent := range outputItem.Summary {
				if summaryContent.Type == "summary_text" {
					reasoningText += summaryContent.Text
				}
			}
		case "function_call":
			// Handle function call items
			if outputItem.CallId != "" && outputItem.Name != "" {
				// Set index for streaming tool calls
				// Use the provided outputIndex from streaming events if available, otherwise use position in slice
				var index int
				if outputIndex != nil {
					index = *outputIndex
				} else {
					index = len(toolCalls)
				}
				tool := model.Tool{
					Id:   outputItem.CallId,
					Type: "function",
					Function: model.Function{
						Name:      outputItem.Name,
						Arguments: outputItem.Arguments,
					},
					Index: &index, // Set index for streaming delta accumulation
				}
				toolCalls = append(toolCalls, tool)
			}
		}
	}

	// Convert status to finish reason for final chunks
	if responseAPIChunk.Status == "completed" {
		reason := "stop"
		finishReason = &reason
	} else if responseAPIChunk.Status == "failed" {
		reason := "stop"
		finishReason = &reason
	} else if responseAPIChunk.Status == "incomplete" {
		reason := "length"
		finishReason = &reason
	}

	// Create the streaming choice
	choice := ChatCompletionsStreamResponseChoice{
		Index: 0,
		Delta: model.Message{
			Role:    "assistant",
			Content: deltaContent,
		},
		FinishReason: finishReason,
	}

	// Set tool calls if present
	if len(toolCalls) > 0 {
		choice.Delta.ToolCalls = toolCalls
	}

	// Set reasoning content if present
	if reasoningText != "" {
		choice.Delta.Reasoning = &reasoningText
	}

	// Create the streaming response
	streamResponse := ChatCompletionsStreamResponse{
		Id:      responseAPIChunk.Id,
		Object:  "chat.completion.chunk",
		Created: responseAPIChunk.CreatedAt,
		Model:   responseAPIChunk.Model,
		Choices: []ChatCompletionsStreamResponseChoice{choice},
	}

	// Add usage if available (typically only in the final chunk)
	if responseAPIChunk.Usage != nil {
		streamResponse.Usage = responseAPIChunk.Usage.ToModelUsage()
	}

	return &streamResponse
}

// ResponseAPIStreamEvent represents a flexible structure for Response API streaming events
// This handles different event types that have varying schemas
type ResponseAPIStreamEvent struct {
	// Common fields for all events
	Type           string `json:"type,omitempty"`            // Event type (e.g., "response.output_text.done")
	SequenceNumber int    `json:"sequence_number,omitempty"` // Sequence number for ordering

	// Response-level events (type starts with "response.")
	Response *ResponseAPIResponse `json:"response,omitempty"` // Full response object for response-level events

	// Output item events (type contains "output_item")
	OutputIndex int         `json:"output_index,omitempty"` // Index of the output item
	Item        *OutputItem `json:"item,omitempty"`         // Output item for item-level events

	// Content events (type contains "content" or "output_text")
	ItemId       string         `json:"item_id,omitempty"`       // ID of the item containing the content
	ContentIndex int            `json:"content_index,omitempty"` // Index of the content within the item
	Part         *OutputContent `json:"part,omitempty"`          // Content part for part-level events
	Delta        string         `json:"delta,omitempty"`         // Delta content for streaming
	Text         string         `json:"text,omitempty"`          // Full text content (for done events)

	// Function call events (type contains "function_call")
	Arguments string `json:"arguments,omitempty"` // Complete function arguments (for done events)

	// General fields that might be in any event
	Id     string       `json:"id,omitempty"`     // Event ID
	Status string       `json:"status,omitempty"` // Event status
	Usage  *model.Usage `json:"usage,omitempty"`  // Usage information
}

// ParseResponseAPIStreamEvent attempts to parse a streaming event as either a full response
// or a streaming event, returning the appropriate data structure
func ParseResponseAPIStreamEvent(data []byte) (*ResponseAPIResponse, *ResponseAPIStreamEvent, error) {
	// First try to parse as a full ResponseAPIResponse (for response-level events)
	var fullResponse ResponseAPIResponse
	if err := json.Unmarshal(data, &fullResponse); err == nil && fullResponse.Id != "" {
		return &fullResponse, nil, nil
	}

	// If that fails, try to parse as a streaming event
	var streamEvent ResponseAPIStreamEvent
	if err := json.Unmarshal(data, &streamEvent); err != nil {
		return nil, nil, err
	}

	return nil, &streamEvent, nil
}

// ConvertStreamEventToResponse converts a streaming event to a ResponseAPIResponse structure
// This allows us to use the existing conversion logic for different event types
func ConvertStreamEventToResponse(event *ResponseAPIStreamEvent) ResponseAPIResponse {
	// Convert model.Usage to ResponseAPIUsage if present
	var responseUsage *ResponseAPIUsage
	if event.Usage != nil {
		responseUsage = (&ResponseAPIUsage{}).FromModelUsage(event.Usage)
	}

	response := ResponseAPIResponse{
		Id:        event.Id,
		Object:    "response",
		Status:    "in_progress", // Default status for streaming events
		Usage:     responseUsage,
		CreatedAt: 0, // Will be filled by the conversion logic if needed
	}

	// If the event already has a specific status, use it
	if event.Status != "" {
		response.Status = event.Status
	}

	// Handle different event types
	switch {
	case event.Response != nil:
		// Handle events that contain a full response object (response.created, response.completed, etc.)
		return *event.Response

	case strings.HasPrefix(event.Type, "response.reasoning_summary_text.delta"):
		// Handle reasoning summary text delta events
		if event.Delta != "" {
			outputItem := OutputItem{
				Type: "reasoning",
				Summary: []OutputContent{
					{
						Type: "summary_text",
						Text: event.Delta,
					},
				},
			}
			response.Output = []OutputItem{outputItem}
		}

	case strings.HasPrefix(event.Type, "response.reasoning_summary_text.done"):
		// Handle reasoning summary text completion events
		if event.Text != "" {
			outputItem := OutputItem{
				Type: "reasoning",
				Summary: []OutputContent{
					{
						Type: "summary_text",
						Text: event.Text,
					},
				},
			}
			response.Output = []OutputItem{outputItem}
		}

	case strings.HasPrefix(event.Type, "response.output_text.delta"):
		// Handle text delta events
		if event.Delta != "" {
			outputItem := OutputItem{
				Type: "message",
				Role: "assistant",
				Content: []OutputContent{
					{
						Type: "output_text",
						Text: event.Delta,
					},
				},
			}
			response.Output = []OutputItem{outputItem}
		}

	case strings.HasPrefix(event.Type, "response.output_text.done"):
		// Handle text completion events
		if event.Text != "" {
			outputItem := OutputItem{
				Type: "message",
				Role: "assistant",
				Content: []OutputContent{
					{
						Type: "output_text",
						Text: event.Text,
					},
				},
			}
			response.Output = []OutputItem{outputItem}
		}

	case strings.HasPrefix(event.Type, "response.output_item"):
		// Handle output item events (added, done)
		if event.Item != nil {
			response.Output = []OutputItem{*event.Item}
		}

	case strings.HasPrefix(event.Type, "response.function_call_arguments.delta"):
		// Handle function call arguments delta events
		if event.Delta != "" {
			outputItem := OutputItem{
				Type:      "function_call",
				Arguments: event.Delta, // This is a delta, not complete arguments
			}
			response.Output = []OutputItem{outputItem}
		}

	case strings.HasPrefix(event.Type, "response.function_call_arguments.done"):
		// Handle function call arguments completion events
		if event.Arguments != "" {
			outputItem := OutputItem{
				Type:      "function_call",
				Arguments: event.Arguments, // Complete arguments
			}
			response.Output = []OutputItem{outputItem}
		}

	case strings.HasPrefix(event.Type, "response.content_part"):
		// Handle content part events (added, done)
		if event.Part != nil {
			outputItem := OutputItem{
				Type:    "message",
				Role:    "assistant",
				Content: []OutputContent{*event.Part},
			}
			response.Output = []OutputItem{outputItem}
		}

	case strings.HasPrefix(event.Type, "response.reasoning_summary_part"):
		// Handle reasoning summary part events (added, done)
		if event.Part != nil {
			outputItem := OutputItem{
				Type:    "reasoning",
				Summary: []OutputContent{*event.Part},
			}
			response.Output = []OutputItem{outputItem}
		}

	case strings.HasPrefix(event.Type, "response."):
		// Handle other response-level events (in_progress, etc.)
		// These typically don't have content but may have metadata
		// The response structure is already set up above with basic fields

	default:
		// Unknown event type - log but don't fail
		// The response structure is already set up above with basic fields
	}

	return response
}
