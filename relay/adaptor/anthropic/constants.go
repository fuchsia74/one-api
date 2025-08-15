package anthropic

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
//
// https://www.anthropic.com/pricing#api
var ModelRatios = map[string]adaptor.ModelConfig{
	// Claude Instant Models
	"claude-instant-1.2": {Ratio: 0.8 * ratio.MilliTokensUsd, CompletionRatio: 3.0, CachedInputRatio: 0.08 * ratio.MilliTokensUsd, CacheWrite5mRatio: 1.0 * ratio.MilliTokensUsd, CacheWrite1hRatio: 1.6 * ratio.MilliTokensUsd},

	// Claude 2 Models
	"claude-2.0": {Ratio: 8 * ratio.MilliTokensUsd, CompletionRatio: 3.0, CachedInputRatio: 0.8 * ratio.MilliTokensUsd, CacheWrite5mRatio: 10 * ratio.MilliTokensUsd, CacheWrite1hRatio: 16 * ratio.MilliTokensUsd},
	"claude-2.1": {Ratio: 8 * ratio.MilliTokensUsd, CompletionRatio: 3.0, CachedInputRatio: 0.8 * ratio.MilliTokensUsd, CacheWrite5mRatio: 10 * ratio.MilliTokensUsd, CacheWrite1hRatio: 16 * ratio.MilliTokensUsd},

	// Claude 3 Haiku Models
	"claude-3-haiku-20240307":   {Ratio: 0.25 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.025 * ratio.MilliTokensUsd, CacheWrite5mRatio: 0.3125 * ratio.MilliTokensUsd, CacheWrite1hRatio: 0.5 * ratio.MilliTokensUsd},
	"claude-3-5-haiku-latest":   {Ratio: 0.8 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.08 * ratio.MilliTokensUsd, CacheWrite5mRatio: 1.0 * ratio.MilliTokensUsd, CacheWrite1hRatio: 1.6 * ratio.MilliTokensUsd},
	"claude-3-5-haiku-20241022": {Ratio: 0.8 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.08 * ratio.MilliTokensUsd, CacheWrite5mRatio: 1.0 * ratio.MilliTokensUsd, CacheWrite1hRatio: 1.6 * ratio.MilliTokensUsd},

	// Claude 3 Sonnet Models
	"claude-3-sonnet-20240229":   {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.3 * ratio.MilliTokensUsd, CacheWrite5mRatio: 3.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 6 * ratio.MilliTokensUsd},
	"claude-3-5-sonnet-latest":   {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.3 * ratio.MilliTokensUsd, CacheWrite5mRatio: 3.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 6 * ratio.MilliTokensUsd},
	"claude-3-5-sonnet-20240620": {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.3 * ratio.MilliTokensUsd, CacheWrite5mRatio: 3.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 6 * ratio.MilliTokensUsd},
	"claude-3-5-sonnet-20241022": {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.3 * ratio.MilliTokensUsd, CacheWrite5mRatio: 3.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 6 * ratio.MilliTokensUsd},
	"claude-3-7-sonnet-latest":   {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 1.5 * ratio.MilliTokensUsd, CacheWrite5mRatio: 18.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 30 * ratio.MilliTokensUsd},
	"claude-3-7-sonnet-20250219": {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 1.5 * ratio.MilliTokensUsd, CacheWrite5mRatio: 18.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 30 * ratio.MilliTokensUsd},

	// Claude 3 Opus Models
	"claude-3-opus-20240229": {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 1.5 * ratio.MilliTokensUsd, CacheWrite5mRatio: 18.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 30 * ratio.MilliTokensUsd},

	// Claude 4 Opus Models
	"claude-opus-4-20250514":   {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 1.5 * ratio.MilliTokensUsd, CacheWrite5mRatio: 18.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 30 * ratio.MilliTokensUsd},
	"claude-opus-4-1-20250805": {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 1.5 * ratio.MilliTokensUsd, CacheWrite5mRatio: 18.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 30 * ratio.MilliTokensUsd},

	// Claude 4 Sonnet Models
	"claude-sonnet-4-20250514": {Ratio: 3 * ratio.MilliTokensUsd, CompletionRatio: 5.0, CachedInputRatio: 0.3 * ratio.MilliTokensUsd, CacheWrite5mRatio: 3.75 * ratio.MilliTokensUsd, CacheWrite1hRatio: 6 * ratio.MilliTokensUsd},
}
