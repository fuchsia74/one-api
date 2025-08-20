package ratio

import (
	"fmt"
	"strings"

	"github.com/songquanpeng/one-api/common/logger"
)

const (
	QuotaPerUsd    = 500000 // $1 = 500,000 quota
	MilliTokensUsd = 0.5    // 0.000001 * 500000 = 0.5 quota per milli-token
	// ImageUsdPerPic equals QuotaPerUsd to eliminate legacy magic numbers.
	// Use: Ratio = usd_per_image * ImageUsdPerPic (backward-identical to (usd/0.002)*1000)
	ImageUsdPerPic   = QuotaPerUsd
	MilliTokensRmb   = 3.5   // 0.000007 * 500000 = 3.5 quota per milli-token
	ImageRmbPerPic   = 7000  // 0.014 * 500000 = 7000 quota per image
	MilliTokensYuan  = 3.5   // 0.000007 * 500000 = 3.5 quota per milli-token
	ImageYuanPerPic  = 7000  // 0.014 * 500000 = 7000 quota per image
	VideoUsdPerVideo = 50000 // 0.1 * 500000 = 50000 quota per video (typical $0.1 per video)
	TokensPerSec     = 10    // Video tokens per second for video generation models
)

// Note: ModelPrice has been moved to relay/adaptor/interface.go
// Each adapter now manages its own pricing independently

// Note: Channel-specific pricing has been moved to individual adapters
// Each adapter now implements GetDefaultModelPricing() method

// Note: Global ModelRatio and CompletionRatio have been REMOVED
// All pricing is now handled by the clean two-layer approach:
// 1. User custom ratio (channel-specific overrides)
// 2. Channel default ratio (adapter's default pricing via GetDefaultModelPricing())
//
// For new models, add them to the appropriate adapter's GetDefaultModelPricing() method.

// AudioRatio represents the price ratio between audio tokens and text tokens
var AudioRatio = map[string]float64{
	"gpt-4o-audio-preview":                 16,
	"gpt-4o-audio-preview-2024-12-17":      16,
	"gpt-4o-audio-preview-2024-10-01":      40,
	"gpt-4o-mini-audio-preview":            10 / 0.15,
	"gpt-4o-mini-audio-preview-2024-12-17": 10 / 0.15,
	"gpt-4o-transcribe":                    6 / 2.5,
	"gpt-4o-mini-transcribe":               3 / 1.25,
}

// GetAudioPromptRatio returns the audio prompt ratio for the given model.
func GetAudioPromptRatio(actualModelName string) float64 {
	var v float64
	if ratio, ok := AudioRatio[actualModelName]; ok {
		v = ratio
	} else {
		v = 16
	}

	return v
}

// AudioCompletionRatio is the completion ratio for audio models.
var AudioCompletionRatio = map[string]float64{
	"whisper-1":                            0,
	"gpt-4o-audio-preview":                 2,
	"gpt-4o-audio-preview-2024-12-17":      2,
	"gpt-4o-audio-preview-2024-10-01":      2,
	"gpt-4o-mini-audio-preview":            2,
	"gpt-4o-mini-audio-preview-2024-12-17": 2,
}

// GetAudioCompletionRatio returns the completion ratio for audio models.
func GetAudioCompletionRatio(actualModelName string) float64 {
	var v float64
	if ratio, ok := AudioCompletionRatio[actualModelName]; ok {
		v = ratio
	} else {
		v = 2
	}

	return v
}

// AudioTokensPerSecond is the number of audio tokens per second for each model.
var AudioPromptTokensPerSecond = map[string]float64{
	// Whisper API price is $0.0001/sec. One-api's historical ratio is 15,
	// corresponding to $0.03/kilo_tokens.
	// So the audio tokens per second is 15 / 0.03 = 500.
	// But we use 10 here to be conservative.
	"whisper-1":                            10,
	"gpt-4o-audio-preview":                 10,
	"gpt-4o-audio-preview-2024-12-17":      10,
	"gpt-4o-audio-preview-2024-10-01":      10,
	"gpt-4o-mini-audio-preview":            10,
	"gpt-4o-mini-audio-preview-2024-12-17": 10,
	"gpt-4o-transcribe":                    10,
	"gpt-4o-mini-transcribe":               10,
	"gpt-4o-mini-tts":                      10,
}

// GetAudioPromptTokensPerSecond returns the number of audio tokens per second
// for the given model.
func GetAudioPromptTokensPerSecond(actualModelName string) float64 {
	var v float64
	if ratio, ok := AudioPromptTokensPerSecond[actualModelName]; ok {
		v = ratio
	} else {
		v = 10
	}

	return v
}

// LEGACY FUNCTIONS - These functions are deprecated and should only be used for legacy compatibility.
// New code should use the two-layer approach directly:
// 1. Check channel-specific overrides first
// 2. Use adapter.GetModelRatio() and adapter.GetCompletionRatio() as fallback

// GetModelRatio is used to get the model ratio for a given model name and channel type.
func GetModelRatio(actualModelName string, channelType int) float64 {
	return GetModelRatioWithChannel(actualModelName, channelType, nil)
}

