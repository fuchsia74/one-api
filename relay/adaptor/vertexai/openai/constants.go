// Package openai provides model pricing constants for OpenAI GPT-OSS models in Vertex AI.
package openai

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains pricing information for OpenAI GPT-OSS models
var ModelRatios = map[string]adaptor.ModelConfig{
	"openai/gpt-oss-20b-maas": {
		Ratio:           0.15 * ratio.MilliTokensUsd, // $0.15 per million tokens input
		CompletionRatio: 0.60 * ratio.MilliTokensUsd, // $0.60 per million tokens output
	},
	"openai/gpt-oss-120b-maas": {
		Ratio:           0.075 * ratio.MilliTokensUsd, // $0.075 per million tokens input
		CompletionRatio: 0.30 * ratio.MilliTokensUsd,  // $0.30 per million tokens output
	},
}

// ModelList contains all OpenAI GPT-OSS models supported by VertexAI
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
