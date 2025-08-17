package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/message"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/monitor"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"

	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/controller"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/pricing"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func buildTestRequest(model string) *relaymodel.GeneralOpenAIRequest {
	if model == "" {
		model = "gpt-4o-mini"
	}
	testRequest := &relaymodel.GeneralOpenAIRequest{
		MaxTokens: config.TestMaxTokens,
		Model:     model,
	}
	testMessage := relaymodel.Message{
		Role:    "user",
		Content: config.TestPrompt,
	}
	testRequest.Messages = append(testRequest.Messages, testMessage)
	return testRequest
}

func parseTestResponse(resp string) (*openai.TextResponse, string, error) {
	var response openai.TextResponse
	err := json.Unmarshal([]byte(resp), &response)
	if err != nil {
		return nil, "", err
	}
	if len(response.Choices) == 0 {
		return nil, "", errors.New("response has no choices")
	}
	stringContent, ok := response.Choices[0].Content.(string)
	if !ok {
		return nil, "", errors.New("response content is not string")
	}
	return &response, stringContent, nil
}

// calculateTestCost calculates the actual cost that would have been charged for a test request
// This is used for informational purposes to track the real cost of testing operations
func calculateTestCost(usage *relaymodel.Usage, meta *meta.Meta, request *relaymodel.GeneralOpenAIRequest) int64 {
	if usage == nil {
		return 0
	}

	// Get model ratio and completion ratio using three-layer pricing system
	pricingAdaptor := relay.GetAdaptor(meta.ChannelType)
	modelRatio := pricing.GetModelRatioWithThreeLayers(request.Model, nil, pricingAdaptor)
	completionRatio := pricing.GetCompletionRatioWithThreeLayers(request.Model, nil, pricingAdaptor)

	// Use the same group ratio as set in the context (typically 1.0 for tests)
	groupRatio := 1.0 // Default group ratio for tests

	// Calculate cost using the same formula as postConsumeQuota
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens
	ratio := modelRatio * groupRatio

	quota := int64(math.Ceil((float64(promptTokens)+float64(completionTokens)*completionRatio)*ratio)) + usage.ToolsCost
	if ratio != 0 && quota <= 0 {
		quota = 1
	}

	return quota
}

func testChannel(ctx context.Context, channel *model.Channel, request *relaymodel.GeneralOpenAIRequest) (responseMessage string, err error, openaiErr *relaymodel.Error) {
	startTime := time.Now()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = &http.Request{
		Method: http.MethodPost,
		URL:    &url.URL{Path: "/v1/chat/completions"},
		Body:   nil,
		Header: make(http.Header),
	}
	c.Request.Header.Set("Authorization", "Bearer "+channel.Key)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(ctxkey.Channel, channel.Type)
	c.Set(ctxkey.BaseURL, channel.GetBaseURL())
	cfg, _ := channel.LoadConfig()
	c.Set(ctxkey.Config, cfg)
	middleware.SetupContextForSelectedChannel(c, channel, "")
	meta := meta.GetByContext(c)
	apiType := channeltype.ToAPIType(channel.Type)
	adaptor := relay.GetAdaptor(apiType)
	if adaptor == nil {
		return "", fmt.Errorf("invalid api type: %d, adaptor is nil", apiType), nil
	}
	adaptor.Init(meta)
	modelName := request.Model
	modelMap := channel.GetModelMapping()
	if modelName == "" || !strings.Contains(channel.Models, modelName) {
		modelNames := strings.Split(channel.Models, ",")
		if len(modelNames) > 0 {
			modelName = modelNames[0]
		}
	}
	if modelMap != nil && modelMap[modelName] != "" {
		modelName = modelMap[modelName]
	}
	// Check for AWS inference profile ARN mapping
	if channel.Type == channeltype.AwsClaude {
		arnMap := channel.GetInferenceProfileArnMap()
		if arnMap != nil {
			if arn, exists := arnMap[modelName]; exists && arn != "" {
				meta.ActualModelName = arn
			}
		}
	}
	meta.OriginModelName = request.Model
	request.Model = modelName
	convertedRequest, err := adaptor.ConvertRequest(c, relaymode.ChatCompletions, request)
	if err != nil {
		return "", err, nil
	}
	c.Set(ctxkey.ConvertedRequest, convertedRequest)

	jsonData, err := json.Marshal(convertedRequest)
	if err != nil {
		return "", err, nil
	}

	// Capture usage information for accurate test logging
	var actualUsage *relaymodel.Usage
	defer func() {
		logContent := fmt.Sprintf("test channel %s succeed，response: %s", channel.Name, responseMessage)
		if err != nil || openaiErr != nil {
			errorMessage := ""
			if err != nil {
				errorMessage = err.Error()
			} else {
				errorMessage = openaiErr.Message
			}
			logContent = fmt.Sprintf("test channel %s failed, error: %s", channel.Name, errorMessage)
		}

		// Create test log with actual usage information if available
		testLog := &model.Log{
			ChannelId:   channel.Id,
			ModelName:   modelName,
			Content:     logContent,
			ElapsedTime: helper.CalcElapsedTime(startTime),
		}

		// Include actual token usage and calculated cost in test logs for accurate cost tracking
		if actualUsage != nil {
			testLog.PromptTokens = actualUsage.PromptTokens
			testLog.CompletionTokens = actualUsage.CompletionTokens

			// Calculate the actual cost that would have been charged (for informational purposes)
			// This helps with cost tracking and budgeting while keeping tests free for users
			actualCost := calculateTestCost(actualUsage, meta, request)
			testLog.Quota = int(actualCost)
		}

		go model.RecordTestLog(ctx, testLog)
	}()
	logger.Logger.Info(string(jsonData))
	requestBody := bytes.NewBuffer(jsonData)
	c.Request.Body = io.NopCloser(requestBody)
	var resp *http.Response
	resp, err = adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		return "", err, nil
	}
	if resp != nil && resp.StatusCode != http.StatusOK {
		wrappedErr := controller.RelayErrorHandler(resp)
		errorMessage := wrappedErr.Error.Message
		if errorMessage != "" {
			errorMessage = ", error message: " + errorMessage
		}
		err = fmt.Errorf("http status code: %d%s", resp.StatusCode, errorMessage)
		return "", err, &wrappedErr.Error
	}
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		err = fmt.Errorf("%s", respErr.Error.Message)
		return "", err, &respErr.Error
	}
	if usage == nil {
		err = errors.New("usage is nil")
		return "", err, nil
	}

	// Capture usage for test logging
	actualUsage = usage
	rawResponse := w.Body.String()
	_, responseMessage, err = parseTestResponse(rawResponse)
	if err != nil {
		logger.Logger.Error("failed to parse error", zap.Error(err), zap.String("response", rawResponse))
		return "", err, nil
	}
	result := w.Result()
	// print result.Body
	var respBody []byte
	respBody, err = io.ReadAll(result.Body)
	if err != nil {
		return "", err, nil
	}
	logger.Logger.Info(fmt.Sprintf("testing channel #%d, response: \n%s", channel.Id, string(respBody)))
	return responseMessage, nil, nil
}

