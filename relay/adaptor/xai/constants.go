package xai

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on X.AI pricing: https://console.x.ai/
var ModelRatios = map[string]adaptor.ModelConfig{
	// Grok Models - Based on https://console.x.ai/
	"grok-4-0709":        {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $3.00 input, $15.00 output
	"grok-3":             {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $3.00 input, $15.00 output
	"grok-3-mini":        {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 1.67},    // $0.30 input, $0.50 output
	"grok-3-fast":        {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $5.00 input, $25.00 output
	"grok-3-mini-fast":   {Ratio: 0.6 * ratio.MilliTokensUsd, CompletionRatio: 6.67},    // $0.60 input, $4.00 output
	"grok-2-vision-1212": {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $2.00 input, $10.00 output
	"grok-2-1212":        {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5},       // $2.00 input, $10.00 output
	
	// Region-specific models
	"grok-3-fastus-east-1":        {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // $5.00 input, $25.00 output
	"grok-3-fasteu-west-1":        {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // $5.00 input, $25.00 output
	"grok-2-vision-1212us-east-1": {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // $2.00 input, $10.00 output
	"grok-2-1212us-east-1":        {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // $2.00 input, $10.00 output
	"grok-2-vision-1212eu-west-1": {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // $2.00 input, $10.00 output
	"grok-2-1212eu-west-1":        {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // $2.00 input, $10.00 output
	
	// Image generation model
	"grok-2-image-1212": {Ratio: (0.07 / 0.002) * ratio.ImageUsdPerPic, CompletionRatio: 1}, // $0.07 per image
	
	// Legacy aliases for backward compatibility
	"grok-beta":        {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // Updated to match grok-2-1212
	"grok-2":           {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // Updated to match grok-2-1212
	"grok-2-latest":    {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // Updated to match grok-2-1212
	"grok-vision-beta": {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5}, // Updated to match grok-2-vision-1212
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
