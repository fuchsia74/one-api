package controller

import (
	"context"
	"math"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/constant/role"
	"github.com/songquanpeng/one-api/relay/controller/validator"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/pricing"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func getAndValidateTextRequest(c *gin.Context, relayMode int) (*relaymodel.GeneralOpenAIRequest, error) {
	textRequest := &relaymodel.GeneralOpenAIRequest{}
	err := common.UnmarshalBodyReusable(c, textRequest)
	if err != nil {
		return nil, err
	}
	if relayMode == relaymode.Moderations && textRequest.Model == "" {
		textRequest.Model = "text-moderation-latest"
	}
	if relayMode == relaymode.Embeddings && textRequest.Model == "" {
		textRequest.Model = c.Param("model")
	}
	err = validator.ValidateTextRequest(textRequest, relayMode)
	if err != nil {
		return nil, err
	}
	return textRequest, nil
}

func getPromptTokens(ctx context.Context, textRequest *relaymodel.GeneralOpenAIRequest, relayMode int) int {
	switch relayMode {
	case relaymode.ChatCompletions:
		actualModel := textRequest.Model
		// video request
		if strings.HasPrefix(actualModel, "veo-") {
			return ratio.TokensPerSec * 8
		}

		// text request
		return openai.CountTokenMessages(ctx, textRequest.Messages, textRequest.Model)
	case relaymode.Completions:
		return openai.CountTokenInput(textRequest.Prompt, textRequest.Model)
	case relaymode.Moderations:
		return openai.CountTokenInput(textRequest.Input, textRequest.Model)
	case relaymode.Embeddings:
		// Use ParseInput to properly handle both string and array inputs
		inputs := textRequest.ParseInput()
		totalTokens := 0
		for _, input := range inputs {
			totalTokens += openai.CountTokenText(input, textRequest.Model)
		}
		return totalTokens
	case relaymode.Rerank:
		return openai.CountTokenInput(textRequest.Input, textRequest.Model)
	case relaymode.Edits:
		return openai.CountTokenInput(textRequest.Instruction, textRequest.Model)
	default:
		// Log error for unhandled relay modes that should have billing
		logger.Logger.Error("getPromptTokens: unhandled relay mode without billing logic",
			zap.Int("relayMode", relayMode),
			zap.String("model", textRequest.Model))
	}

	return 0
}

func getPreConsumedQuota(textRequest *relaymodel.GeneralOpenAIRequest, promptTokens int, ratio float64) int64 {
	preConsumedTokens := config.PreConsumedQuota + int64(promptTokens)
	if textRequest.MaxTokens != 0 {
		preConsumedTokens += int64(textRequest.MaxTokens)
	}

	baseQuota := int64(float64(preConsumedTokens) * ratio)

	// Add estimated structured output cost if using JSON schema
	// This ensures pre-consumption quota accounts for the additional structured output costs
	if textRequest.ResponseFormat != nil &&
		textRequest.ResponseFormat.Type == "json_schema" &&
		textRequest.ResponseFormat.JsonSchema != nil {
		// Estimate structured output cost based on max tokens (conservative approach)
		estimatedCompletionTokens := textRequest.MaxTokens
		if estimatedCompletionTokens == 0 {
			// If no max tokens specified, use a conservative estimate
			estimatedCompletionTokens = 1000
		}

		// Apply the same 25% multiplier used in post-consumption
		// Note: We can't get exact model ratio here easily, so use the base ratio as approximation
		estimatedStructuredCost := int64(float64(estimatedCompletionTokens) * 0.25 * ratio)
		baseQuota += estimatedStructuredCost

		logger.Logger.Debug("Pre-consumption: added estimated structured output cost",
			zap.Int64("structured_output_cost", estimatedStructuredCost))
	}

	return baseQuota
}

