package openai

import (
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
)

// ModelRatios contains all supported models and their pricing ratios
// Model list is derived from the keys of this map, eliminating redundancy
// Based on OpenAI pricing: https://platform.openai.com/docs/pricing
var ModelRatios = map[string]adaptor.ModelConfig{
	// GPT-3.5 Models (Standard)
	"gpt-3.5-turbo":          {Ratio: 0.5 * ratio.MilliTokensUsd, CompletionRatio: 1.5 / 0.5},
	"gpt-3.5-turbo-0301":     {Ratio: 1.5 * ratio.MilliTokensUsd, CompletionRatio: 2.0 / 1.5},
	"gpt-3.5-turbo-0613":     {Ratio: 1.5 * ratio.MilliTokensUsd, CompletionRatio: 2.0 / 1.5},
	"gpt-3.5-turbo-1106":     {Ratio: 1.0 * ratio.MilliTokensUsd, CompletionRatio: 2.0 / 1.0},
	"gpt-3.5-turbo-0125":     {Ratio: 0.5 * ratio.MilliTokensUsd, CompletionRatio: 1.5 / 0.5},
	"gpt-3.5-turbo-16k":      {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0 / 3.0},
	"gpt-3.5-turbo-16k-0613": {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0 / 3.0},
	"gpt-3.5-turbo-instruct": {Ratio: 1.5 * ratio.MilliTokensUsd, CompletionRatio: 2.0 / 1.5},

	// GPT-4 Models (Standard)
	"gpt-4":                  {Ratio: 30.0 * ratio.MilliTokensUsd, CompletionRatio: 60.0 / 30.0},
	"gpt-4-0314":             {Ratio: 30.0 * ratio.MilliTokensUsd, CompletionRatio: 60.0 / 30.0},
	"gpt-4-0613":             {Ratio: 30.0 * ratio.MilliTokensUsd, CompletionRatio: 60.0 / 30.0},
	"gpt-4-32k":              {Ratio: 60.0 * ratio.MilliTokensUsd, CompletionRatio: 120.0 / 60.0},
	"gpt-4-32k-0314":         {Ratio: 60.0 * ratio.MilliTokensUsd, CompletionRatio: 120.0 / 60.0},
	"gpt-4-32k-0613":         {Ratio: 60.0 * ratio.MilliTokensUsd, CompletionRatio: 120.0 / 60.0},
	"gpt-4-1106-preview":     {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 30.0 / 10.0},
	"gpt-4-0125-preview":     {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 30.0 / 10.0},
	"gpt-4-turbo-preview":    {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 30.0 / 10.0},
	"gpt-4-vision-preview":   {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 30.0 / 10.0},
	"gpt-4-turbo":            {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 30.0 / 10.0},
	"gpt-4-turbo-2024-04-09": {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 30.0 / 10.0},

	// GPT-4o Models
	"gpt-4o":                               {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 1.25 * ratio.MilliTokensUsd},
	"gpt-4o-2024-05-13":                    {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 3.0},
	"gpt-4o-2024-08-06":                    {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 1.25 * ratio.MilliTokensUsd},
	"gpt-4o-2024-11-20":                    {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 1.25 * ratio.MilliTokensUsd},
	"gpt-4o-mini":                          {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.075 * ratio.MilliTokensUsd},
	"gpt-4o-mini-2024-07-18":               {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.075 * ratio.MilliTokensUsd},
	"gpt-4o-mini-audio-preview":            {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"gpt-4o-mini-audio-preview-2024-12-17": {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"gpt-4o-audio-preview":                 {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"gpt-4o-audio-preview-2024-12-17":      {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"gpt-4o-audio-preview-2024-10-01":      {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"gpt-4o-audio-preview-2025-06-03":      {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0},

	// Realtime Models
	"gpt-4o-realtime-preview":                 {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 2.5 * ratio.MilliTokensUsd},
	"gpt-4o-realtime-preview-2025-06-03":      {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 2.5 * ratio.MilliTokensUsd},
	"gpt-4o-mini-realtime-preview":            {Ratio: 0.6 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.3 * ratio.MilliTokensUsd},
	"gpt-4o-mini-realtime-preview-2024-12-17": {Ratio: 0.6 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.3 * ratio.MilliTokensUsd},

	// GPT-4.5 Models
	"gpt-4.5-preview":            {Ratio: 75.0 * ratio.MilliTokensUsd, CompletionRatio: 2.0},
	"gpt-4.5-preview-2025-02-27": {Ratio: 75.0 * ratio.MilliTokensUsd, CompletionRatio: 2.0},

	// GPT-4.1 Models
	"gpt-4.1":                 {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.5 * ratio.MilliTokensUsd},
	"gpt-4.1-2025-04-14":      {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.5 * ratio.MilliTokensUsd},
	"gpt-4.1-mini":            {Ratio: 0.4 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.1 * ratio.MilliTokensUsd},
	"gpt-4.1-mini-2025-04-14": {Ratio: 0.4 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.1 * ratio.MilliTokensUsd},
	"gpt-4.1-nano":            {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.025 * ratio.MilliTokensUsd},
	"gpt-4.1-nano-2025-04-14": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.025 * ratio.MilliTokensUsd},

	// GPT-5 Models
	"gpt-5-chat-latest":     {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 10 / 1.25, CachedInputRatio: 0.125 * ratio.MilliTokensUsd},
	"gpt-5-codex":           {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 10 / 1.25, CachedInputRatio: 0.125 * ratio.MilliTokensUsd},
	"gpt-5-pro":             {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 120 / 15},
	"gpt-5-pro-2025-10-06":  {Ratio: 15 * ratio.MilliTokensUsd, CompletionRatio: 120 / 15},
	"gpt-5":                 {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 10 / 1.25, CachedInputRatio: 0.125 * ratio.MilliTokensUsd},
	"gpt-5-2025-08-07":      {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 10 / 1.25, CachedInputRatio: 0.125 * ratio.MilliTokensUsd},
	"gpt-5-mini":            {Ratio: 0.25 * ratio.MilliTokensUsd, CompletionRatio: 2 / 0.25, CachedInputRatio: 0.025 * ratio.MilliTokensUsd},
	"gpt-5-mini-2025-08-07": {Ratio: 0.25 * ratio.MilliTokensUsd, CompletionRatio: 2 / 0.25, CachedInputRatio: 0.025 * ratio.MilliTokensUsd},
	"gpt-5-nano":            {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 0.4 / 0.05, CachedInputRatio: 0.005 * ratio.MilliTokensUsd},
	"gpt-5-nano-2025-08-07": {Ratio: 0.05 * ratio.MilliTokensUsd, CompletionRatio: 0.4 / 0.05, CachedInputRatio: 0.005 * ratio.MilliTokensUsd},

	// o1 Models
	"o1":                    {Ratio: 15.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 7.5 * ratio.MilliTokensUsd},
	"o1-2024-12-17":         {Ratio: 15.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 7.5 * ratio.MilliTokensUsd},
	"o1-pro":                {Ratio: 150.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"o1-pro-2025-03-19":     {Ratio: 150.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"o1-preview":            {Ratio: 15.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 7.5 * ratio.MilliTokensUsd},
	"o1-preview-2024-09-12": {Ratio: 15.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 7.5 * ratio.MilliTokensUsd},
	"o1-mini":               {Ratio: 1.1 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.55 * ratio.MilliTokensUsd},
	"o1-mini-2024-09-12":    {Ratio: 1.1 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.55 * ratio.MilliTokensUsd},

	// o3 Models
	"o3":                          {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.5 * ratio.MilliTokensUsd},
	"o3-2025-04-16":               {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.5 * ratio.MilliTokensUsd},
	"o3-mini":                     {Ratio: 1.1 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.55 * ratio.MilliTokensUsd},
	"o3-mini-2025-01-31":          {Ratio: 1.1 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.55 * ratio.MilliTokensUsd},
	"o3-pro":                      {Ratio: 20.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"o3-pro-2025-06-10":           {Ratio: 20.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"o3-deep-research":            {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 2.5 * ratio.MilliTokensUsd},
	"o3-deep-research-2025-06-26": {Ratio: 10.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 2.5 * ratio.MilliTokensUsd},

	// o4 Models
	"o4-mini":                          {Ratio: 1.1 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.275 * ratio.MilliTokensUsd},
	"o4-mini-2025-04-16":               {Ratio: 1.1 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.275 * ratio.MilliTokensUsd},
	"o4-mini-deep-research":            {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.5 * ratio.MilliTokensUsd},
	"o4-mini-deep-research-2025-06-26": {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.5 * ratio.MilliTokensUsd},

	// Codex Models
	"codex-mini-latest": {Ratio: 1.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0, CachedInputRatio: 0.375 * ratio.MilliTokensUsd},

	// Search Models
	"gpt-4o-mini-search-preview":            {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"gpt-4o-mini-search-preview-2025-03-11": {Ratio: 0.15 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"gpt-4o-search-preview":                 {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"gpt-4o-search-preview-2025-03-11":      {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0},

	// Computer Use Models
	"computer-use-preview":            {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0},
	"computer-use-preview-2025-03-11": {Ratio: 3.0 * ratio.MilliTokensUsd, CompletionRatio: 4.0},

	// ChatGPT Models (Standard)
	"chatgpt-4o-latest": {Ratio: 5.0 * ratio.MilliTokensUsd, CompletionRatio: 15.0 / 5.0},

	// Embedding Models (Standard)
	"text-embedding-ada-002": {Ratio: 0.1 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"text-embedding-3-small": {Ratio: 0.02 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"text-embedding-3-large": {Ratio: 0.13 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"text-curie-001":         {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"text-babbage-001":       {Ratio: 0.5 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"text-ada-001":           {Ratio: 0.4 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"text-davinci-002":       {Ratio: 20.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"text-davinci-003":       {Ratio: 20.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"text-davinci-edit-001":  {Ratio: 20.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"davinci-002":            {Ratio: 2.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0},
	"babbage-002":            {Ratio: 0.4 * ratio.MilliTokensUsd, CompletionRatio: 1.0},

	// Moderation Models
	"text-moderation-latest":     {Ratio: 0.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // Free
	"text-moderation-stable":     {Ratio: 0.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // Free
	"text-moderation-007":        {Ratio: 0.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // Free
	"omni-moderation-latest":     {Ratio: 0.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // Free
	"omni-moderation-2024-09-26": {Ratio: 0.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // Free

	// Audio Models
	"whisper-1": {Ratio: 6.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $0.006 per minute

	// TTS Models
	"tts-1":                  {Ratio: 15.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $15.00 per 1M characters
	"tts-1-1106":             {Ratio: 15.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $15.00 per 1M characters
	"tts-1-hd":               {Ratio: 30.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $30.00 per 1M characters
	"tts-1-hd-1106":          {Ratio: 30.0 * ratio.MilliTokensUsd, CompletionRatio: 1.0}, // $30.00 per 1M characters
	"gpt-4o-transcribe":      {Ratio: 2.5 * ratio.MilliTokensUsd, CompletionRatio: 4.0},  // $2.50 input, $10.00 output per 1M tokens
	"gpt-4o-mini-transcribe": {Ratio: 1.25 * ratio.MilliTokensUsd, CompletionRatio: 4.0}, // $1.25 input, $5.00 output per 1M tokens
	"gpt-4o-mini-tts":        {Ratio: 0.6 * ratio.MilliTokensUsd, CompletionRatio: 20.0}, // $0.60 input, $12.00 output per 1M tokens

	// Image Generation Models (Standard)
	// Policy: If a model is billed per image, set Ratio=0 and use ImagePriceUsd.
	// GPT Image 1 image generations are billed per image by size and fidelity tiers; tokens billed elsewhere do not apply to /images endpoints.
	"dall-e-2": {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.016},
	"dall-e-3": {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.04},
	// Base price for gpt-image-1 at 1024x1024 with low fidelity is $0.011; tier tables in ratio.ImageTierTables scale this.
	"gpt-image-1": {Ratio: 0, CompletionRatio: 1.0, ImagePriceUsd: 0.011},
}

// ModelList derived from ModelRatios for backward compatibility
var ModelList = adaptor.GetModelListFromPricing(ModelRatios)
