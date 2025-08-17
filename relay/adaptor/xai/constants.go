package xai

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// Pricing constants for input token pricing (USD per million tokens)
//
// Note: This is now easier to maintain than hardcoded.
const (
	// Standard model pricing tiers (USD per million tokens)
	Grok4InputPrice     = 3.0 // $3.00 per million input tokens for Grok-4 models
	Grok3InputPrice     = 3.0 // $3.00 per million input tokens for Grok-3 models
	Grok3MiniInputPrice = 0.3 // $0.30 per million input tokens for Grok-3-mini
	Grok3FastInputPrice = 5.0 // $5.00 per million input tokens for Grok-3-fast
	Grok3MiniFastPrice  = 0.6 // $0.60 per million input tokens for Grok-3-mini-fast
	Grok2InputPrice     = 2.0 // $2.00 per million input tokens for Grok-2 models

	// Completion price ratios (multiplier for output tokens relative to input)
	StandardCompletionRatio      = 5.0  // 5x price for output tokens (e.g., $15.00 per million for Grok-4)
	Grok3MiniCompletionRatio     = 1.67 // 1.67x price for Grok-3-mini output tokens ($0.50 per million)
	Grok3MiniFastCompletionRatio = 6.67 // 6.67x price for Grok-3-mini-fast output tokens ($4.00 per million)
	ImageCompletionRatio         = 1.0  // Same price for input and output tokens for image models

	// Special pricing constants
	ImagePrice = 0.07 / 0.002 // Conversion factor for image pricing ($0.07 per image)
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on X.AI pricing: https://console.x.ai/
var ModelRatios = map[string]adaptor.ModelConfig{
	// Grok Models - Based on https://console.x.ai/
	"grok-4-0709":        {Ratio: Grok4InputPrice * ratio.MilliTokensUsd, CompletionRatio: StandardCompletionRatio},         // $3.00 input, $15.00 output
	"grok-3":             {Ratio: Grok3InputPrice * ratio.MilliTokensUsd, CompletionRatio: StandardCompletionRatio},         // $3.00 input, $15.00 output
	"grok-3-mini":        {Ratio: Grok3MiniInputPrice * ratio.MilliTokensUsd, CompletionRatio: Grok3MiniCompletionRatio},    // $0.30 input, $0.50 output
	"grok-3-fast":        {Ratio: Grok3FastInputPrice * ratio.MilliTokensUsd, CompletionRatio: StandardCompletionRatio},     // $5.00 input, $25.00 output
	"grok-3-mini-fast":   {Ratio: Grok3MiniFastPrice * ratio.MilliTokensUsd, CompletionRatio: Grok3MiniFastCompletionRatio}, // $0.60 input, $4.00 output
	"grok-2-vision-1212": {Ratio: Grok2InputPrice * ratio.MilliTokensUsd, CompletionRatio: StandardCompletionRatio},         // $2.00 input, $10.00 output
	"grok-2-1212":        {Ratio: Grok2InputPrice * ratio.MilliTokensUsd, CompletionRatio: StandardCompletionRatio},         // $2.00 input, $10.00 output

	// Image generation model
	"grok-2-image-1212": {Ratio: ImagePrice * ratio.ImageUsdPerPic, CompletionRatio: ImageCompletionRatio}, // $0.07 per image

	// Legacy aliases for backward compatibility
	"grok-beta":        {Ratio: Grok2InputPrice * ratio.MilliTokensUsd, CompletionRatio: StandardCompletionRatio}, // Updated to match grok-2-1212
	"grok-2":           {Ratio: Grok2InputPrice * ratio.MilliTokensUsd, CompletionRatio: StandardCompletionRatio}, // Updated to match grok-2-1212
	"grok-2-latest":    {Ratio: Grok2InputPrice * ratio.MilliTokensUsd, CompletionRatio: StandardCompletionRatio}, // Updated to match grok-2-1212
	"grok-vision-beta": {Ratio: Grok2InputPrice * ratio.MilliTokensUsd, CompletionRatio: StandardCompletionRatio}, // Updated to match grok-2-vision-1212
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
