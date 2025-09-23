package model

// Usage is the token usage information returned by OpenAI API.
type Usage struct {
	// Omitting this field using 'omitempty' is crucial to avoid returning zero values
	// when conversion mechanisms are not employed, particularly in scenarios like image generation.
	//
	// TODO: With Go 1.24 ~ latest potentially supporting 'omitzero', do we need to use both 'omitempty' and 'omitzero' here?
	PromptTokens     int `json:"prompt_tokens,omitempty"`
	CompletionTokens int `json:"completion_tokens,omitempty"`
	TotalTokens      int `json:"total_tokens,omitempty"`
	// PromptTokensDetails may be empty for some models
	PromptTokensDetails *UsagePromptTokensDetails `json:"prompt_tokens_details,omitempty"`
	// CompletionTokensDetails may be empty for some models
	CompletionTokensDetails *UsageCompletionTokensDetails `json:"completion_tokens_details,omitempty"`
	ServiceTier             string                        `json:"service_tier,omitempty"`
	SystemFingerprint       string                        `json:"system_fingerprint,omitempty"`

	// -------------------------------------
	// Custom fields
	// -------------------------------------
	// ToolsCost is the cost of using tools, in quota.
	ToolsCost int64 `json:"tools_cost,omitempty"`

	// Cache write token details (Anthropic Claude prompt caching)
	// These fields capture how many input tokens were charged as cache creation writes.
	// They are optional and only set when providers return such details.
	CacheWrite5mTokens int `json:"cache_write_5m_tokens,omitempty"`
	CacheWrite1hTokens int `json:"cache_write_1h_tokens,omitempty"`
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Param   string `json:"param"`
	Code    any    `json:"code"`
	// RawError preserves the original upstream or internal error for diagnostics.
	// Omitted from JSON to avoid leaking provider internals.
	RawError error `json:"-"`
}

type ErrorWithStatusCode struct {
	Error
	StatusCode int `json:"status_code"`
}

// UsagePromptTokensDetails contains details about the prompt tokens used in a request.
type UsagePromptTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
	AudioTokens  int `json:"audio_tokens"`
	// TextTokens could be zero for pure text chats
	TextTokens  int `json:"text_tokens"`
	ImageTokens int `json:"image_tokens"`
}

// UsageCompletionTokensDetails contains details about the completion tokens used in a request.
type UsageCompletionTokensDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens"`
	AudioTokens              int `json:"audio_tokens"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	// TextTokens could be zero for pure text chats
	TextTokens int `json:"text_tokens"`
	// CachedTokens indicates the count of completion tokens served from cache
	CachedTokens int `json:"cached_tokens"`
}
