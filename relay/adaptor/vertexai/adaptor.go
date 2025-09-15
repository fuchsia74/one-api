package vertexai

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Laisky/errors/v2"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/relay/adaptor"
	channelhelper "github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/geminiOpenaiCompatible"
	vertexaiClaude "github.com/songquanpeng/one-api/relay/adaptor/vertexai/claude"
	"github.com/songquanpeng/one-api/relay/adaptor/vertexai/imagen"
	"github.com/songquanpeng/one-api/relay/adaptor/vertexai/veo"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	relayModel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

var _ adaptor.Adaptor = new(Adaptor)

const channelName = "vertexai"

// IsRequireGlobalEndpoint determines if the given model requires a global endpoint
//
//   - https://cloud.google.com/vertex-ai/generative-ai/docs/models/gemini/2-5-pro
//   - https://cloud.google.com/vertex-ai/generative-ai/docs/learn/locations#global-preview
func IsRequireGlobalEndpoint(model string) bool {
	// gemini-2.5-pro-preview models use global endpoint
	if strings.HasPrefix(model, "gemini-2.5-pro") ||
		strings.HasPrefix(model, "gemini-2.5-flash") {
		return true
	}

	return false
}

type Adaptor struct {
}

func (a *Adaptor) Init(meta *meta.Meta) {
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, request *model.ImageRequest) (any, error) {
	meta := meta.GetByContext(c)

	if request.ResponseFormat == nil || *request.ResponseFormat != "b64_json" {
		return nil, errors.New("only support b64_json response format")
	}

	adaptor := GetAdaptor(meta.ActualModelName)
	if adaptor == nil {
		return nil, errors.Errorf("cannot found vertex image adaptor for model %s", meta.ActualModelName)
	}

	return adaptor.ConvertImageRequest(c, request)
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	meta := meta.GetByContext(c)

	adaptor := GetAdaptor(meta.ActualModelName)
	if adaptor == nil {
		return nil, errors.Errorf("cannot found vertex chat adaptor for model %s", meta.ActualModelName)
	}

	return adaptor.ConvertRequest(c, relayMode, request)
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, request *model.ClaudeRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}

	meta := meta.GetByContext(c)

	adaptor := GetAdaptor(meta.ActualModelName)
	if adaptor == nil {
		return nil, errors.Errorf("cannot found vertex adaptor for model %s", meta.ActualModelName)
	}

	// Convert Claude Messages API request to OpenAI format first
	openaiRequest := &model.GeneralOpenAIRequest{
		Model:       request.Model,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Stream:      request.Stream != nil && *request.Stream,
		Stop:        request.StopSequences,
		Thinking:    request.Thinking,
	}

	// Add system message if present
	if request.System != "" {
		systemMessage := model.Message{
			Role:    "system",
			Content: request.System,
		}
		openaiRequest.Messages = append(openaiRequest.Messages, systemMessage)
	}

	// Convert messages
	for _, msg := range request.Messages {
		openaiMessage := model.Message{
			Role: msg.Role,
		}

		// Convert content based on type
		switch content := msg.Content.(type) {
		case string:
			// Simple string content
			openaiMessage.Content = content
		case []any:
			// Structured content blocks - convert to OpenAI format
			var contentParts []model.MessageContent
			for _, block := range content {
				if blockMap, ok := block.(map[string]any); ok {
					if blockType, exists := blockMap["type"]; exists {
						switch blockType {
						case "text":
							if text, exists := blockMap["text"]; exists {
								if textStr, ok := text.(string); ok {
									contentParts = append(contentParts, model.MessageContent{
										Type: "text",
										Text: &textStr,
									})
								}
							}
						case "image":
							if source, exists := blockMap["source"]; exists {
								if sourceMap, ok := source.(map[string]any); ok {
									if mediaType, exists := sourceMap["media_type"]; exists {
										if data, exists := sourceMap["data"]; exists {
											if mediaTypeStr, ok := mediaType.(string); ok {
												if dataStr, ok := data.(string); ok {
													imageURL := fmt.Sprintf("data:%s;base64,%s", mediaTypeStr, dataStr)
													contentParts = append(contentParts, model.MessageContent{
														Type: "image_url",
														ImageURL: &model.ImageURL{
															Url: imageURL,
														},
													})
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
			if len(contentParts) > 0 {
				openaiMessage.Content = contentParts
			}
		}

		openaiRequest.Messages = append(openaiRequest.Messages, openaiMessage)
	}

	// Convert tools if present
	if len(request.Tools) > 0 {
		var tools []model.Tool
		for _, claudeTool := range request.Tools {
			tool := model.Tool{
				Type: "function",
				Function: &model.Function{
					Name:        claudeTool.Name,
					Description: claudeTool.Description,
					Parameters:  claudeTool.InputSchema.(map[string]any),
				},
			}
			tools = append(tools, tool)
		}
		openaiRequest.Tools = tools
	}

	// Convert tool choice if present
	if request.ToolChoice != nil {
		openaiRequest.ToolChoice = request.ToolChoice
	}

	// Mark this as a Claude Messages conversion for response handling
	c.Set(ctxkey.ClaudeMessagesConversion, true)
	c.Set(ctxkey.OriginalClaudeRequest, request)

	// Now convert the OpenAI request to VertexAI format using existing logic
	return adaptor.ConvertRequest(c, relaymode.ChatCompletions, openaiRequest)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	adaptor := GetAdaptor(meta.ActualModelName)
	if adaptor == nil {
		return nil, &relayModel.ErrorWithStatusCode{
			StatusCode: http.StatusInternalServerError,
			Error: relayModel.Error{
				Message: "adaptor not found",
			},
		}
	}

	return adaptor.DoResponse(c, resp, meta)
}

func (a *Adaptor) GetModelList() []string {
	// Aggregate model lists from all subadaptors
	var models []string

	// Add models from each subadaptor
	models = append(models, adaptor.GetModelListFromPricing(vertexaiClaude.ModelRatios)...)
	models = append(models, adaptor.GetModelListFromPricing(imagen.ModelRatios)...)
	models = append(models, adaptor.GetModelListFromPricing(geminiOpenaiCompatible.ModelRatios)...)
	models = append(models, adaptor.GetModelListFromPricing(veo.ModelRatios)...)

	// Add VertexAI-specific models
	models = append(models, "text-embedding-004", "aqa")

	return models
}

func (a *Adaptor) GetChannelName() string {
	return channelName
}

// Pricing methods - VertexAI adapter aggregates pricing from subadaptors
// Following DRY principles by importing ratios from each subadaptor
func (a *Adaptor) GetDefaultModelPricing() map[string]adaptor.ModelConfig {
	// Import pricing from subadaptors to eliminate redundancy
	pricing := make(map[string]adaptor.ModelConfig)

	// Import Claude models from claude subadaptor
	for model, config := range vertexaiClaude.ModelRatios {
		pricing[model] = config
	}

	// Import Imagen models from imagen subadaptor
	for model, config := range imagen.ModelRatios {
		pricing[model] = config
	}

	// Import Gemini models from geminiOpenaiCompatible (shared with VertexAI)
	for model, config := range geminiOpenaiCompatible.ModelRatios {
		pricing[model] = config
	}

	// Import Veo models from veo subadaptor
	for model, config := range veo.ModelRatios {
		pricing[model] = config
	}

	// Add VertexAI-specific models that don't belong to subadaptors
	// Using global ratio.MilliTokensUsd = 0.5 for consistent quota-based pricing

	// VertexAI-specific models
	pricing["text-embedding-004"] = adaptor.ModelConfig{Ratio: 0.00001 * ratio.MilliTokensUsd, CompletionRatio: 1}
	pricing["aqa"] = adaptor.ModelConfig{Ratio: 1, CompletionRatio: 1}

	return pricing
}

func (a *Adaptor) GetModelRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.Ratio
	}
	// Default VertexAI pricing (similar to Gemini)
	return 0.5 * ratio.MilliTokensUsd // Default quota-based pricing
}

func (a *Adaptor) GetCompletionRatio(modelName string) float64 {
	pricing := a.GetDefaultModelPricing()
	if price, exists := pricing[modelName]; exists {
		return price.CompletionRatio
	}
	// Default completion ratio for VertexAI
	return 3.0
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// Determine the endpoint suffix based on model type and streaming
	var suffix string

	// Check if this is a non-Gemini model (like Claude) that uses rawPredict
	if strings.Contains(meta.ActualModelName, "claude") {
		suffix = "rawPredict"
	} else {
		// Gemini models use generateContent/streamGenerateContent
		if meta.IsStream {
			suffix = "streamGenerateContent?alt=sse"
		} else {
			suffix = "generateContent"
		}
	}

	location := "us-central1"
	baseHost := "us-central1-aiplatform.googleapis.com"

	// Check if model requires global endpoint
	if IsRequireGlobalEndpoint(meta.ActualModelName) {
		location = "global"
		baseHost = "aiplatform.googleapis.com"
	} else if meta.Config.Region != "" {
		location = meta.Config.Region
		baseHost = fmt.Sprintf("%s-aiplatform.googleapis.com", location)
	}

	// Handle custom base URL
	if meta.BaseURL != "" {
		baseHost = strings.TrimPrefix(meta.BaseURL, "https://")
		baseHost = strings.TrimPrefix(baseHost, "http://")
		baseHost = strings.TrimSuffix(baseHost, "/")
	}

	return fmt.Sprintf(
		"https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
		baseHost,
		meta.Config.VertexAIProjectID,
		location,
		meta.ActualModelName,
		suffix,
	), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	token, err := getToken(c, meta.ChannelId, meta.Config.VertexAIADC)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return channelhelper.DoRequestHelper(a, c, meta, requestBody)
}
