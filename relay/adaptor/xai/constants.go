package xai

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	ratio "github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on X.AI pricing: https://console.x.ai/
var ModelRatios = map[string]adaptor.ModelConfig{
	// Grok Models - Based on https://console.x.ai/
	//
	// Note: Some prices are the same because they are aliases.
	"grok-code-fast-1":          {Ratio: 0.2 * ratio.MilliTokensUsd, CompletionRatio: 7.5, CachedInputRatio: 0.02 * ratio.MilliTokensUsd},        // $0.20 input, $0.02 cached input, $1.50 output
	"grok-4-0709":               {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.75 * ratio.MilliTokensUsd},        // $3.00 input, $0.75 cached input, $15.00 output
	"grok-4-fast-reasoning":     {Ratio: 0.2 * ratio.MilliTokensUsd, CompletionRatio: 2.5, CachedInputRatio: 0.05 * ratio.MilliTokensUsd},        // $0.20 input, $0.05 cached input, $0.50 output
	"grok-4-fast-non-reasoning": {Ratio: 0.2 * ratio.MilliTokensUsd, CompletionRatio: 2.5, CachedInputRatio: 0.05 * ratio.MilliTokensUsd},        // $0.20 input, $0.05 cached input, $0.50 output
	"grok-3":                    {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.75 * ratio.MilliTokensUsd},        // $3.00 input, $0.75 cached input, $15.00 output
	"grok-3-mini":               {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 0.5 / 0.3, CachedInputRatio: 0.075 * ratio.MilliTokensUsd}, // $0.30 input, $0.075 cached input, $0.50 output
	"grok-3-fast":               {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.75 * ratio.MilliTokensUsd},        // $3.00 input, $0.75 cached input, $15.00 output
	"grok-3-mini-fast":          {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 0.5 / 0.3, CachedInputRatio: 0.075 * ratio.MilliTokensUsd}, // $0.30 input, $0.075 cached input, $0.50 output
	"grok-2-vision-1212":        {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},                                                       // $2.00 input, $10.00 output
	"grok-2-1212":               {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},                                                       // $2.00 input, $10.00 output

	// Image generation model (no per-token charge)
	"grok-2-image-1212": {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.07}, // $0.07 per image
	"grok-2-image":      {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.07}, // $0.07 per image

	// Legacy aliases for backward compatibility
	"grok-beta":        {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0}, // Updated to match grok-2-1212
	"grok-2":           {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0}, // Updated to match grok-2-1212
	"grok-2-latest":    {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0}, // Updated to match grok-2-1212
	"grok-vision-beta": {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0}, // Updated to match grok-2-vision-1212
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
