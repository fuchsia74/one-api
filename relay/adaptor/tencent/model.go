package tencent

type Message struct {
	Role    string `json:"Role"`
	Content string `json:"Content"`
}

type ChatRequest struct {
	// Model name, optional values include hunyuan-lite, hunyuan-standard, hunyuan-standard-256K, hunyuan-pro.
	// For descriptions of each model, please read the [Product Overview](https://cloud.tencent.com/document/product/1729/104753).
	//
	// Note:
	// Different models have different pricing. Please refer to the [Purchase Guide](https://cloud.tencent.com/document/product/1729/97731) for details.
	Model *string `json:"Model"`
	// Chat context information.
	// Description:
	// 1. The maximum length is 40, arranged in the array in chronological order from oldest to newest.
	// 2. Message.Role optional values: system, user, assistant.
	//    Among them, the system role is optional. If it exists, it must be at the beginning of the list.
	//    User and assistant must alternate (one question and one answer), starting and ending with user,
	//    and Content cannot be empty. The order of roles is as follows: [system (optional) user assistant user assistant user ...].
	// 3. The total length of Content in Messages cannot exceed the model's length limit
	//    (refer to the [Product Overview](https://cloud.tencent.com/document/product/1729/104753) document).
	//    If it exceeds, the earliest content will be truncated, leaving only the latest content.
	Messages []*Message `json:"Messages"`
	// Stream call switch.
	// Description:
	// 1. If not provided, the default is non-streaming call (false).
	// 2. In streaming calls, results are returned incrementally using the SSE protocol
	//    (the return value is taken from Choices[n].Delta, and incremental data needs to be concatenated to obtain the complete result).
	// 3. In non-streaming calls:
	// The call method is the same as a regular HTTP request.
	// The interface response time is relatively long. **If lower latency is required, it is recommended to set this to true**.
	// Only the final result is returned once (the return value is taken from Choices[n].Message).
	//
	// Note:
	// When calling through the SDK, different methods are required to obtain return values for streaming and non-streaming calls.
	// Refer to the comments or examples in the SDK (in the examples/hunyuan/v20230901/ directory of each language SDK code repository).
	Stream *bool `json:"Stream"`
	// Description:
	// 1. Affects the diversity of the output text. The larger the value, the more diverse the generated text.
	// 2. The value range is [0.0, 1.0]. If not provided, the recommended value for each model is used.
	// 3. It is not recommended to use this unless necessary, as unreasonable values can affect the results.
	TopP *float64 `json:"TopP,omitempty"`
	// Description:
	// 1. Higher values make the output more random, while lower values make it more focused and deterministic.
	// 2. The value range is [0.0, 2.0]. If not provided, the recommended value for each model is used.
	// 3. It is not recommended to use this unless necessary, as unreasonable values can affect the results.
	Temperature *float64 `json:"Temperature,omitempty"`
}

type Error struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type Usage struct {
	PromptTokens     int `json:"PromptTokens"`
	CompletionTokens int `json:"CompletionTokens"`
	TotalTokens      int `json:"TotalTokens"`
}

type ResponseChoices struct {
	FinishReason string  `json:"FinishReason,omitempty"` // Stream end flag, "stop" indicates the end packet
	Messages     Message `json:"Message"`                // Content, returned in synchronous mode, null in stream mode. The total content supports up to 1024 tokens.
	Delta        Message `json:"Delta"`                  // Content, returned in stream mode, null in synchronous mode. The total content supports up to 1024 tokens.
}

type ChatResponse struct {
	Choices []ResponseChoices `json:"Choices,omitempty"`   // Results
	Created int64             `json:"Created,omitempty"`   // Unix timestamp string
	Id      string            `json:"Id,omitempty"`        // Session id
	Usage   Usage             `json:"Usage"`               // Token count
	Error   Error             `json:"Error"`               // Error message. Note: this field may return null, indicating no valid value can be obtained
	Note    string            `json:"Note,omitempty"`      // Comment
	ReqID   string            `json:"RequestId,omitempty"` // Unique request Id, returned with each request. Used for feedback interface parameters
}

type ChatResponseP struct {
	Response ChatResponse `json:"Response"`
}

type EmbeddingRequest struct {
	InputList []string `json:"InputList"`
}

type EmbeddingData struct {
	Embedding []float64 `json:"Embedding"`
	Index     int       `json:"Index"`
	Object    string    `json:"Object"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"PromptTokens"`
	TotalTokens  int `json:"TotalTokens"`
}

type EmbeddingResponse struct {
	Data           []EmbeddingData `json:"Data"`
	EmbeddingUsage EmbeddingUsage  `json:"Usage"`
	RequestId      string          `json:"RequestId,omitempty"`
	Error          Error           `json:"Error"`
}

type EmbeddingResponseP struct {
	Response EmbeddingResponse `json:"Response"`
}
