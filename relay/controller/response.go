package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/metrics"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/pricing"
)

// RelayResponseAPIHelper handles Response API requests with direct pass-through
func RelayResponseAPIHelper(c *gin.Context) *relaymodel.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := metalib.GetByContext(c)

	// get & validate Response API request
	responseAPIRequest, err := getAndValidateResponseAPIRequest(c)
	if err != nil {
		return openai.ErrorWrapper(err, "invalid_response_api_request", http.StatusBadRequest)
	}
	meta.IsStream = responseAPIRequest.Stream != nil && *responseAPIRequest.Stream

	if reqBody, ok := c.Get(ctxkey.KeyRequestBody); ok {
		logger.Logger.Debug("get response api request", zap.ByteString("body", reqBody.([]byte)))
	}

	// Check if channel supports Response API
	if meta.ChannelType != 1 { // Only OpenAI channels support Response API for now
		return openai.ErrorWrapper(errors.New("Response API is only supported for OpenAI channels"), "unsupported_channel", http.StatusBadRequest)
	}

	// get channel model ratio
	channelModelRatio, channelCompletionRatio := getChannelRatios(c, meta.ChannelId)

	// get model ratio using three-layer pricing system
	pricingAdaptor := relay.GetAdaptor(meta.ChannelType)
	modelRatio := pricing.GetModelRatioWithThreeLayers(responseAPIRequest.Model, channelModelRatio, pricingAdaptor)
	groupRatio := c.GetFloat64(ctxkey.ChannelRatio)

	ratio := modelRatio * groupRatio

	// pre-consume quota based on estimated input tokens
	promptTokens := getResponseAPIPromptTokens(c.Request.Context(), responseAPIRequest)
	meta.PromptTokens = promptTokens
	preConsumedQuota, bizErr := preConsumeResponseAPIQuota(c, responseAPIRequest, promptTokens, ratio, meta)
	if bizErr != nil {
		logger.Logger.Warn("preConsumeResponseAPIQuota failed", zap.Any("error", *bizErr))
		return bizErr
	}

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(errors.New("invalid api type"), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	// get request body - for Response API, we pass through directly without conversion
	requestBody, err := getResponseAPIRequestBody(c, meta, responseAPIRequest, adaptor)
	if err != nil {
		return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
	}

	// for debug
	requestBodyBytes, _ := io.ReadAll(requestBody)
	requestBody = bytes.NewBuffer(requestBodyBytes)

	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		// ErrorWrapper will log the error, so we don't need to log it here
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, c.GetInt(ctxkey.TokenId))
		return RelayErrorHandler(resp)
	}

	// do response
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		// DoResponse already logs errors internally, so we don't need to log it here
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, c.GetInt(ctxkey.TokenId))
		return respErr
	}

	// post-consume quota
	quotaId := c.GetInt(ctxkey.Id)
	requestId := c.GetString(ctxkey.RequestId)

	go func() {
		// Use configurable billing timeout with model-specific adjustments
		baseBillingTimeout := time.Duration(config.BillingTimeoutSec) * time.Second
		billingTimeout := baseBillingTimeout

		ctx, cancel := context.WithTimeout(context.Background(), billingTimeout)
		defer cancel()

		// Monitor for timeout and log critical errors
		done := make(chan bool, 1)
		var quota int64

		go func() {
			quota = postConsumeResponseAPIQuota(ctx, usage, meta, responseAPIRequest, ratio, preConsumedQuota, modelRatio, groupRatio, channelCompletionRatio)

			// also update user request cost
			if quota != 0 {
				docu := model.NewUserRequestCost(
					quotaId,
					requestId,
					quota,
				)
				if err = docu.Insert(); err != nil {
					logger.Logger.Error("insert user request cost failed", zap.Error(err))
				}
			}
			done <- true
		}()

		select {
		case <-done:
			// Billing completed successfully
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				estimatedQuota := float64(usage.PromptTokens+usage.CompletionTokens) * ratio
				elapsedTime := time.Since(meta.StartTime)

				logger.Logger.Error("CRITICAL BILLING TIMEOUT",
					zap.String("model", responseAPIRequest.Model),
					zap.String("requestId", requestId),
					zap.Int("userId", meta.UserId),
					zap.Int64("estimatedQuota", int64(estimatedQuota)),
					zap.Duration("elapsedTime", elapsedTime))

				// Record billing timeout in metrics
				metrics.GlobalRecorder.RecordBillingTimeout(meta.UserId, meta.ChannelId, responseAPIRequest.Model, estimatedQuota, elapsedTime)

				// TODO: Implement dead letter queue or retry mechanism for failed billing
			}
		}
	}()

	return nil
}

// getChannelRatios gets channel model and completion ratios from unified ModelConfigs
func getChannelRatios(c *gin.Context, channelId int) (map[string]float64, map[string]float64) {
	channel := c.MustGet(ctxkey.ChannelModel).(*model.Channel)

	// Only use unified ModelConfigs after migration
	modelRatios := channel.GetModelRatioFromConfigs()
	completionRatios := channel.GetCompletionRatioFromConfigs()

	return modelRatios, completionRatios
}

