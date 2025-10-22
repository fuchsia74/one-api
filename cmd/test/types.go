package main

import "time"

// requestType identifies which API surface should be exercised.
type requestType string

const (
	requestTypeChatCompletion requestType = "chat_completion"
	requestTypeResponseAPI    requestType = "response_api"
	requestTypeClaudeMessages requestType = "claude_messages"
)

// requestVariant describes a single test permutation.
type requestVariant struct {
	Key         string
	Header      string
	Type        requestType
	Path        string
	Stream      bool
	Expectation expectation
	Aliases     []string
}

// expectation describes what a request variant should validate in a response.
type expectation int

const (
	expectationDefault expectation = iota
	expectationToolInvocation
	expectationVision
)

// testResult captures the outcome for a single request.
type testResult struct {
	Model        string
	Variant      string
	Label        string
	Type         requestType
	Stream       bool
	Success      bool
	Skipped      bool
	StatusCode   int
	Duration     time.Duration
	ErrorReason  string
	RequestBody  string
	ResponseBody string
}

// requestSpec contains the fully constructed request that will be executed.
type requestSpec struct {
	Variant     string
	Label       string
	Type        requestType
	Path        string
	Body        any
	Stream      bool
	Expectation expectation
}
