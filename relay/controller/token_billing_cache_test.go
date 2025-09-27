package controller

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/channeltype"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/pricing"
)

// absDiffI64 returns absolute difference for int64
func absDiffI64(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}

// Test that changing cached prompt tokens only affects input costs, not completion costs.
// We validate by computing the expected delta between cached and non-cached scenarios
// using the adapter's cached-input pricing, and ensure postConsumeQuota matches it (within rounding tolerance).
func TestPostConsumeQuota_OutputPricingIndependentOfCache(t *testing.T) {
	// Arrange
	modelName := "gpt-4o" // has explicit cached input pricing in OpenAI adapter
	channelType := channeltype.OpenAI
	adaptor := relay.GetAdaptor(channelType)
	if adaptor == nil {
		t.Fatalf("nil adaptor for channel %d", channelType)
	}
	modelRatio := adaptor.GetModelRatio(modelName)
	if modelRatio <= 0 {
		t.Fatalf("unexpected model ratio: %v", modelRatio)
	}
	// Resolve effective cached/input ratios at our prompt token scale (for tier handling)
	// Use a large token count to minimize rounding effects from ceil()
	promptTokens := 1_000_000
	completionTokens := 500_000
	eff := pricing.ResolveEffectivePricing(modelName, promptTokens, adaptor)
	// Prices per token (quota units per token)
	groupRatio := 1.0
	normalInputPrice := modelRatio * groupRatio
	cachedInputPrice := normalInputPrice
	if eff.CachedInputRatio < 0 {
		cachedInputPrice = 0
	} else if eff.CachedInputRatio > 0 {
		cachedInputPrice = eff.CachedInputRatio * groupRatio
	}

	// Meta and request
	// Use TokenId=0 to disable DB writes in billing during tests
	meta := &metalib.Meta{
		ChannelType: channelType,
		ChannelId:   1,
		TokenId:     0,
		UserId:      1,
		TokenName:   "test-token",
		StartTime:   time.Now(),
		IsStream:    false,
	}
	req := &relaymodel.GeneralOpenAIRequest{Model: modelName}

	// Case A: No cache
	usageNoCache := &relaymodel.Usage{PromptTokens: promptTokens, CompletionTokens: completionTokens}
	quotaNoCache := postConsumeQuota(context.Background(), usageNoCache, meta, req, 0, 0, 0, modelRatio, groupRatio, false, nil)

	// Case B: Some cached prompt tokens (e.g., 60%)
	cachedPrompt := int(float64(promptTokens) * 0.6)
	usageCached := &relaymodel.Usage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		PromptTokensDetails: &relaymodel.UsagePromptTokensDetails{
			CachedTokens: cachedPrompt,
		},
	}
	quotaCached := postConsumeQuota(context.Background(), usageCached, meta, req, 0, 0, 0, modelRatio, groupRatio, false, nil)

	// Expected delta arises only from input pricing change on cached tokens
	// Base prompt tokens: promptTokens. With caching, cachedPrompt tokens charged at cachedInputPrice instead of normalInputPrice.
	// Delta = cachedPrompt*(cachedInputPrice - normalInputPrice), then ceil() may shift by up to 1 token in each call.
	expectedDelta := int64(math.Ceil(float64(cachedPrompt) * (cachedInputPrice - normalInputPrice)))
	actualDelta := quotaCached - quotaNoCache

	// Allow for rounding differences (ceil applied twice) and potential +1 guard path; keep tight tolerance.
	if absDiffI64(actualDelta, expectedDelta) > 2 {
		t.Fatalf("unexpected quota delta: got %d, want ~%d (±2). no-cache=%d cached=%d", actualDelta, expectedDelta, quotaNoCache, quotaCached)
	}
}

