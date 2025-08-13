package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/metrics"
	"github.com/songquanpeng/one-api/model"
)

func ReturnPreConsumedQuota(ctx context.Context, preConsumedQuota int64, tokenId int) {
	if preConsumedQuota != 0 {
		go func(ctx context.Context) {
			// return pre-consumed quota
			err := model.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
			if err != nil {
				logger.Logger.Warn("failed to return pre-consumed quota - cleanup operation failed",
					zap.Error(err),
					zap.Int("tokenId", tokenId),
					zap.Int64("preConsumedQuota", preConsumedQuota),
					zap.String("note", "main billing already completed successfully"))
			}
		}(ctx)
	}
}

// PostConsumeQuota handles simple billing for Audio API (legacy compatibility)
// SAFETY: This function is preserved for backward compatibility with Audio API
// WARNING: This function logs totalQuota as promptTokens and sets completionTokens to 0
func PostConsumeQuota(ctx context.Context, tokenId int, quotaDelta int64, totalQuota int64, userId int, channelId int, modelRatio float64, groupRatio float64, modelName string, tokenName string) {
	// Input validation for safety
	if ctx == nil {
		logger.Logger.Error("PostConsumeQuota: context is nil")
		return
	}
	if tokenId <= 0 {
		logger.Logger.Error("PostConsumeQuota: invalid tokenId", zap.Int("token_id", tokenId))
		return
	}
	if userId <= 0 {
		logger.Logger.Error("PostConsumeQuota: invalid userId", zap.Int("user_id", userId))
		return
	}
	if channelId <= 0 {
		logger.Logger.Error("PostConsumeQuota: invalid channelId", zap.Int("channel_id", channelId))
		return
	}
	if modelName == "" {
		logger.Logger.Error("PostConsumeQuota: modelName is empty")
		return
	}

	// quotaDelta is remaining quota to be consumed
	err := model.PostConsumeTokenQuota(tokenId, quotaDelta)
	if err != nil {
		logger.Logger.Error("CRITICAL: upstream request was sent but billing failed - unbilled request detected",
			zap.Error(err),
			zap.Int("tokenId", tokenId),
			zap.Int("userId", userId),
			zap.Int("channelId", channelId),
			zap.String("model", modelName),
			zap.Int64("quotaDelta", quotaDelta),
			zap.Int64("totalQuota", totalQuota))
	}
	err = model.CacheUpdateUserQuota(ctx, userId)
	if err != nil {
		logger.Logger.Warn("user quota cache update failed - billing completed successfully",
			zap.Error(err),
			zap.Int("userId", userId),
			zap.Int("channelId", channelId),
			zap.String("model", modelName),
			zap.Int64("totalQuota", totalQuota),
			zap.String("note", "database billing succeeded, cache will be refreshed on next request"))
	}
	// totalQuota is total quota consumed
	// Always log the request for tracking purposes, regardless of quota amount
	logContent := fmt.Sprintf("model rate %.2f, group rate %.2f", modelRatio, groupRatio)
	model.RecordConsumeLog(ctx, &model.Log{
		UserId:           userId,
		ChannelId:        channelId,
		PromptTokens:     int(totalQuota), // NOTE: For Audio API, total quota is logged as prompt tokens
		CompletionTokens: 0,               // NOTE: Audio API doesn't have separate completion tokens
		ModelName:        modelName,
		TokenName:        tokenName,
		Quota:            int(totalQuota),
		Content:          logContent,
	})

	// Only update quotas when totalQuota > 0
	if totalQuota > 0 {
		model.UpdateUserUsedQuotaAndRequestCount(userId, totalQuota)
		model.UpdateChannelUsedQuota(channelId, totalQuota)
	}
	if totalQuota <= 0 {
		logger.Logger.Error("totalQuota consumed is invalid", zap.Int64("total_quota", totalQuota))
	}
}

// QuotaConsumeDetail encapsulates all parameters for detailed quota consumption billing
type QuotaConsumeDetail struct {
	Ctx                    context.Context
	TokenId                int
	QuotaDelta             int64
	TotalQuota             int64
	UserId                 int
	ChannelId              int
	PromptTokens           int
	CompletionTokens       int
	ModelRatio             float64
	GroupRatio             float64
	ModelName              string
	TokenName              string
	IsStream               bool
	StartTime              time.Time
	SystemPromptReset      bool
	CompletionRatio        float64
	ToolsCost              int64
	CachedPromptTokens     int
	CachedCompletionTokens int
}

