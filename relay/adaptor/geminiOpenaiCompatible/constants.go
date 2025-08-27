package geminiOpenaiCompatible

import (
	"strings"

	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Google AI pricing: https://ai.google.dev/pricing
var ModelRatios = map[string]adaptor.ModelConfig{
	// Gemini Pro Models
	"gemini-pro":     {Ratio: 0.5 * ratio.MilliTokensUsd, CompletionRatio: 3},
	"gemini-1.0-pro": {Ratio: 0.5 * ratio.MilliTokensUsd, CompletionRatio: 3},

	// Gemma Models
	"gemma-2-2b-it":  {Ratio: 0.35 * ratio.MilliTokensUsd, CompletionRatio: 1.4},
	"gemma-2-9b-it":  {Ratio: 0.35 * ratio.MilliTokensUsd, CompletionRatio: 1.4},
	"gemma-2-27b-it": {Ratio: 0.35 * ratio.MilliTokensUsd, CompletionRatio: 1.4},
	"gemma-3-27b-it": {Ratio: 0.35 * ratio.MilliTokensUsd, CompletionRatio: 1.4},

	// Gemini 1.5 Flash Models
	"gemini-1.5-flash":    {Ratio: 0.075 * ratio.MilliTokensUsd, CompletionRatio: 4},
	"gemini-1.5-flash-8b": {Ratio: 0.0375 * ratio.MilliTokensUsd, CompletionRatio: 4},

	// Gemini 1.5 Pro Models
	"gemini-1.5-pro":              {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 4},
	"gemini-1.5-pro-experimental": {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 4},

	// Embedding Models
	"text-embedding-004": {Ratio: 0.00001 * ratio.MilliTokensUsd, CompletionRatio: 1},
	"aqa":                {Ratio: 1, CompletionRatio: 1},

	// Gemini 2.0 Flash Models
	"gemini-2.0-flash":                      {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 4},
	"gemini-2.0-flash-exp":                  {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 4},
	"gemini-2.0-flash-lite":                 {Ratio: 0.075 * ratio.MilliTokensUsd, CompletionRatio: 0.3 / 0.075},
	"gemini-2.0-flash-thinking-exp-01-21":   {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 4},
	"gemini-2.0-flash-exp-image-generation": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 4},

	// Gemini 2.0 Pro Models
	"gemini-2.0-pro-exp-02-05": {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 4},

	// Gemini 2.5 Flash Models
	"gemini-2.5-flash-lite-preview-06-17": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 4},
	"gemini-2.5-flash":                    {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2.5 / 0.3},
	"gemini-2.5-flash-preview-04-17":      {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2.5 / 0.3},
	"gemini-2.5-flash-preview-05-20":      {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2.5 / 0.3},
	"gemini-2.5-flash-image-preview":      {Ratio: 0.3 * ratio.MilliTokensUsd, CompletionRatio: 2.5 / 0.3, ImagePriceUsd: 0.039 * ratio.ImageUsdPerPic},

	// Gemini 2.5 Pro Models
	"gemini-2.5-pro":               {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 8},
	"gemini-2.5-pro-exp-03-25":     {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 8},
	"gemini-2.5-pro-preview-05-06": {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 8},
	"gemini-2.5-pro-preview-06-05": {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 8},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)

const (
	ModalityText  = "TEXT"
	ModalityImage = "IMAGE"
)

// GetModelModalities returns the modalities of the model.
func GetModelModalities(model string) []string {
	if strings.Contains(model, "-image-generation") {
		return []string{ModalityText, ModalityImage}
	}

	// Until 2025-03-26, the following models do not accept the responseModalities field
	if model == "aqa" ||
		strings.HasPrefix(model, "gemini-2.5") ||
		strings.HasPrefix(model, "gemma") ||
		strings.HasPrefix(model, "text-embed") {
		return nil
	}

	return []string{ModalityText}
}
