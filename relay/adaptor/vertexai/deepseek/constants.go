// Package deepseek provides model pricing constants for DeepSeek AI models in Vertex AI.
package deepseek

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains DeepSeek models and their pricing ratios
var ModelRatios = map[string]adaptor.ModelConfig{
	// DeepSeek V3.1 - Input: $0.60 / million tokens, Output: $1.70 / million tokens
	"deepseek-ai/deepseek-v3.1-maas": {
		Ratio:           0.60 * ratio.MilliTokensUsd, // Input price: $0.60 per million tokens
		CompletionRatio: 1.70 / 0.60,                 // Output/Input ratio: $1.70 / $0.60 = 2.833
	},
	// DeepSeek R1 - Input: $1.35 / million tokens, Output: $5.40 / million tokens
	"deepseek-ai/deepseek-r1-0528-maas": {
		Ratio:           1.35 * ratio.MilliTokensUsd, // Input price: $1.35 per million tokens
		CompletionRatio: 5.40 / 1.35,                 // Output/Input ratio: $5.40 / $1.35 = 4.0
	},
}

// ModelList derived from ModelRatios keys
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
