package groq

import (
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai_compatible"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct {
	adaptor.DefaultPricingMethods
}

func (a *Adaptor) GetChannelName() string {
	return "groq"
}

func (a *Adaptor) GetModelList() []string {
	return adaptor.GetModelListFromPricing(ModelRatios)
}

// GetDefaultModelPricing returns the pricing information for Groq models
// Based on Groq pricing: https://groq.com/pricing/
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelConfig {
	return ModelRatios
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Use default fallback from DefaultPricingMethods
	return a.DefaultPricingMethods.GetModelRatio(modelName)
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Use default fallback from DefaultPricingMethods
	return a.DefaultPricingMethods.GetCompletionRatio(modelName)
}

// Implement required adaptor interface methods (Groq uses OpenAI-compatible API)
func (a *Adaptor) Init(meta *meta.Meta) {}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// Handle Claude Messages requests - convert to OpenAI Chat Completions endpoint
	requestPath := meta.RequestURLPath
	if idx := strings.Index(requestPath, "?"); idx >= 0 {
		requestPath = requestPath[:idx]
	}
	if requestPath == "/v1/messages" {
		// Claude Messages requests should use OpenAI's chat completions endpoint
		chatCompletionsPath := "/v1/chat/completions"
		return openai_compatible.GetFullRequestURL(meta.BaseURL, chatCompletionsPath, meta.ChannelType), nil
	}

	// Groq uses OpenAI-compatible API endpoints
	return openai_compatible.GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	// Groq is OpenAI-compatible, so we can pass the request through with minimal changes
	// Remove reasoning_effort as Groq doesn't support it
	if request.ReasoningEffort != nil {
		request.ReasoningEffort = nil
	}
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	return nil, errors.New("groq does not support image generation")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
	// Use the shared OpenAI-compatible Claude Messages conversion
	return openai_compatible.ConvertClaudeRequest(c, request)
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	// Log request details for debugging
	logger := gmw.GetLogger(c)
	logger.Debug("sending request to groq",
		zap.String("model", meta.ActualModelName),
		zap.String("url_path", meta.RequestURLPath),
		zap.Bool("is_stream", meta.IsStream))

	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	// Add logging for debugging
	logger := gmw.GetLogger(c)
	logger.Debug("processing groq response",
		zap.String("model", meta.ActualModelName),
		zap.Bool("is_stream", meta.IsStream),
		zap.Int("status_code", resp.StatusCode),
		zap.String("content_type", resp.Header.Get("Content-Type")))

	// If this request originated from Claude Messages, convert the OpenAI-compatible
	// response to Claude Messages format and let controller forward it.
	if isClaudeConversion, exists := c.Get(ctxkey.ClaudeMessagesConversion); exists && isClaudeConversion.(bool) {
		// Convert to Claude Messages format
		claudeResp, convErr := openai_compatible.ConvertOpenAIResponseToClaudeResponse(c, resp)
		if convErr != nil {
			return nil, convErr
		}

		// Store converted response for the controller to forward verbatim
		c.Set(ctxkey.ConvertedResponse, claudeResp)

		// Do not return usage here; controller will extract it from Claude body if present
		return nil, nil
	}

	// Use the shared OpenAI-compatible response handling
	if meta.IsStream {
		err, usage = openai_compatible.StreamHandler(c, resp, meta.PromptTokens, meta.ActualModelName)
	} else {
		err, usage = openai_compatible.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	}

	// Log any errors for debugging
	if err != nil {
		logger.Error("groq response processing failed",
			zap.String("model", meta.ActualModelName),
			zap.Any("error_code", err.Error.Code),
			zap.String("error_message", err.Error.Message),
			zap.Int("error_status", err.StatusCode))
	} else if usage != nil {
		logger.Debug("groq response processed successfully",
			zap.String("model", meta.ActualModelName),
			zap.Int("prompt_tokens", usage.PromptTokens),
			zap.Int("completion_tokens", usage.CompletionTokens),
			zap.Int("total_tokens", usage.TotalTokens))
	}

	return
}
