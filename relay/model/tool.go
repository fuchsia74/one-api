package model

// Tool represents a tool definition used in AI model interactions.
// It contains metadata about the tool and its associated function or MCP server configuration.
// This struct supports both function-based tools and Remote MCP server tools.
type Tool struct {
	Id       string   `json:"id,omitempty"`       // Unique identifier for the tool
	Type     string   `json:"type,omitempty"`     // Tool type (e.g., "function", "mcp"), may be empty when splicing claude tools stream messages
	Function Function `json:"function,omitempty"` // Function definition (for type="function")
	Index    *int     `json:"index,omitempty"`    // Index identifies which function call the delta is for in streaming responses

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