func preConsumeQuota(c *gin.Context, textRequest *relaymodel.GeneralOpenAIRequest, promptTokens int, ratio float64, meta *meta.Meta) (int64, *relaymodel.ErrorWithStatusCode) {
	preConsumedQuota := getPreConsumedQuota(textRequest, promptTokens, ratio)

	tokenQuota := c.GetInt64(ctxkey.TokenQuota)
	tokenQuotaUnlimited := c.GetBool(ctxkey.TokenQuotaUnlimited)
	userQuota, err := model.CacheGetUserQuota(c.Request.Context(), meta.UserId)
	if err != nil {
		return preConsumedQuota, openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota-preConsumedQuota < 0 {
		return preConsumedQuota, openai.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(meta.UserId, preConsumedQuota)
	if err != nil {
		return preConsumedQuota, openai.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota &&
		(tokenQuotaUnlimited || tokenQuota > 100*preConsumedQuota) {
		// in this case, we do not pre-consume quota
		// because the user and token have enough quota
		preConsumedQuota = 0
		logger.Logger.Info("user has enough quota, trusted and no need to pre-consume", zap.Int("user_id", meta.UserId), zap.Int64("user_quota", userQuota))
	}
	if preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(meta.TokenId, preConsumedQuota)
		if err != nil {
			return preConsumedQuota, openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}
	return preConsumedQuota, nil
}

func postConsumeQuota(ctx context.Context,
	usage *relaymodel.Usage,
	meta *meta.Meta,
	textRequest *relaymodel.GeneralOpenAIRequest,
	ratio float64,
	preConsumedQuota int64,
	modelRatio float64,
	groupRatio float64,
	systemPromptReset bool,
	channelCompletionRatio map[string]float64) (quota int64) {
	if usage == nil {
		logger.Logger.Error("usage is nil, which is unexpected")
		return
	}

	// Resolve completion ratio (three-layer) and apply tiered pricing + cached discounts
	pricingAdaptor := relay.GetAdaptor(meta.ChannelType)
	completionRatioResolved := pricing.GetCompletionRatioWithThreeLayers(textRequest.Model, channelCompletionRatio, pricingAdaptor)
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens

	// Tier resolution based on input token count (promptTokens)
	eff := pricing.ResolveEffectivePricing(textRequest.Model, promptTokens, pricingAdaptor)

	// Decide whether to use adapter's tiered base ratios or keep externally resolved ones
	usedModelRatio := modelRatio
	usedCompletionRatio := completionRatioResolved
	if pricingAdaptor != nil {
		defaultPricing := pricingAdaptor.GetDefaultModelPricing()
		if _, ok := defaultPricing[textRequest.Model]; ok {
			// If the adapter has native pricing for this model and modelRatio equals adaptor's base,
			// we consider that no channel override applied, so we can adopt tiered base ratios.
			adaptorBase := pricingAdaptor.GetModelRatio(textRequest.Model)
			if math.Abs(modelRatio-adaptorBase) < 1e-12 {
				usedModelRatio = eff.InputRatio
				// Derive completion ratio from eff if available
				baseComp := eff.OutputRatio
				if eff.InputRatio != 0 {
					baseComp = eff.OutputRatio / eff.InputRatio
				} else {
					baseComp = 1.0
				}
				usedCompletionRatio = baseComp
			}
		}
	}

	// Split cached vs non-cached tokens
	// https://platform.openai.com/docs/guides/prompt-caching
	cachedPrompt := 0
	if usage.PromptTokensDetails != nil {
		cachedPrompt = usage.PromptTokensDetails.CachedTokens
		if cachedPrompt < 0 {
			cachedPrompt = 0
		}
		if cachedPrompt > promptTokens {
			cachedPrompt = promptTokens
		}
	}
	nonCachedPrompt := promptTokens - cachedPrompt
	// No cached completion billing; all completion tokens are non-cached
	nonCachedCompletion := completionTokens

	// Base per-token prices (include group ratio)
	normalInputPrice := usedModelRatio * groupRatio
	normalOutputPrice := usedModelRatio * usedCompletionRatio * groupRatio

	// Cached per-token prices; negative means free; zero means no special discount
	cachedInputPrice := normalInputPrice
	if eff.CachedInputRatio < 0 {
		cachedInputPrice = 0
	} else if eff.CachedInputRatio > 0 {
		cachedInputPrice = eff.CachedInputRatio * groupRatio
	}
	// No separate cached completion price

	// Cache-write tokens (Claude prompt caching)
	write5m := usage.CacheWrite5mTokens
	write1h := usage.CacheWrite1hTokens
	if write5m < 0 {
		write5m = 0
	}
	if write1h < 0 {
		write1h = 0
	}
	// Prevent double-charging: remove write tokens from normal input bucket
	if write5m+write1h > nonCachedPrompt {
		// Clamp to avoid negative counts due to inconsistent upstream reporting
		writeExcess := write5m + write1h - nonCachedPrompt
		if write1h >= writeExcess {
			write1h -= writeExcess
		} else {
			writeExcess -= write1h
			write1h = 0
			if write5m >= writeExcess {
				write5m -= writeExcess
			} else {
				write5m = 0
			}
		}
		nonCachedPrompt = 0
	} else {
		nonCachedPrompt -= (write5m + write1h)
	}

	// Determine write prices (fall back to normal input price if not configured)
	write5mPrice := normalInputPrice
	if eff.CacheWrite5mRatio < 0 {
		write5mPrice = 0
	} else if eff.CacheWrite5mRatio > 0 {
		write5mPrice = eff.CacheWrite5mRatio * groupRatio
	}
	write1hPrice := normalInputPrice
	if eff.CacheWrite1hRatio < 0 {
		write1hPrice = 0
	} else if eff.CacheWrite1hRatio > 0 {
		write1hPrice = eff.CacheWrite1hRatio * groupRatio
	}

	cost := float64(nonCachedPrompt)*normalInputPrice + float64(cachedPrompt)*cachedInputPrice +
		float64(nonCachedCompletion)*normalOutputPrice +
		float64(write5m)*write5mPrice + float64(write1h)*write1hPrice

	quota = int64(math.Ceil(cost)) + usage.ToolsCost
	if (usedModelRatio*groupRatio) != 0 && quota <= 0 {
		quota = 1
	}

	totalTokens := promptTokens + completionTokens
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
	}
	// Use centralized detailed billing function to follow DRY principle
	quotaDelta := quota - preConsumedQuota
	billing.PostConsumeQuotaDetailed(billing.QuotaConsumeDetail{
		Ctx:                    ctx,
		TokenId:                meta.TokenId,
		QuotaDelta:             quotaDelta,
		TotalQuota:             quota,
		UserId:                 meta.UserId,
		ChannelId:              meta.ChannelId,
		PromptTokens:           promptTokens,
		CompletionTokens:       completionTokens,
		ModelRatio:             usedModelRatio,
		GroupRatio:             groupRatio,
		ModelName:              textRequest.Model,
		TokenName:              meta.TokenName,
		IsStream:               meta.IsStream,
		StartTime:              meta.StartTime,
		SystemPromptReset:      systemPromptReset,
		CompletionRatio:        usedCompletionRatio,
		ToolsCost:              usage.ToolsCost,
		CachedPromptTokens:     cachedPrompt,
		CachedCompletionTokens: 0,
	})

	return quota
}

func postConsumeQuotaWithTraceID(ctx context.Context, traceId string,
	usage *relaymodel.Usage,
	meta *meta.Meta,
	textRequest *relaymodel.GeneralOpenAIRequest,
	ratio float64,
	preConsumedQuota int64,
	modelRatio float64,
	groupRatio float64,
	systemPromptReset bool,
	channelCompletionRatio map[string]float64) (quota int64) {
	if usage == nil {
		logger.Logger.Error("usage is nil, which is unexpected")
		return
	}

	// Resolve completion ratio (three-layer) and apply tiered pricing + cached discounts
	pricingAdaptor := relay.GetAdaptor(meta.ChannelType)
	completionRatioResolved := pricing.GetCompletionRatioWithThreeLayers(textRequest.Model, channelCompletionRatio, pricingAdaptor)
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens

	eff := pricing.ResolveEffectivePricing(textRequest.Model, promptTokens, pricingAdaptor)

	usedModelRatio := modelRatio
	usedCompletionRatio := completionRatioResolved
	if pricingAdaptor != nil {
		defaultPricing := pricingAdaptor.GetDefaultModelPricing()
		if _, ok := defaultPricing[textRequest.Model]; ok {
			adaptorBase := pricingAdaptor.GetModelRatio(textRequest.Model)
			if math.Abs(modelRatio-adaptorBase) < 1e-12 {
				usedModelRatio = eff.InputRatio
				baseComp := eff.OutputRatio
				if eff.InputRatio != 0 {
					baseComp = eff.OutputRatio / eff.InputRatio
				} else {
					baseComp = 1.0
				}
				usedCompletionRatio = baseComp
			}
		}
	}

	cachedPrompt := 0
	if usage.PromptTokensDetails != nil {
		cachedPrompt = usage.PromptTokensDetails.CachedTokens
		if cachedPrompt < 0 {
			cachedPrompt = 0
		}
		if cachedPrompt > promptTokens {
			cachedPrompt = promptTokens
		}
	}
	nonCachedPrompt := promptTokens - cachedPrompt
	// No cached completion billing; all completion tokens are non-cached
	nonCachedCompletion := completionTokens

	normalInputPrice := usedModelRatio * groupRatio
	normalOutputPrice := usedModelRatio * usedCompletionRatio * groupRatio

	cachedInputPrice := normalInputPrice
	if eff.CachedInputRatio < 0 {
		cachedInputPrice = 0
	} else if eff.CachedInputRatio > 0 {
		cachedInputPrice = eff.CachedInputRatio * groupRatio
	}
	// No separate cached completion price
	// Cache-write tokens (Claude prompt caching)
	write5m := usage.CacheWrite5mTokens
	write1h := usage.CacheWrite1hTokens
	if write5m < 0 {
		write5m = 0
	}
	if write1h < 0 {
		write1h = 0
	}
	if write5m+write1h > nonCachedPrompt {
		writeExcess := write5m + write1h - nonCachedPrompt
		if write1h >= writeExcess {
			write1h -= writeExcess
		} else {
			writeExcess -= write1h
			write1h = 0
			if write5m >= writeExcess {
				write5m -= writeExcess
			} else {
				write5m = 0
			}
		}
		nonCachedPrompt = 0
	} else {
		nonCachedPrompt -= (write5m + write1h)
	}

	write5mPrice := normalInputPrice
	if eff.CacheWrite5mRatio < 0 {
		write5mPrice = 0
	} else if eff.CacheWrite5mRatio > 0 {
		write5mPrice = eff.CacheWrite5mRatio * groupRatio
	}
	write1hPrice := normalInputPrice
	if eff.CacheWrite1hRatio < 0 {
		write1hPrice = 0
	} else if eff.CacheWrite1hRatio > 0 {
		write1hPrice = eff.CacheWrite1hRatio * groupRatio
	}

	cost := float64(nonCachedPrompt)*normalInputPrice + float64(cachedPrompt)*cachedInputPrice +
		float64(nonCachedCompletion)*normalOutputPrice +
		float64(write5m)*write5mPrice + float64(write1h)*write1hPrice

	quota = int64(math.Ceil(cost)) + usage.ToolsCost
	if (usedModelRatio*groupRatio) != 0 && quota <= 0 {
		quota = 1
	}

	totalTokens := promptTokens + completionTokens
	if totalTokens == 0 {
		// in this case, must be some error happened
		// we cannot just return, because we may have to return the pre-consumed quota
		quota = 0
	}
	// Use centralized detailed billing function with explicit trace ID
	quotaDelta := quota - preConsumedQuota
	billing.PostConsumeQuotaDetailedWithTraceID(ctx, traceId, meta.TokenId, quotaDelta, quota, meta.UserId, meta.ChannelId,
		promptTokens, completionTokens, usedModelRatio, groupRatio, textRequest.Model, meta.TokenName,
		meta.IsStream, meta.StartTime, systemPromptReset, usedCompletionRatio, usage.ToolsCost,
		cachedPrompt, 0) // Set cachedCompletion to 0

	return quota
}

func isErrorHappened(meta *meta.Meta, resp *http.Response) bool {
	if resp == nil {
		if meta.ChannelType == channeltype.AwsClaude {
			return false
		}
		return true
	}
	if resp.StatusCode != http.StatusOK &&
		// replicate return 201 to create a task
		resp.StatusCode != http.StatusCreated {
		return true
	}
	if meta.ChannelType == channeltype.DeepL {
		// skip stream check for deepl
		return false
	}

	if meta.IsStream && strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") &&
		// Even if stream mode is enabled, replicate will first return a task info in JSON format,
		// requiring the client to request the stream endpoint in the task info
		meta.ChannelType != channeltype.Replicate {
		return true
	}
	return false
}

func setSystemPrompt(ctx context.Context, request *relaymodel.GeneralOpenAIRequest, prompt string) (reset bool) {
	if prompt == "" {
		return false
	}
	if len(request.Messages) == 0 {
		return false
	}
	if request.Messages[0].Role == role.System {
		request.Messages[0].Content = prompt
		logger.Logger.Info("rewrite system prompt")
		return true
	}
	request.Messages = append([]relaymodel.Message{{
		Role:    role.System,
		Content: prompt,
	}}, request.Messages...)
	logger.Logger.Info("add system prompt")
	return true
}
