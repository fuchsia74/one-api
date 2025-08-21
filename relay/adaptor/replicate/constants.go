package replicate

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	ratio "github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on Replicate pricing: https://replicate.com/pricing
var ModelRatios = map[string]adaptor.ModelConfig{
	// -------------------------------------
	// Image Generation Models
	// -------------------------------------
	"black-forest-labs/flux-kontext-pro":            {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.04},  // $0.04 per image
	"black-forest-labs/flux-1.1-pro":                {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.04},  // $0.04 per image
	"black-forest-labs/flux-1.1-pro-ultra":          {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.06},  // $0.06 per image
	"black-forest-labs/flux-canny-dev":              {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.025}, // $0.025 per image
	"black-forest-labs/flux-canny-pro":              {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.05},  // $0.05 per image
	"black-forest-labs/flux-depth-dev":              {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.025}, // $0.025 per image
	"black-forest-labs/flux-depth-pro":              {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.05},  // $0.05 per image
	"black-forest-labs/flux-dev":                    {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.025}, // $0.025 per image
	"black-forest-labs/flux-dev-lora":               {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.032}, // $0.032 per image
	"black-forest-labs/flux-fill-dev":               {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.04},  // $0.04 per image
	"black-forest-labs/flux-fill-pro":               {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.05},  // $0.05 per image
	"black-forest-labs/flux-pro":                    {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.055}, // $0.055 per image
	"black-forest-labs/flux-redux-dev":              {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.025}, // $0.025 per image
	"black-forest-labs/flux-redux-schnell":          {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.003}, // $0.003 per image
	"black-forest-labs/flux-schnell":                {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.003}, // $0.003 per image
	"black-forest-labs/flux-schnell-lora":           {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.02},  // $0.02 per image
	"ideogram-ai/ideogram-v2":                       {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.08},  // $0.08 per image
	"ideogram-ai/ideogram-v2-turbo":                 {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.05},  // $0.05 per image
	"recraft-ai/recraft-v3":                         {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.04},  // $0.04 per image
	"recraft-ai/recraft-v3-svg":                     {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.08},  // $0.08 per image
	"stability-ai/stable-diffusion-3":               {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.035}, // $0.035 per image
	"stability-ai/stable-diffusion-3.5-large":       {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.065}, // $0.065 per image
	"stability-ai/stable-diffusion-3.5-large-turbo": {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.04},  // $0.04 per image
	"stability-ai/stable-diffusion-3.5-medium":      {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.035}, // $0.035 per image

	// -------------------------------------
	// Language Models
	// -------------------------------------
	"anthropic/claude-3.5-haiku":                {Ratio: 1.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $1.0/$5.0 per 1M tokens
	"anthropic/claude-3.5-sonnet":               {Ratio: 3.75 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $3.75/$18.75 per 1M tokens
	"anthropic/claude-3.7-sonnet":               {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $3.0/$15.0 per 1M tokens
	"deepseek-ai/deepseek-r1":                   {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},  // $10.0 per 1M tokens
	"ibm-granite/granite-20b-code-instruct-8k":  {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $0.1/$0.5 per 1M tokens
	"ibm-granite/granite-3.0-2b-instruct":       {Ratio: 0.03 * ratio.MilliTokensUsd, CompletionRatio: 8.33}, // $0.03/$0.25 per 1M tokens
	"ibm-granite/granite-3.0-8b-instruct":       {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"ibm-granite/granite-3.1-2b-instruct":       {Ratio: 0.03 * ratio.MilliTokensUsd, CompletionRatio: 8.33}, // $0.03/$0.25 per 1M tokens
	"ibm-granite/granite-3.1-8b-instruct":       {Ratio: 0.03 * ratio.MilliTokensUsd, CompletionRatio: 8.33}, // $0.03/$0.25 per 1M tokens
	"ibm-granite/granite-3.2-8b-instruct":       {Ratio: 0.03 * ratio.MilliTokensUsd, CompletionRatio: 8.33}, // $0.03/$0.25 per 1M tokens
	"ibm-granite/granite-8b-code-instruct-128k": {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"meta/llama-2-13b":                          {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $0.1/$0.5 per 1M tokens
	"meta/llama-2-13b-chat":                     {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 5.0},   // $0.1/$0.5 per 1M tokens
	"meta/llama-2-70b":                          {Ratio: 0.65 * ratio.MilliTokensUsd, CompletionRatio: 4.23}, // $0.65/$2.75 per 1M tokens
	"meta/llama-2-70b-chat":                     {Ratio: 0.65 * ratio.MilliTokensUsd, CompletionRatio: 4.23}, // $0.65/$2.75 per 1M tokens
	"meta/llama-2-7b":                           {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"meta/llama-2-7b-chat":                      {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"meta/meta-llama-3.1-405b-instruct":         {Ratio: 9.5 * ratio.MilliTokensUsd, CompletionRatio: 1.0},   // $9.5 per 1M tokens
	"meta/meta-llama-3-70b":                     {Ratio: 0.65 * ratio.MilliTokensUsd, CompletionRatio: 4.23}, // $0.65/$2.75 per 1M tokens
	"meta/meta-llama-3-70b-instruct":            {Ratio: 0.65 * ratio.MilliTokensUsd, CompletionRatio: 4.23}, // $0.65/$2.75 per 1M tokens
	"meta/meta-llama-3-8b":                      {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"meta/meta-llama-3-8b-instruct":             {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"mistralai/mistral-7b-instruct-v0.2":        {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens
	"mistralai/mistral-7b-v0.1":                 {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 5.0},  // $0.05/$0.25 per 1M tokens

	// -------------------------------------
	// Video Models (TODO: implement the adaptor)
	// -------------------------------------
	// "minimax/video-01": {Ratio: 1.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