// PostConsumeQuotaDetailed handles detailed billing for ChatCompletion and Response API requests
// This function properly logs individual prompt and completion tokens with additional metadata
// SAFETY: This function validates all inputs to prevent billing errors
func PostConsumeQuotaDetailed(detail QuotaConsumeDetail) {

	// Record billing operation start time for monitoring
	billingStartTime := time.Now()
	billingSuccess := true

	// Input validation for safety
	if detail.Ctx == nil {
		logger.Logger.Error("PostConsumeQuotaDetailed: context is nil")
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", detail.UserId, detail.ChannelId, detail.ModelName)
		return
	}
	if detail.TokenId <= 0 {
		logger.Logger.Error("PostConsumeQuotaDetailed: invalid tokenId", zap.Int("token_id", detail.TokenId))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", detail.UserId, detail.ChannelId, detail.ModelName)
		return
	}
	if detail.UserId <= 0 {
		logger.Logger.Error("PostConsumeQuotaDetailed: invalid userId", zap.Int("user_id", detail.UserId))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", detail.UserId, detail.ChannelId, detail.ModelName)
		return
	}
	if detail.ChannelId <= 0 {
		logger.Logger.Error("PostConsumeQuotaDetailed: invalid channelId", zap.Int("channel_id", detail.ChannelId))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", detail.UserId, detail.ChannelId, detail.ModelName)
		return
	}
	if detail.PromptTokens < 0 || detail.CompletionTokens < 0 {
		logger.Logger.Error("PostConsumeQuotaDetailed: negative token counts",
			zap.Int("prompt_tokens", detail.PromptTokens),
			zap.Int("completion_tokens", detail.CompletionTokens))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", detail.UserId, detail.ChannelId, detail.ModelName)
		return
	}
	if detail.ModelName == "" {
		logger.Logger.Error("PostConsumeQuotaDetailed: modelName is empty")
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", detail.UserId, detail.ChannelId, detail.ModelName)
		return
	}

	// quotaDelta is remaining quota to be consumed
	err := model.PostConsumeTokenQuota(detail.TokenId, detail.QuotaDelta)
	if err != nil {
		logger.Logger.Error("CRITICAL: upstream request was sent but billing failed - unbilled request detected",
			zap.Error(err),
			zap.Int("tokenId", detail.TokenId),
			zap.Int("userId", detail.UserId),
			zap.Int("channelId", detail.ChannelId),
			zap.String("model", detail.ModelName),
			zap.Int64("quotaDelta", detail.QuotaDelta),
			zap.Int64("totalQuota", detail.TotalQuota))
		metrics.GlobalRecorder.RecordBillingError("database_error", "post_consume_token_quota", detail.UserId, detail.ChannelId, detail.ModelName)
		billingSuccess = false
	}
	err = model.CacheUpdateUserQuota(detail.Ctx, detail.UserId)
	if err != nil {
		logger.Logger.Warn("user quota cache update failed - billing completed successfully",
			zap.Error(err),
			zap.Int("userId", detail.UserId),
			zap.Int("channelId", detail.ChannelId),
			zap.String("model", detail.ModelName),
			zap.Int64("totalQuota", detail.TotalQuota),
			zap.String("note", "database billing succeeded, cache will be refreshed on next request"))
		metrics.GlobalRecorder.RecordBillingError("cache_error", "update_user_quota_cache", detail.UserId, detail.ChannelId, detail.ModelName)
		billingSuccess = false
	}

	// totalQuota is total quota consumed
	// Always log the request for tracking purposes, regardless of quota amount
	var logContent string
	if detail.ToolsCost == 0 {
		logContent = fmt.Sprintf("model rate %.2f, group rate %.2f, completion rate %.2f, cached_prompt %d, cached_completion %d",
			detail.ModelRatio, detail.GroupRatio, detail.CompletionRatio, detail.CachedPromptTokens, detail.CachedCompletionTokens)
	} else {
		logContent = fmt.Sprintf("model rate %.2f, group rate %.2f, completion rate %.2f, tools cost %d, cached_prompt %d, cached_completion %d",
			detail.ModelRatio, detail.GroupRatio, detail.CompletionRatio, detail.ToolsCost, detail.CachedPromptTokens, detail.CachedCompletionTokens)
	}
	model.RecordConsumeLog(detail.Ctx, &model.Log{
		UserId:                 detail.UserId,
		ChannelId:              detail.ChannelId,
		PromptTokens:           detail.PromptTokens,
		CompletionTokens:       detail.CompletionTokens,
		ModelName:              detail.ModelName,
		TokenName:              detail.TokenName,
		Quota:                  int(detail.TotalQuota),
		Content:                logContent,
		IsStream:               detail.IsStream,
		ElapsedTime:            helper.CalcElapsedTime(detail.StartTime),
		SystemPromptReset:      detail.SystemPromptReset,
		CachedPromptTokens:     detail.CachedPromptTokens,
		CachedCompletionTokens: detail.CachedCompletionTokens,
	})

	// Only update quotas when totalQuota > 0
	if detail.TotalQuota > 0 {
		model.UpdateUserUsedQuotaAndRequestCount(detail.UserId, detail.TotalQuota)
		model.UpdateChannelUsedQuota(detail.ChannelId, detail.TotalQuota)
	}
	if detail.TotalQuota <= 0 {
		logger.Logger.Error("invalid totalQuota consumed - something is wrong",
			zap.Int64("total_quota", detail.TotalQuota),
			zap.Int("user_id", detail.UserId),
			zap.Int("channel_id", detail.ChannelId),
			zap.String("model_name", detail.ModelName))
		metrics.GlobalRecorder.RecordBillingError("calculation_error", "post_consume_detailed", detail.UserId, detail.ChannelId, detail.ModelName)
		billingSuccess = false
	}

	// Record billing operation completion
	metrics.GlobalRecorder.RecordBillingOperation(billingStartTime, "post_consume_detailed", billingSuccess, detail.UserId, detail.ChannelId, detail.ModelName, float64(detail.TotalQuota))
}

