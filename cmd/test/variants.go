package main

var requestVariants = []requestVariant{
	{Key: "chat_stream_false", Header: "Chat (stream=false)", Type: requestTypeChatCompletion, Path: "/v1/chat/completions", Stream: false, Expectation: expectationDefault},
	{Key: "chat_stream_true", Header: "Chat (stream=true)", Type: requestTypeChatCompletion, Path: "/v1/chat/completions", Stream: true, Expectation: expectationDefault},
	{Key: "chat_tools_stream_false", Header: "Chat Tools (stream=false)", Type: requestTypeChatCompletion, Path: "/v1/chat/completions", Stream: false, Expectation: expectationToolInvocation, Aliases: []string{"chat_tools"}},
	{Key: "chat_tools_stream_true", Header: "Chat Tools (stream=true)", Type: requestTypeChatCompletion, Path: "/v1/chat/completions", Stream: true, Expectation: expectationToolInvocation, Aliases: []string{"chat_tools_stream"}},
	{Key: "chat_structured_stream_false", Header: "Chat Structured (stream=false)", Type: requestTypeChatCompletion, Path: "/v1/chat/completions", Stream: false, Expectation: expectationStructuredOutput, Aliases: []string{"chat_structured"}},
	{Key: "chat_structured_stream_true", Header: "Chat Structured (stream=true)", Type: requestTypeChatCompletion, Path: "/v1/chat/completions", Stream: true, Expectation: expectationStructuredOutput, Aliases: []string{"chat_structured_stream"}},

	{Key: "response_stream_false", Header: "Response (stream=false)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: false, Expectation: expectationDefault},
	{Key: "response_stream_true", Header: "Response (stream=true)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: true, Expectation: expectationDefault},
	{Key: "response_vision_stream_false", Header: "Response Vision (stream=false)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: false, Expectation: expectationVision, Aliases: []string{"response_vision"}},
	{Key: "response_vision_stream_true", Header: "Response Vision (stream=true)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: true, Expectation: expectationVision, Aliases: []string{"response_vision_stream"}},
	{Key: "response_tools_stream_false", Header: "Response Tools (stream=false)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: false, Expectation: expectationToolInvocation, Aliases: []string{"response_tools"}},
	{Key: "response_tools_stream_true", Header: "Response Tools (stream=true)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: true, Expectation: expectationToolInvocation, Aliases: []string{"response_tools_stream"}},
	{Key: "response_structured_stream_false", Header: "Response Structured (stream=false)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: false, Expectation: expectationStructuredOutput, Aliases: []string{"response_structured"}},
	{Key: "response_structured_stream_true", Header: "Response Structured (stream=true)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: true, Expectation: expectationStructuredOutput, Aliases: []string{"response_structured_stream"}},

	{Key: "claude_stream_false", Header: "Claude (stream=false)", Type: requestTypeClaudeMessages, Path: "/v1/messages", Stream: false, Expectation: expectationDefault},
	{Key: "claude_stream_true", Header: "Claude (stream=true)", Type: requestTypeClaudeMessages, Path: "/v1/messages", Stream: true, Expectation: expectationDefault},
	{Key: "claude_tools_stream_false", Header: "Claude Tools (stream=false)", Type: requestTypeClaudeMessages, Path: "/v1/messages", Stream: false, Expectation: expectationToolInvocation, Aliases: []string{"claude_tools"}},
	{Key: "claude_tools_stream_true", Header: "Claude Tools (stream=true)", Type: requestTypeClaudeMessages, Path: "/v1/messages", Stream: true, Expectation: expectationToolInvocation, Aliases: []string{"claude_tools_stream"}},
	{Key: "claude_structured_stream_false", Header: "Claude Structured (stream=false)", Type: requestTypeClaudeMessages, Path: "/v1/messages", Stream: false, Expectation: expectationStructuredOutput, Aliases: []string{"claude_structured"}},
	{Key: "claude_structured_stream_true", Header: "Claude Structured (stream=true)", Type: requestTypeClaudeMessages, Path: "/v1/messages", Stream: true, Expectation: expectationStructuredOutput, Aliases: []string{"claude_structured_stream"}},
}
