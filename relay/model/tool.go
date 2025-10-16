package model

import (
	"encoding/json"
	"net/url"

	"github.com/Laisky/errors/v2"
)

// Tool represents a tool definition used in AI model interactions.
// It contains metadata about the tool and its associated function or MCP server configuration.
// This struct supports both function-based tools and Remote MCP server tools.
type Tool struct {
	Id       string    `json:"id,omitempty"`       // Unique identifier for the tool
	Type     string    `json:"type,omitempty"`     // Tool type (e.g., "function", "mcp"), may be empty when splicing claude tools stream messages
	Function *Function `json:"function,omitempty"` // Function definition (for type="function")
	Index    *int      `json:"index,omitempty"`    // Index identifies which function call the delta is for in streaming responses
	// Web-search specific configuration (type="web_search")
	SearchContextSize *string           `json:"search_context_size,omitempty"`
	Filters           *WebSearchFilters `json:"filters,omitempty"`
	UserLocation      *UserLocation     `json:"user_location,omitempty"`

	// MCP-specific fields (for type="mcp")
	ServerLabel     string            `json:"server_label,omitempty"`     // Label for the MCP server
	ServerUrl       string            `json:"server_url,omitempty"`       // URL of the remote MCP server
	RequireApproval any               `json:"require_approval,omitempty"` // Approval requirement: "never", or object with tool-specific settings
	AllowedTools    []string          `json:"allowed_tools,omitempty"`    // List of allowed tool names from the MCP server
	Headers         map[string]string `json:"headers,omitempty"`          // Additional headers for MCP server requests (e.g., Authorization)
}

// Function represents a function definition within a tool.
// It contains the function's metadata including its description, name, parameters for requests,
// and arguments for responses. Used for both tool calling requests and responses.
type Function struct {
	Description string   `json:"description,omitempty"` // Human-readable description of what the function does
	Name        string   `json:"name,omitempty"`        // Function name, may be empty when splicing claude tools stream messages
	Parameters  any      `json:"parameters,omitempty"`  // Function parameters schema for requests (typically JSON Schema)
	Arguments   any      `json:"arguments,omitempty"`   // Function arguments data for responses (actual values passed to function)
	Required    []string `json:"required,omitempty"`    // Required parameter names for function validation
	Strict      *bool    `json:"strict,omitempty"`      // Whether to enforce strict parameter validation
}

// ValidateFunction validates function tool configuration
func (t *Tool) ValidateFunction() error {
	if t.Type == "function" {
		if t.Function == nil {
			return errors.New("function tool requires function definition")
		}
		if t.Function.Name == "" {
			return errors.New("function name is required")
		}
	}
	return nil
}

// ValidateMCP validates MCP tool configuration
func (t *Tool) ValidateMCP() error {
	if t.Type == "mcp" {
		if t.ServerLabel == "" {
			return errors.New("MCP tool requires server_label")
		}
		if t.ServerUrl == "" {
			return errors.New("MCP tool requires server_url")
		}
		// Validate URL format and scheme
		parsedURL, err := url.Parse(t.ServerUrl)
		if err != nil {
			return errors.Wrap(err, "invalid server_url format")
		}
		if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
			return errors.New("server_url must use http or https scheme")
		}
	}
	return nil
}

// Validate validates tool configuration based on type
func (t *Tool) Validate() error {
	switch t.Type {
	case "function":
		return t.ValidateFunction()
	case "mcp":
		return t.ValidateMCP()
	default:
		// Default to function validation for backward compatibility
		if t.Function != nil {
			return t.ValidateFunction()
		}
	}
	return nil
}

// UnmarshalJSON supports both nested OpenAI function definitions and flattened
// legacy payloads where function fields appear at the top level of the tool
// object. The upstream providers expect the nested format, so we normalize the
// data into the Function struct during decoding to ensure consistent marshaling
// later in the pipeline.
func (t *Tool) UnmarshalJSON(data []byte) error {
	type alias Tool
	var raw struct {
		alias
		Function    *Function `json:"function"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Parameters  any       `json:"parameters"`
		Arguments   any       `json:"arguments"`
		Required    []string  `json:"required"`
		Strict      *bool     `json:"strict"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return errors.Wrap(err, "unmarshal tool")
	}

	*t = Tool(raw.alias)
	t.Function = raw.Function

	if t.Function == nil {
		if hasFunctionShape(raw.Name, raw.Description, raw.Parameters, raw.Arguments, raw.Required, raw.Strict) {
			t.Function = &Function{
				Name:        raw.Name,
				Description: raw.Description,
				Parameters:  raw.Parameters,
				Arguments:   raw.Arguments,
				Required:    raw.Required,
				Strict:      raw.Strict,
			}
		}
		return nil
	}

	// Merge any flattened fields that were provided alongside the nested function
	if raw.Name != "" && t.Function.Name == "" {
		t.Function.Name = raw.Name
	}
	if raw.Description != "" && t.Function.Description == "" {
		t.Function.Description = raw.Description
	}
	if raw.Parameters != nil && t.Function.Parameters == nil {
		t.Function.Parameters = raw.Parameters
	}
	if raw.Arguments != nil && t.Function.Arguments == nil {
		t.Function.Arguments = raw.Arguments
	}
	if len(raw.Required) > 0 && len(t.Function.Required) == 0 {
		t.Function.Required = raw.Required
	}
	if raw.Strict != nil && t.Function.Strict == nil {
		t.Function.Strict = raw.Strict
	}

	return nil
}

func hasFunctionShape(name, description string, parameters, arguments any, required []string, strict *bool) bool {
	if name != "" || description != "" || parameters != nil || arguments != nil || strict != nil {
		return true
	}
	return len(required) > 0
}