// PostConsumeQuotaDetailedWithTraceID handles detailed billing with explicit trace ID
func PostConsumeQuotaDetailedWithTraceID(ctx context.Context, traceId string, tokenId int, quotaDelta int64, totalQuota int64,
	userId int, channelId int, promptTokens int, completionTokens int,
	modelRatio float64, groupRatio float64, modelName string, tokenName string,
	isStream bool, startTime time.Time, systemPromptReset bool,
	completionRatio float64, toolsCost int64,
	cachedPromptTokens int, cachedCompletionTokens int) {

	// Record billing operation start time for monitoring
	billingStartTime := time.Now()
	billingSuccess := true

	// Input validation for safety
	if ctx == nil {
		logger.Logger.Error("PostConsumeQuotaDetailedWithTraceID: context is nil")
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed_with_trace", userId, channelId, modelName)
		return
	}
	if tokenId <= 0 {
		logger.Logger.Error("PostConsumeQuotaDetailedWithTraceID: invalid tokenId", zap.Int("token_id", tokenId))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed_with_trace", userId, channelId, modelName)
		return
	}
	if userId <= 0 {
		logger.Logger.Error("PostConsumeQuotaDetailedWithTraceID: invalid userId", zap.Int("user_id", userId))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed_with_trace", userId, channelId, modelName)
		return
	}
	if channelId <= 0 {
		logger.Logger.Error("PostConsumeQuotaDetailedWithTraceID: invalid channelId", zap.Int("channel_id", channelId))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed_with_trace", userId, channelId, modelName)
		return
	}
	if modelName == "" {
		logger.Logger.Error("PostConsumeQuotaDetailedWithTraceID: modelName is empty")
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed_with_trace", userId, channelId, modelName)
		return
	}

	// quotaDelta is remaining quota to be consumed
	err := model.PostConsumeTokenQuota(tokenId, quotaDelta)
	if err != nil {
		logger.Logger.Error("CRITICAL: upstream request was sent but billing failed - unbilled request detected",
			zap.Error(err),
			zap.Int("tokenId", tokenId),
			zap.Int("userId", userId),
			zap.Int("channelId", channelId),
			zap.String("model", modelName),
			zap.Int64("quotaDelta", quotaDelta),
			zap.Int64("totalQuota", totalQuota))
		metrics.GlobalRecorder.RecordBillingError("database_error", "post_consume_token_quota_with_trace", userId, channelId, modelName)
		billingSuccess = false
	}

	// Prepare log content with detailed breakdown
	var logContent string
	if toolsCost > 0 {
		logContent = fmt.Sprintf("model rate %.2f, group rate %.2f, completion rate %.2f, tools cost %d, cached_prompt %d, cached_completion %d",
			modelRatio, groupRatio, completionRatio, toolsCost, cachedPromptTokens, cachedCompletionTokens)
	} else {
		logContent = fmt.Sprintf("model rate %.2f, group rate %.2f, completion rate %.2f, cached_prompt %d, cached_completion %d",
			modelRatio, groupRatio, completionRatio, cachedPromptTokens, cachedCompletionTokens)
	}

	// Always log the request for tracking purposes, regardless of quota amount
	model.RecordConsumeLogWithTraceID(ctx, traceId, &model.Log{
		UserId:                 userId,
		ChannelId:              channelId,
		PromptTokens:           promptTokens,
		CompletionTokens:       completionTokens,
		ModelName:              modelName,
		TokenName:              tokenName,
		Quota:                  int(totalQuota),
		Content:                logContent,
		IsStream:               isStream,
		ElapsedTime:            helper.CalcElapsedTime(startTime),
		SystemPromptReset:      systemPromptReset,
		CachedPromptTokens:     cachedPromptTokens,
		CachedCompletionTokens: cachedCompletionTokens,
	})

	// Only update quotas when totalQuota > 0
	if totalQuota > 0 {
		model.UpdateUserUsedQuotaAndRequestCount(userId, totalQuota)
		model.UpdateChannelUsedQuota(channelId, totalQuota)
	} else {
		logger.Logger.Error("invalid totalQuota consumed - something is wrong",
			zap.Int64("total_quota", totalQuota),
			zap.Int("user_id", userId),
			zap.Int("channel_id", channelId),
			zap.String("model_name", modelName))
		metrics.GlobalRecorder.RecordBillingError("calculation_error", "post_consume_detailed_with_trace", userId, channelId, modelName)
		billingSuccess = false
	}

	// Record billing operation completion
	metrics.GlobalRecorder.RecordBillingOperation(billingStartTime, "post_consume_detailed_with_trace", billingSuccess, userId, channelId, modelName, float64(totalQuota))
}
