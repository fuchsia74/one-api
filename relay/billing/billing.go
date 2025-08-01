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
		logger.Logger.Error(fmt.Sprintf("PostConsumeQuota: invalid tokenId %d", tokenId))
		return
	}
	if userId <= 0 {
		logger.Logger.Error(fmt.Sprintf("PostConsumeQuota: invalid userId %d", userId))
		return
	}
	if channelId <= 0 {
		logger.Logger.Error(fmt.Sprintf("PostConsumeQuota: invalid channelId %d", channelId))
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
		logger.Logger.Error(fmt.Sprintf("totalQuota consumed is %d, something is wrong", totalQuota))
	}
}

// PostConsumeQuotaDetailed handles detailed billing for ChatCompletion and Response API requests
// This function properly logs individual prompt and completion tokens with additional metadata
// SAFETY: This function validates all inputs to prevent billing errors
func PostConsumeQuotaDetailed(ctx context.Context, tokenId int, quotaDelta int64, totalQuota int64,
	userId int, channelId int, promptTokens int, completionTokens int,
	modelRatio float64, groupRatio float64, modelName string, tokenName string,
	isStream bool, startTime time.Time, systemPromptReset bool,
	completionRatio float64, toolsCost int64) {

	// Record billing operation start time for monitoring
	billingStartTime := time.Now()
	billingSuccess := true

	// Input validation for safety
	if ctx == nil {
		logger.Logger.Error("PostConsumeQuotaDetailed: context is nil")
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", userId, channelId, modelName)
		return
	}
	if tokenId <= 0 {
		logger.Logger.Error(fmt.Sprintf("PostConsumeQuotaDetailed: invalid tokenId %d", tokenId))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", userId, channelId, modelName)
		return
	}
	if userId <= 0 {
		logger.Logger.Error(fmt.Sprintf("PostConsumeQuotaDetailed: invalid userId %d", userId))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", userId, channelId, modelName)
		return
	}
	if channelId <= 0 {
		logger.Logger.Error(fmt.Sprintf("PostConsumeQuotaDetailed: invalid channelId %d", channelId))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", userId, channelId, modelName)
		return
	}
	if promptTokens < 0 || completionTokens < 0 {
		logger.Logger.Error(fmt.Sprintf("PostConsumeQuotaDetailed: negative token counts - prompt: %d, completion: %d", promptTokens, completionTokens))
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", userId, channelId, modelName)
		return
	}
	if modelName == "" {
		logger.Logger.Error("PostConsumeQuotaDetailed: modelName is empty")
		metrics.GlobalRecorder.RecordBillingError("validation_error", "post_consume_detailed", userId, channelId, modelName)
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
		metrics.GlobalRecorder.RecordBillingError("database_error", "post_consume_token_quota", userId, channelId, modelName)
		billingSuccess = false
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
		metrics.GlobalRecorder.RecordBillingError("cache_error", "update_user_quota_cache", userId, channelId, modelName)
		billingSuccess = false
	}

	// totalQuota is total quota consumed
	// Always log the request for tracking purposes, regardless of quota amount
	var logContent string
	if toolsCost == 0 {
		logContent = fmt.Sprintf("model rate %.2f, group rate %.2f, completion rate %.2f", modelRatio, groupRatio, completionRatio)
	} else {
		logContent = fmt.Sprintf("model rate %.2f, group rate %.2f, completion rate %.2f, tools cost %d", modelRatio, groupRatio, completionRatio, toolsCost)
	}
	model.RecordConsumeLog(ctx, &model.Log{
		UserId:            userId,
		ChannelId:         channelId,
		PromptTokens:      promptTokens,
		CompletionTokens:  completionTokens,
		ModelName:         modelName,
		TokenName:         tokenName,
		Quota:             int(totalQuota),
		Content:           logContent,
		IsStream:          isStream,
		ElapsedTime:       helper.CalcElapsedTime(startTime),
		SystemPromptReset: systemPromptReset,
	})

	// Only update quotas when totalQuota > 0
	if totalQuota > 0 {
		model.UpdateUserUsedQuotaAndRequestCount(userId, totalQuota)
		model.UpdateChannelUsedQuota(channelId, totalQuota)
	}
	if totalQuota <= 0 {
		logger.Logger.Error("invalid totalQuota consumed - something is wrong",
			zap.Int64("total_quota", totalQuota),
			zap.Int("user_id", userId),
			zap.Int("channel_id", channelId),
			zap.String("model_name", modelName))
		metrics.GlobalRecorder.RecordBillingError("calculation_error", "post_consume_detailed", userId, channelId, modelName)
		billingSuccess = false
	}

	// Record billing operation completion
	metrics.GlobalRecorder.RecordBillingOperation(billingStartTime, "post_consume_detailed", billingSuccess, userId, channelId, modelName, float64(totalQuota))
}
