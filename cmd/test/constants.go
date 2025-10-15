package main

const (
	defaultAPIBase    = "https://oneapi.laisky.com"
	defaultTestModels = "gpt-4o-mini,gpt-5-mini,claude-3.5-haiku,gemini-2.5-flash,openai/gpt-oss-20b,deepseek-chat"

	defaultMaxTokens   = 2048
	defaultTemperature = 0.7
	defaultTopP        = 0.9
	defaultTopK        = 40

	maxResponseBodySize = 1 << 20 // 1 MiB
	maxLoggedBodyBytes  = 2048
)
