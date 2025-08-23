package deepseek

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on official DeepSeek pricing: https://platform.deepseek.com/api-docs/pricing/
var ModelRatios = map[string]adaptor.ModelConfig{
	"deepseek-chat":     {Ratio: 0.27 * ratio.MilliTokensUsd, CompletionRatio: 1.1 / 0.27},
	"deepseek-reasoner": {Ratio: 0.55 * ratio.MilliTokensUsd, CompletionRatio: 2.19 / 0.55},
	// FIXME: SINCE 16:00 UTC Time on Sept 5th, 2025
	// 	"deepseek-chat": {Ratio: 0.56 * ratio.MilliTokensUsd,
	// 		CachedInputRatio: 0.07 * ratio.MilliTokensUsd,
	// 		CompletionRatio:  1.68 / 0.56},
	// 	"deepseek-reasoner": {Ratio: 0.56 * ratio.MilliTokensUsd,
	// 		CachedInputRatio: 0.07 * ratio.MilliTokensUsd,
	// 		CompletionRatio:  1.68 / 0.56},
}
