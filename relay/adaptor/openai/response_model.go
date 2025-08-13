package openai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Laisky/errors/v2"

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
		return errors.Wrap(err, "ResponseAPIInput.UnmarshalJSON: failed to unmarshal as array")
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
			b, err := json.Marshal(str)
			if err != nil {
				return nil, errors.Wrap(err, "ResponseAPIInput.MarshalJSON: failed to marshal string")
			}
			return b, nil
		}
	}
	// Otherwise, marshal as array
	b, err := json.Marshal([]any(r))
	if err != nil {
		return nil, errors.Wrap(err, "ResponseAPIInput.MarshalJSON: failed to marshal array")
	}
	return b, nil
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
// Supports both function tools and MCP tools
type ResponseAPITool struct {
	Type        string                 `json:"type"`                  // Required: "function" or "mcp"
	Name        string                 `json:"name,omitempty"`        // Function name (for function tools)
	Description string                 `json:"description,omitempty"` // Function description (for function tools)
	Parameters  map[string]interface{} `json:"parameters,omitempty"`  // Function parameters (for function tools)

	// MCP-specific fields (for MCP tools)
	ServerLabel     string            `json:"server_label,omitempty"`
	ServerUrl       string            `json:"server_url,omitempty"`
	RequireApproval any               `json:"require_approval,omitempty"`
	AllowedTools    []string          `json:"allowed_tools,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
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

// MCPApprovalResponseInput represents the input structure for MCP approval responses
// Used when responding to mcp_approval_request output items to approve or deny MCP tool calls
type MCPApprovalResponseInput struct {
	Type              string `json:"type"`                // Required: Always "mcp_approval_response"
	Approve           bool   `json:"approve"`             // Required: Whether to approve the MCP tool call
	ApprovalRequestId string `json:"approval_request_id"` // Required: ID of the approval request being responded to
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
				assistantMsg := model.Message{
					Role:             message.Role,
					Content:          message.Content,
					Name:             message.Name,
					Reasoning:        message.Reasoning,
					ReasoningContent: message.ReasoningContent,
				}
				responseReq.Input = append(responseReq.Input, assistantMsg)
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
				responseReq.Input = append(responseReq.Input, summaryMsg)

				// Clear pending calls and results
				pendingToolCalls = nil
				pendingToolResults = nil
			}

			// Add the regular message
			responseReq.Input = append(responseReq.Input, message)
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
		responseReq.Input = append(responseReq.Input, summaryMsg)
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
			if tool.Type == "mcp" {
				// For MCP tools, preserve the original MCP structure
				// Response API expects MCP tools to maintain their MCP-specific fields
				responseAPITools = append(responseAPITools, ResponseAPITool{
					Type:            tool.Type,
					ServerLabel:     tool.ServerLabel,
					ServerUrl:       tool.ServerUrl,
					RequireApproval: tool.RequireApproval,
					AllowedTools:    tool.AllowedTools,
					Headers:         tool.Headers,
				})
			} else {
				// For function tools, use the existing logic
				responseAPITool := ResponseAPITool{
					Type:        tool.Type,
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
				}

				// Convert Parameters from any to map[string]interface{}
				if tool.Function.Parameters != nil {
					if params, ok := tool.Function.Parameters.(map[string]interface{}); ok {
						responseAPITool.Parameters = params
					}
				}

				responseAPITools = append(responseAPITools, responseAPITool)
			}
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
			}

			// Convert Parameters from any to map[string]interface{}
			if function.Parameters != nil {
				if params, ok := function.Parameters.(map[string]interface{}); ok {
					responseAPITool.Parameters = params
				}
			}

			responseAPITools = append(responseAPITools, responseAPITool)
		}
		responseReq.Tools = responseAPITools
		responseReq.ToolChoice = request.FunctionCall
	}

	// Handle thinking/reasoning
	if request.ReasoningEffort != nil || request.Thinking != nil {
		if responseReq.Reasoning == nil {
			responseReq.Reasoning = &model.OpenAIResponseReasoning{}
		}

		if responseReq.Reasoning.Summary == nil {
			reasoningSummary := "detailed"
			responseReq.Reasoning.Summary = &reasoningSummary
		}

		if request.ReasoningEffort != nil {
			responseReq.Reasoning.Effort = request.ReasoningEffort
		}
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
	Type    string          `json:"type"`              // Type of output item (e.g., "message", "reasoning", "function_call", "mcp_list_tools", "mcp_call", "mcp_approval_request")
	Id      string          `json:"id,omitempty"`      // Unique identifier for this item
	Status  string          `json:"status,omitempty"`  // Status of this item (e.g., "completed")
	Role    string          `json:"role,omitempty"`    // Role of the message (e.g., "assistant")
	Content []OutputContent `json:"content,omitempty"` // Array of content items
	Summary []OutputContent `json:"summary,omitempty"` // Array of summary items (for reasoning)

	// Function call fields
	CallId    string `json:"call_id,omitempty"`   // Call ID for function calls
	Name      string `json:"name,omitempty"`      // Function name for function calls
	Arguments string `json:"arguments,omitempty"` // Function arguments for function calls

	// MCP-specific fields
	ServerLabel       string       `json:"server_label,omitempty"`        // Label for the MCP server (for mcp_list_tools, mcp_call, mcp_approval_request)
	Tools             []model.Tool `json:"tools,omitempty"`               // Array of tools from MCP server (for mcp_list_tools)
	ApprovalRequestId *string      `json:"approval_request_id,omitempty"` // ID of approval request (for mcp_call)
	Error             *string      `json:"error,omitempty"`               // Error message if MCP call failed (for mcp_call)
	Output            string       `json:"output,omitempty"`              // Output from MCP tool call (for mcp_call)
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
					Function: &model.Function{
						Name:      outputItem.Name,
						Arguments: outputItem.Arguments,
					},
				}
				tools = append(tools, tool)
			}
		case "mcp_list_tools":
			// Handle MCP list tools output - add server tools information to response text
			if outputItem.ServerLabel != "" && len(outputItem.Tools) > 0 {
				responseText += fmt.Sprintf("\nMCP Server '%s' tools imported: %d tools available",
					outputItem.ServerLabel, len(outputItem.Tools))
			}
		case "mcp_call":
			// Handle MCP tool call output - add call result to response text
			if outputItem.Name != "" && outputItem.Output != "" {
				responseText += fmt.Sprintf("\nMCP Tool '%s' result: %s", outputItem.Name, outputItem.Output)
			} else if outputItem.Error != nil && *outputItem.Error != "" {
				responseText += fmt.Sprintf("\nMCP Tool '%s' error: %s", outputItem.Name, *outputItem.Error)
			}
		case "mcp_approval_request":
			// Handle MCP approval request - add approval request info to response text
			if outputItem.ServerLabel != "" && outputItem.Name != "" {
				responseText += fmt.Sprintf("\nMCP Approval Required: Server '%s' requests approval to call '%s'",
					outputItem.ServerLabel, outputItem.Name)
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
					Function: &model.Function{
						Name:      outputItem.Name,
						Arguments: outputItem.Arguments,
					},
					Index: &index, // Set index for streaming delta accumulation
				}
				toolCalls = append(toolCalls, tool)
			}
		// Note: This is currently unavailable in the OpenAI Docs.
		// It's added here for reference because OpenAI's Remote MCP is included in their tools, unlike other Remote MCPs such as Anthropic Claude.
		case "mcp_list_tools":
			// Handle MCP list tools output in streaming - add server tools information as delta content
			if outputItem.ServerLabel != "" && len(outputItem.Tools) > 0 {
				deltaContent += fmt.Sprintf("\nMCP Server '%s' tools imported: %d tools available",
					outputItem.ServerLabel, len(outputItem.Tools))
			}
		case "mcp_call":
			// Handle MCP tool call output in streaming - add call result as delta content
			if outputItem.Name != "" && outputItem.Output != "" {
				deltaContent += fmt.Sprintf("\nMCP Tool '%s' result: %s", outputItem.Name, outputItem.Output)
			} else if outputItem.Error != nil && *outputItem.Error != "" {
				deltaContent += fmt.Sprintf("\nMCP Tool '%s' error: %s", outputItem.Name, *outputItem.Error)
			}
		case "mcp_approval_request":
			// Handle MCP approval request in streaming - add approval request info as delta content
			if outputItem.ServerLabel != "" && outputItem.Name != "" {
				deltaContent += fmt.Sprintf("\nMCP Approval Required: Server '%s' requests approval to call '%s'",
					outputItem.ServerLabel, outputItem.Name)
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
		return nil, nil, errors.Wrap(err, "ParseResponseAPIStreamEvent: failed to unmarshal as stream event")
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