// GetModelRatioWithChannel - LEGACY FUNCTION
// This function is deprecated and should only be used for legacy compatibility.
// New code should use the three-layer approach directly via pricing.GetModelRatioWithThreeLayers()
//
// This function now implements the three-layer system for backward compatibility:
// 1. Channel-specific overrides (highest priority)
// 2. Adapter default pricing (second priority)
// 3. Global pricing fallback (third priority)
// 4. Legacy special models and final default (lowest priority)
func GetModelRatioWithChannel(actualModelName string, channelType int, channelModelRatio map[string]float64) float64 {
	// Normalize model names for internet variants
	if strings.HasPrefix(actualModelName, "qwen-") &&
		strings.HasSuffix(actualModelName, "-internet") {
		actualModelName = strings.TrimSuffix(actualModelName, "-internet")
	}
	if strings.HasPrefix(actualModelName, "command-") &&
		strings.HasSuffix(actualModelName, "-internet") {
		actualModelName = strings.TrimSuffix(actualModelName, "-internet")
	}

	nameWithChannel := fmt.Sprintf("%s(%d)", actualModelName, channelType)

	// Layer 1: Check channel-specific pricing if provided
	if channelModelRatio != nil {
		for _, targetName := range []string{nameWithChannel, actualModelName} {
			if ratio, ok := channelModelRatio[targetName]; ok {
				return ratio
			}
		}
	}

	// Layer 2: Try adapter pricing (if available)
	// Note: We can't import relay package here due to import cycles,
	// so we skip this layer in the legacy function.
	// Modern code should use pricing.GetModelRatioWithThreeLayers() instead.

	// Layer 3: Try global pricing (if available)
	// Note: We can't import pricing package here due to import cycles,
	// so we skip this layer in the legacy function.
	// Modern code should use pricing.GetModelRatioWithThreeLayers() instead.

	// Layer 4: Legacy fallback - check remaining global maps for special models
	for _, targetName := range []string{nameWithChannel, actualModelName} {
		// Check audio models (still needed for special audio pricing)
		if ratio, ok := AudioRatio[targetName]; ok {
			return ratio
		}
	}

	// Final fallback: reasonable default
	logger.Logger.Warn(fmt.Sprintf("Legacy pricing fallback for model %s (channel type: %d), consider migrating to three-layer pricing", actualModelName, channelType))
	return 2.5 * MilliTokensUsd
}

// GetCompletionRatio returns the completion ratio for the given model name and channel type.
func GetCompletionRatio(name string, channelType int) float64 {
	return GetCompletionRatioWithChannel(name, channelType, nil)
}

// GetCompletionRatioWithChannel - LEGACY FUNCTION
// This function is deprecated and should only be used for legacy compatibility.
// New code should use the three-layer approach directly via pricing.GetCompletionRatioWithThreeLayers()
//
// This function now implements the three-layer system for backward compatibility:
// 1. Channel-specific overrides (highest priority)
// 2. Adapter default pricing (second priority) - skipped due to import cycles
// 3. Global pricing fallback (third priority) - skipped due to import cycles
// 4. Legacy special models and final default (lowest priority)
func GetCompletionRatioWithChannel(name string, channelType int, channelCompletionRatio map[string]float64) float64 {
	if strings.HasPrefix(name, "qwen-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	model := fmt.Sprintf("%s(%d)", name, channelType)

	name = strings.TrimPrefix(name, "openai/")

	// Layer 1: Check channel-specific pricing if provided
	if channelCompletionRatio != nil {
		for _, targetName := range []string{model, name} {
			// first try the model name
			if ratio, ok := channelCompletionRatio[targetName]; ok {
				return ratio
			}

			// then try the model name without some special prefix
			normalizedTargetName := strings.TrimPrefix(targetName, "openai/")
			if ratio, ok := channelCompletionRatio[normalizedTargetName]; ok {
				return ratio
			}
		}
	}

	// Layer 2: Try adapter pricing (if available)
	// Note: We can't import relay package here due to import cycles,
	// so we skip this layer in the legacy function.
	// Modern code should use pricing.GetCompletionRatioWithThreeLayers() instead.

	// Layer 3: Try global pricing (if available)
	// Note: We can't import pricing package here due to import cycles,
	// so we skip this layer in the legacy function.
	// Modern code should use pricing.GetCompletionRatioWithThreeLayers() instead.

	// Layer 4: Legacy fallback - check remaining global maps for special models
	for _, targetName := range []string{model, name} {
		// Check audio completion ratios (still needed for special audio pricing)
		if ratio, ok := AudioCompletionRatio[targetName]; ok {
			return ratio
		}
	}

	// Final fallback: reasonable default
	logger.Logger.Warn(fmt.Sprintf("completion ratio not found for model: %s (channel type: %d), using default value 1, consider migrating to three-layer pricing", name, channelType))
	return 1
}

// LEGACY ADMIN FUNCTIONS - These are kept for backward compatibility with admin interfaces
// that may still reference the global pricing maps

// ModelRatio2JSONString returns an empty JSON object since global ModelRatio has been removed
func ModelRatio2JSONString() string {
	// Return empty object since global ModelRatio has been removed
	return "{}"
}

// UpdateModelRatioByJSONString is a no-op since global ModelRatio has been removed
func UpdateModelRatioByJSONString(jsonStr string) error {
	// No-op since global ModelRatio has been removed
	logger.Logger.Warn("UpdateModelRatioByJSONString called but global ModelRatio has been removed. Use adapter-specific pricing instead.")
	return nil
}

// CompletionRatio2JSONString returns an empty JSON object since global CompletionRatio has been removed
func CompletionRatio2JSONString() string {
	// Return empty object since global CompletionRatio has been removed
	return "{}"
}

// UpdateCompletionRatioByJSONString is a no-op since global CompletionRatio has been removed
func UpdateCompletionRatioByJSONString(jsonStr string) error {
	// No-op since global CompletionRatio has been removed
	logger.Logger.Warn("UpdateCompletionRatioByJSONString called but global CompletionRatio has been removed. Use adapter-specific pricing instead.")
	return nil
}

// AddNewMissingRatio returns the input unchanged since global pricing has been removed
func AddNewMissingRatio(oldRatio string) string {
	// Return unchanged since global pricing has been removed
	return oldRatio
}

// Note: All channel-specific pricing functions have been moved to individual adapters
// Each adapter now implements its own GetDefaultModelPricing() method