// Test that cache-write tokens only affect input buckets and never the output price.
func TestPostConsumeQuota_CacheWriteDoesNotAffectOutput(t *testing.T) {
	modelName := "gpt-4o"
	channelType := channeltype.OpenAI
	adaptor := relay.GetAdaptor(channelType)
	if adaptor == nil {
		t.Fatalf("nil adaptor for channel %d", channelType)
	}
	modelRatio := adaptor.GetModelRatio(modelName)
	groupRatio := 1.0

	// Use large counts to reduce rounding influence
	promptTokens := 1_000_000
	completionTokens := 500_000
	write5m := 200_000 // 20% of prompt tokens written to cache window

	eff := pricing.ResolveEffectivePricing(modelName, promptTokens, adaptor)
	normalInputPrice := modelRatio * groupRatio
	write5mPrice := normalInputPrice
	if eff.CacheWrite5mRatio < 0 {
		write5mPrice = 0
	} else if eff.CacheWrite5mRatio > 0 {
		write5mPrice = eff.CacheWrite5mRatio * groupRatio
	}

	// Use TokenId=0 to disable DB writes in billing during tests
	meta := &metalib.Meta{ChannelType: channelType, ChannelId: 1, TokenId: 0, UserId: 1, TokenName: "test-token", StartTime: time.Now()}
	req := &relaymodel.GeneralOpenAIRequest{Model: modelName}

	// Base: no cache writes
	usageBase := &relaymodel.Usage{PromptTokens: promptTokens, CompletionTokens: completionTokens}
	base := postConsumeQuota(context.Background(), usageBase, meta, req, 0, 0, 0, modelRatio, groupRatio, false, nil)

	// With write tokens
	usageWrite := &relaymodel.Usage{PromptTokens: promptTokens, CompletionTokens: completionTokens, CacheWrite5mTokens: write5m}
	withWrite := postConsumeQuota(context.Background(), usageWrite, meta, req, 0, 0, 0, modelRatio, groupRatio, false, nil)

	// Expected delta is purely input-side: write tokens shift from normalInputPrice to write5mPrice
	expectedDelta := int64(math.Ceil(float64(write5m) * (write5mPrice - normalInputPrice)))
	actualDelta := withWrite - base
	if absDiffI64(actualDelta, expectedDelta) > 2 {
		t.Fatalf("unexpected write quota delta: got %d, want ~%d (±2). base=%d withWrite=%d", actualDelta, expectedDelta, base, withWrite)
	}
}

// Response API: cached prompt details should not affect quota (and specifically not the output term)
func TestPostConsumeResponseAPIQuota_IgnoresCachedPromptDetails(t *testing.T) {
	modelName := "gpt-4o"
	channelType := channeltype.OpenAI
	adaptor := relay.GetAdaptor(channelType)
	if adaptor == nil {
		t.Fatalf("nil adaptor for channel %d", channelType)
	}
	modelRatio := adaptor.GetModelRatio(modelName)
	groupRatio := 1.0
	ratio := modelRatio * groupRatio

	promptTokens := 200_000
	completionTokens := 300_000

	// Use TokenId=0 to disable DB writes in billing during tests
	meta := &metalib.Meta{ChannelType: channelType, ChannelId: 1, TokenId: 0, UserId: 1, TokenName: "test-token", StartTime: time.Now()}
	// Minimal response API request
	respReq := &openai.ResponseAPIRequest{Model: modelName}

	// Base usage
	usageBase := &relaymodel.Usage{PromptTokens: promptTokens, CompletionTokens: completionTokens}
	base := postConsumeResponseAPIQuota(context.Background(), usageBase, meta, respReq, ratio, 0, modelRatio, groupRatio, nil)

	// With cached prompt details present (should not change anything for Response API)
	usageCached := &relaymodel.Usage{PromptTokens: promptTokens, CompletionTokens: completionTokens, PromptTokensDetails: &relaymodel.UsagePromptTokensDetails{CachedTokens: promptTokens}}
	withCache := postConsumeResponseAPIQuota(context.Background(), usageCached, meta, respReq, ratio, 0, modelRatio, groupRatio, nil)

	if base != withCache {
		t.Fatalf("Response API quota changed due to cached prompt details: base=%d withCache=%d", base, withCache)
	}
}
