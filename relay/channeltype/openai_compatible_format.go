package channeltype

import "strings"

const (
	// OpenAICompatibleAPIFormatChatCompletion indicates the upstream expects ChatCompletion payloads.
	OpenAICompatibleAPIFormatChatCompletion = "chat_completion"
	// OpenAICompatibleAPIFormatResponse indicates the upstream expects Response API payloads.
	OpenAICompatibleAPIFormatResponse = "response"
)

// NormalizeOpenAICompatibleAPIFormat trims and normalizes the configured API format for
// OpenAI-compatible channels. Unknown or empty values default to ChatCompletion for
// backward compatibility.
func NormalizeOpenAICompatibleAPIFormat(format string) string {
	normalized := strings.ToLower(strings.TrimSpace(format))
	switch normalized {
	case "response", "responses", "response_api", "responseapi":
		return OpenAICompatibleAPIFormatResponse
	case "chat_completion", "chat-completion", "chat", "chatcompletion":
		return OpenAICompatibleAPIFormatChatCompletion
	default:
		return OpenAICompatibleAPIFormatChatCompletion
	}
}

// UseOpenAICompatibleResponseAPI reports whether the upstream should receive Response API payloads.
func UseOpenAICompatibleResponseAPI(format string) bool {
	return NormalizeOpenAICompatibleAPIFormat(format) == OpenAICompatibleAPIFormatResponse
}
