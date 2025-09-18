package common

import (
	"fmt"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/relay/model"
)

func LogQuota(quota int64) string {
	if config.DisplayInCurrencyEnabled {
		return fmt.Sprintf("＄%.6f quota", float64(quota)/config.QuotaPerUnit)
	} else {
		return fmt.Sprintf("%d point quota", quota)
	}
}

// ImageUsageInputTokensDetails represents the input tokens details for image usage
type ImageUsageInputTokensDetails struct {
	TextTokens  int `json:"text_tokens"`
	ImageTokens int `json:"image_tokens"`
}

// ImageUsage represents the usage information for image generation requests
type ImageUsage struct {
	TotalTokens        int                          `json:"total_tokens"`
	InputTokens        int                          `json:"input_tokens"`
	OutputTokens       int                          `json:"output_tokens"`
	InputTokensDetails ImageUsageInputTokensDetails `json:"input_tokens_details"`
}

// ConvertImageUsageToGeneralUsage converts ImageUsage to model.Usage
// with omitempty behavior - only non-zero values are included in the result.
// This is useful for providers like xAI that don't provide all token fields,
// allowing the omitempty JSON tags in model.Usage to exclude zero values.
func ConvertImageUsageToGeneralUsage(imageUsage *ImageUsage) *model.Usage {
	if imageUsage == nil {
		return nil
	}

	usage := &model.Usage{}

	// Only set non-zero values to leverage omitempty behavior
	if imageUsage.InputTokens > 0 {
		usage.PromptTokens = imageUsage.InputTokens
	}

	if imageUsage.OutputTokens > 0 {
		usage.CompletionTokens = imageUsage.OutputTokens
	}

	if imageUsage.TotalTokens > 0 {
		usage.TotalTokens = imageUsage.TotalTokens
	}

	// Only create PromptTokensDetails if we have non-zero token details
	if imageUsage.InputTokensDetails.TextTokens > 0 || imageUsage.InputTokensDetails.ImageTokens > 0 {
		usage.PromptTokensDetails = &model.UsagePromptTokensDetails{}

		if imageUsage.InputTokensDetails.TextTokens > 0 {
			usage.PromptTokensDetails.TextTokens = imageUsage.InputTokensDetails.TextTokens
		}

		if imageUsage.InputTokensDetails.ImageTokens > 0 {
			usage.PromptTokensDetails.ImageTokens = imageUsage.InputTokensDetails.ImageTokens
		}
	}

	return usage
}
