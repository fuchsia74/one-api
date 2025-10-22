package main

const (
	defaultAPIBase    = "https://oneapi.laisky.com"
	defaultTestModels = "gpt-4o-mini,gpt-5-mini,claude-haiku-4-5,gemini-2.5-flash,openai/gpt-oss-20b,deepseek-chat,grok-4-fast-non-reasoning,azure-gpt-5-nano"
	// defaultTestModels = "azure-gpt-5-nano"

	defaultMaxTokens   = 2048
	defaultTemperature = 0.7
	defaultTopP        = 0.9
	defaultTopK        = 40

	maxResponseBodySize = 1 << 20 // 1 MiB
	maxLoggedBodyBytes  = 2048
)

// visionUnsupportedModels enumerates models that are known to reject vision payloads.
var visionUnsupportedModels = map[string]struct{}{
	"deepseek-chat":      {},
	"openai/gpt-oss-20b": {},
}