func TestChannel(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, true)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	modelName := c.Query("model")
	testRequest := buildTestRequest(modelName)
	tik := time.Now()
	responseMessage, err, _ := testChannel(ctx, channel, testRequest)
	tok := time.Now()
	milliseconds := tok.Sub(tik).Milliseconds()
	if err != nil {
		milliseconds = 0
	}
	go channel.UpdateResponseTime(milliseconds)
	consumedTime := float64(milliseconds) / 1000.0
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"message":   err.Error(),
			"time":      consumedTime,
			"modelName": modelName,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   responseMessage,
		"time":      consumedTime,
		"modelName": modelName,
	})
	return
}

var testAllChannelsLock sync.Mutex
var testAllChannelsRunning bool = false

func testChannels(ctx context.Context, notify bool, scope string) error {
	if config.RootUserEmail == "" {
		config.RootUserEmail = model.GetRootUserEmail()
	}
	testAllChannelsLock.Lock()
	if testAllChannelsRunning {
		testAllChannelsLock.Unlock()
		return errors.New("Test is already running")
	}
	testAllChannelsRunning = true
	testAllChannelsLock.Unlock()
	channels, err := model.GetAllChannels(0, 0, scope, "", "")
	if err != nil {
		return err
	}
	var disableThreshold = int64(config.ChannelDisableThreshold * 1000)
	if disableThreshold == 0 {
		disableThreshold = 10000000 // a impossible value
	}
	go func() {
		for _, channel := range channels {
			isChannelEnabled := channel.Status == model.ChannelStatusEnabled
			tik := time.Now()
			testRequest := buildTestRequest("")
			_, err, openaiErr := testChannel(ctx, channel, testRequest)
			tok := time.Now()
			milliseconds := tok.Sub(tik).Milliseconds()
			if isChannelEnabled && milliseconds > disableThreshold {
				err = fmt.Errorf("Response time %.2fs exceeds threshold %.2fs", float64(milliseconds)/1000.0, float64(disableThreshold)/1000.0)
				if config.AutomaticDisableChannelEnabled {
					monitor.DisableChannel(channel.Id, channel.Name, err.Error())
				} else {
					_ = message.Notify(message.ByAll, fmt.Sprintf("Channel %s （%d）Test超时", channel.Name, channel.Id), "", err.Error())
				}
			}
			if isChannelEnabled && (err != nil || monitor.ShouldDisableChannel(openaiErr, -1)) {
				monitor.DisableChannel(channel.Id, channel.Name, err.Error())
			}
			if !isChannelEnabled && (err == nil && monitor.ShouldEnableChannel(err, openaiErr)) {
				monitor.EnableChannel(channel.Id, channel.Name)
			}
			channel.UpdateResponseTime(milliseconds)
			time.Sleep(config.RequestInterval)
		}
		testAllChannelsLock.Lock()
		testAllChannelsRunning = false
		testAllChannelsLock.Unlock()
		if notify {
			err := message.Notify(message.ByAll, "Channel test completed", "", "Channel test completed, if you have not received the disable notification, it means that all channels are normal")
			if err != nil {
				logger.Logger.Error("failed to send notify", zap.Error(err))
			}
		}
	}()
	return nil
}

func TestChannels(c *gin.Context) {
	ctx := gmw.Ctx(c)
	scope := c.Query("scope")
	if scope == "" {
		scope = "all"
	}
	err := testChannels(ctx, true, scope)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
	return
}

func AutomaticallyTestChannels(frequency int) {
	ctx := context.Background()
	for {
		time.Sleep(time.Duration(frequency) * time.Minute)
		logger.Logger.Info("testing all channels")
		_ = testChannels(ctx, false, "all")
		logger.Logger.Info("channel test finished")
	}
}