// getAndValidateResponseAPIRequest gets and validates Response API request
func getAndValidateResponseAPIRequest(c *gin.Context) (*openai.ResponseAPIRequest, error) {
	responseAPIRequest := &openai.ResponseAPIRequest{}
	err := common.UnmarshalBodyReusable(c, responseAPIRequest)
	if err != nil {
		return nil, err
	}

	// Basic validation
	if responseAPIRequest.Model == "" {
		return nil, errors.New("model is required")
	}

	// Either input or prompt is required, but not both
	hasInput := len(responseAPIRequest.Input) > 0
	hasPrompt := responseAPIRequest.Prompt != nil

	if !hasInput && !hasPrompt {
		return nil, errors.New("either input or prompt is required")
	}
	if hasInput && hasPrompt {
		return nil, errors.New("input and prompt are mutually exclusive - provide only one")
	}

	return responseAPIRequest, nil
}

// getResponseAPIPromptTokens estimates prompt tokens for Response API requests
func getResponseAPIPromptTokens(ctx context.Context, responseAPIRequest *openai.ResponseAPIRequest) int {
	// For now, use a simple estimation based on input content
	// This will be improved with proper token counting
	totalTokens := 0

	// Count tokens from input array (if present)
	for _, input := range responseAPIRequest.Input {
		switch v := input.(type) {
		case map[string]interface{}:
			if content, ok := v["content"].(string); ok {
				// Simple estimation: ~4 characters per token
				totalTokens += len(content) / 4
			}
		case string:
			totalTokens += len(v) / 4
		}
	}

	// Count tokens from prompt template (if present)
	if responseAPIRequest.Prompt != nil {
		// Estimate tokens for prompt template ID (small fixed cost)
		totalTokens += 10

		// Count tokens from prompt variables
		for _, value := range responseAPIRequest.Prompt.Variables {
			switch v := value.(type) {
			case string:
				totalTokens += len(v) / 4
			case map[string]interface{}:
				// For complex variables like input_file, add a fixed cost
				totalTokens += 20
			}
		}
	}

	// Add instruction tokens if present
	if responseAPIRequest.Instructions != nil {
		totalTokens += len(*responseAPIRequest.Instructions) / 4
	}

	// Minimum token count
	if totalTokens < 10 {
		totalTokens = 10
	}

	return totalTokens
}

// preConsumeResponseAPIQuota pre-consumes quota for Response API requests
func preConsumeResponseAPIQuota(c *gin.Context, responseAPIRequest *openai.ResponseAPIRequest, promptTokens int, ratio float64, meta *metalib.Meta) (int64, *relaymodel.ErrorWithStatusCode) {
	// Use similar logic to ChatCompletion pre-consumption
	preConsumedTokens := int64(promptTokens)
	if responseAPIRequest.MaxOutputTokens != nil {
		preConsumedTokens += int64(*responseAPIRequest.MaxOutputTokens)
	}

	baseQuota := int64(float64(preConsumedTokens) * ratio)
	if ratio != 0 && baseQuota <= 0 {
		baseQuota = 1
	}

	tokenQuota := c.GetInt64(ctxkey.TokenQuota)
	tokenQuotaUnlimited := c.GetBool(ctxkey.TokenQuotaUnlimited)
	userQuota, err := model.CacheGetUserQuota(c.Request.Context(), meta.UserId)
	if err != nil {
		return baseQuota, openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota-baseQuota < 0 {
		return baseQuota, openai.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}

	if !tokenQuotaUnlimited && tokenQuota > 0 && tokenQuota-baseQuota < 0 {
		return baseQuota, openai.ErrorWrapper(errors.New("token quota is not enough"), "insufficient_token_quota", http.StatusForbidden)
	}

	err = model.PreConsumeTokenQuota(c.GetInt(ctxkey.TokenId), baseQuota)
	if err != nil {
		return baseQuota, openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
	}

	return baseQuota, nil
}

// postConsumeResponseAPIQuota calculates final quota consumption for Response API requests
// Following DRY principle by reusing the centralized billing.PostConsumeQuota function
func postConsumeResponseAPIQuota(ctx context.Context,
	usage *relaymodel.Usage,
	meta *metalib.Meta,
	responseAPIRequest *openai.ResponseAPIRequest,
	ratio float64,
	preConsumedQuota int64,
	modelRatio float64,
	groupRatio float64,
	channelCompletionRatio map[string]float64) (quota int64) {

	if usage == nil {
		logger.Logger.Error("usage is nil, which is unexpected")
		return
	}

	// Use three-layer pricing system for completion ratio
	pricingAdaptor := relay.GetAdaptor(meta.ChannelType)
	completionRatio := pricing.GetCompletionRatioWithThreeLayers(responseAPIRequest.Model, channelCompletionRatio, pricingAdaptor)

	// Calculate quota using the same formula as ChatCompletion
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens
	quota = int64((float64(promptTokens)+float64(completionTokens)*completionRatio)*ratio) + usage.ToolsCost
	if ratio != 0 && quota <= 0 {
		quota = 1
	}

	// Use centralized detailed billing function to follow DRY principle
	quotaDelta := quota - preConsumedQuota
	billing.PostConsumeQuotaDetailed(ctx, meta.TokenId, quotaDelta, quota, meta.UserId, meta.ChannelId,
		promptTokens, completionTokens, modelRatio, groupRatio, responseAPIRequest.Model, meta.TokenName,
		meta.IsStream, meta.StartTime, false, // Response API doesn't have system prompt reset concept
		completionRatio, usage.ToolsCost)

	return quota
}

// getResponseAPIRequestBody gets the request body for Response API requests
func getResponseAPIRequestBody(c *gin.Context, meta *metalib.Meta, responseAPIRequest *openai.ResponseAPIRequest, adaptor adaptor.Adaptor) (io.Reader, error) {
	// For Response API, we pass through the request directly without conversion
	// The request is already in the correct format
	jsonData, err := json.Marshal(responseAPIRequest)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(jsonData), nil
}
