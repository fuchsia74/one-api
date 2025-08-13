package controller

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/model"
	relay "github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/apitype"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// https://platform.openai.com/docs/api-reference/models/list

type OpenAIModelPermission struct {
	Id                 string  `json:"id"`
	Object             string  `json:"object"`
	Created            int     `json:"created"`
	AllowCreateEngine  bool    `json:"allow_create_engine"`
	AllowSampling      bool    `json:"allow_sampling"`
	AllowLogprobs      bool    `json:"allow_logprobs"`
	AllowSearchIndices bool    `json:"allow_search_indices"`
	AllowView          bool    `json:"allow_view"`
	AllowFineTuning    bool    `json:"allow_fine_tuning"`
	Organization       string  `json:"organization"`
	Group              *string `json:"group"`
	IsBlocking         bool    `json:"is_blocking"`
}

type OpenAIModels struct {
	// Id model's name
	//
	// BUG: Different channels may have the same model name
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	// OwnedBy is the channel's adaptor name
	OwnedBy    string                  `json:"owned_by"`
	Permission []OpenAIModelPermission `json:"permission"`
	Root       string                  `json:"root"`
	Parent     *string                 `json:"parent"`
}

// BUG(#39): 更新 custom channel 时，应该同步更新所有自定义的 models 到 allModels
var allModels []OpenAIModels
var modelsMap map[string]OpenAIModels
var channelId2Models map[int][]string

func init() {
	var permission []OpenAIModelPermission
	permission = append(permission, OpenAIModelPermission{
		Id:                 "modelperm-LwHkVFn8AcMItP432fKKDIKJ",
		Object:             "model_permission",
		Created:            1626777600,
		AllowCreateEngine:  true,
		AllowSampling:      true,
		AllowLogprobs:      true,
		AllowSearchIndices: false,
		AllowView:          true,
		AllowFineTuning:    false,
		Organization:       "*",
		Group:              nil,
		IsBlocking:         false,
	})
	// https://platform.openai.com/docs/models/model-endpoint-compatibility
	for i := 0; i < apitype.Dummy; i++ {
		if i == apitype.AIProxyLibrary {
			continue
		}
		adaptor := relay.GetAdaptor(i)
		if adaptor == nil {
			continue
		}

		channelName := adaptor.GetChannelName()
		modelNames := adaptor.GetModelList()
		for _, modelName := range modelNames {
			allModels = append(allModels, OpenAIModels{
				Id:         modelName,
				Object:     "model",
				Created:    1626777600,
				OwnedBy:    channelName,
				Permission: permission,
				Root:       modelName,
				Parent:     nil,
			})
		}
	}
	for _, channelType := range openai.CompatibleChannels {
		if channelType == channeltype.Azure {
			continue
		}
		channelName, channelModelList := openai.GetCompatibleChannelMeta(channelType)
		for _, modelName := range channelModelList {
			allModels = append(allModels, OpenAIModels{
				Id:         modelName,
				Object:     "model",
				Created:    1626777600,
				OwnedBy:    channelName,
				Permission: permission,
				Root:       modelName,
				Parent:     nil,
			})
		}
	}
	modelsMap = make(map[string]OpenAIModels)
	for _, model := range allModels {
		modelsMap[model.Id] = model
	}
	channelId2Models = make(map[int][]string)
	for i := 1; i < channeltype.Dummy; i++ {
		adaptor := relay.GetAdaptor(channeltype.ToAPIType(i))
		if adaptor == nil {
			continue
		}

		meta := &meta.Meta{
			ChannelType: i,
		}
		adaptor.Init(meta)
		channelId2Models[i] = adaptor.GetModelList()
	}
}

func DashboardListModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    channelId2Models,
	})
}

func ListAllModels(c *gin.Context) {
	c.JSON(200, gin.H{
		"object": "list",
		"data":   allModels,
	})
}

// ModelsDisplayResponse represents the response structure for the models display page
type ModelsDisplayResponse struct {
	Success bool                                `json:"success"`
	Message string                              `json:"message"`
	Data    map[string]ChannelModelsDisplayInfo `json:"data"`
}

// ChannelModelsDisplayInfo represents model information for a specific channel/adaptor
type ChannelModelsDisplayInfo struct {
	ChannelName string                      `json:"channel_name"`
	ChannelType int                         `json:"channel_type"`
	Models      map[string]ModelDisplayInfo `json:"models"`
}

// ModelDisplayInfo represents display information for a single model
type ModelDisplayInfo struct {
	InputPrice  float64 `json:"input_price"`  // Price per 1M input tokens in USD
	OutputPrice float64 `json:"output_price"` // Price per 1M output tokens in USD
	MaxTokens   int32   `json:"max_tokens"`   // Maximum tokens limit, 0 means unlimited
}

// GetModelsDisplay returns models available to the current user grouped by channel/adaptor with pricing information
// This endpoint is designed for the Models display page in the frontend
func GetModelsDisplay(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	keyword := strings.ToLower(strings.TrimSpace(c.Query("keyword")))

	// Get user's group to determine available models
	userGroup, err := model.CacheGetUserGroup(userId)
	if err != nil {
		c.JSON(http.StatusOK, ModelsDisplayResponse{
			Success: false,
			Message: "Failed to get user group: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// Get available models with their channel information for this user group
	availableAbilities, err := model.CacheGetGroupModelsV2(c.Request.Context(), userGroup)
	if err != nil {
		c.JSON(http.StatusOK, ModelsDisplayResponse{
			Success: false,
			Message: "Failed to get available models: " + err.Error(),
			Data:    nil,
		})
		return
	}

	// Group models by individual channels to support Custom Channels and distinguish between multiple channels of the same type
	result := make(map[string]ChannelModelsDisplayInfo)
	channelToModels := make(map[int][]string)

	// Group models by individual channel ID
	for _, ability := range availableAbilities {
		if keyword != "" && !strings.Contains(strings.ToLower(ability.Model), keyword) { // apply keyword filter
			continue
		}
		channelToModels[ability.ChannelId] = append(channelToModels[ability.ChannelId], ability.Model)
	}

	// For each channel, get the channel information and pricing
	for channelId, models := range channelToModels {
		// Get channel information from database
		channel, err := model.GetChannelById(channelId, true)
		if err != nil {
			continue
		}

		// Get adaptor for this channel type
		adaptor := relay.GetAdaptor(channeltype.ToAPIType(channel.Type))
		if adaptor == nil {
			// For Custom Channels and other unsupported types, use OpenAI adaptor as fallback
			adaptor = relay.GetAdaptor(apitype.OpenAI)
			if adaptor == nil {
				continue
			}
		}

		meta := &meta.Meta{
			ChannelType: channel.Type,
		}
		adaptor.Init(meta)

		// Get channel type name and pricing
		channelTypeName := channeltype.IdToName(channel.Type)
		pricing := adaptor.GetDefaultModelPricing()

		// Get model mapping for this channel to resolve mapped model pricing
		modelMapping := channel.GetModelMapping()

		// Create display key in format {channel_type}:{channel_name}
		displayKey := fmt.Sprintf("%s:%s", channelTypeName, channel.Name)

		// Create models display info for available models only
		modelsInfo := make(map[string]ModelDisplayInfo)
		for _, modelName := range models {
			var inputPrice, outputPrice float64
			var maxTokens int32

			// Resolve the actual model name for pricing lookup
			// If the model is mapped, use the target model for pricing
			actualModelName := modelName
			if modelMapping != nil {
				if mappedModel, exists := modelMapping[modelName]; exists && mappedModel != "" {
					actualModelName = mappedModel
				}
			}

			if modelConfig, exists := pricing[actualModelName]; exists {
				// Convert pricing based on the format used by the adapter
				// Some adapters use quota-based pricing (ratio.MilliTokensUsd = 0.5)
				// Others use USD-based pricing (MilliTokensUsd = 0.000001)
				// We detect the format by checking the magnitude of the ratio
				if modelConfig.Ratio < 0.001 {
					// USD-based pricing (like AWS): ratio is already in USD per token
					// Just convert to USD per 1M tokens
					inputPrice = modelConfig.Ratio * 1000000
				} else {
					// Quota-based pricing (like OpenAI): ratio is in quota per token
					// Convert from quota to USD: ratio.QuotaPerUsd = 500000
					inputPrice = (modelConfig.Ratio * 1000000) / 500000
				}
				outputPrice = inputPrice * modelConfig.CompletionRatio
				maxTokens = modelConfig.MaxTokens
			} else {
				// Fallback to adapter methods if not in pricing map
				inputRatio := adaptor.GetModelRatio(actualModelName)
				completionRatio := adaptor.GetCompletionRatio(actualModelName)
				// Apply the same detection logic for fallback pricing
				if inputRatio < 0.001 {
					// USD-based pricing
					inputPrice = inputRatio * 1000000
				} else {
					// Quota-based pricing
					inputPrice = (inputRatio * 1000000) / 500000
				}
				outputPrice = inputPrice * completionRatio
				maxTokens = 0 // Default to unlimited
			}

			modelsInfo[modelName] = ModelDisplayInfo{
				InputPrice:  inputPrice,
				OutputPrice: outputPrice,
				MaxTokens:   maxTokens,
			}
		}

		if len(modelsInfo) > 0 {
			result[displayKey] = ChannelModelsDisplayInfo{
				ChannelName: displayKey,
				ChannelType: channel.Type,
				Models:      modelsInfo,
			}
		}
	}

	c.JSON(http.StatusOK, ModelsDisplayResponse{
		Success: true,
		Message: "",
		Data:    result,
	})
}

// ListModels lists all models available to the user.
func ListModels(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)

	userGroup, err := model.CacheGetUserGroup(userId)
	if err != nil {
		middleware.AbortWithError(c, http.StatusBadRequest, err)
		return
	}

	// Get available models with their channel names
	availableAbilities, err := model.CacheGetGroupModelsV2(c.Request.Context(), userGroup)
	if err != nil {
		middleware.AbortWithError(c, http.StatusBadRequest, err)
		return
	}

	// fix(#39): Previously, to fix #31, I concatenated model_name with adaptor name to return models.
	// But this caused an issue with custom channels, where the returned adaptor is "openai",
	// resulting in adaptor name and ownedBy field mismatches when matching against allModels.
	// For deepseek example, the adaptor is "openai" but ownedBy is "deepseek", causing mismatch.
	// Our current solution: for models from custom channels, don't concatenate adaptor name,
	// just match by model name only. However, this may reintroduce the duplicate models bug
	// mentioned in #31. A complete fix would require significant changes, so I'll leave it for now.

	// Create ability maps for both exact matches and model-only matches
	exactMatches := make(map[string]bool)
	modelMatches := make(map[string]bool)

	for _, ability := range availableAbilities {
		adaptor := relay.GetAdaptor(channeltype.ToAPIType(ability.ChannelType))
		// Store exact match
		key := ability.Model + ":" + adaptor.GetChannelName()
		exactMatches[key] = true

		// Store model name for fallback matching
		modelMatches[ability.Model] = true
	}

	userAvailableModels := make([]OpenAIModels, 0)
	for _, model := range allModels {
		key := model.Id + ":" + model.OwnedBy

		// Check for exact match first
		if exactMatches[key] {
			userAvailableModels = append(userAvailableModels, model)
			continue
		}

		// Fall back to model-only match if:
		// 1. Model name matches
		// 2. No exact match exists for this model name
		if modelMatches[model.Id] {
			hasExactMatch := false
			for exactKey := range exactMatches {
				if strings.HasPrefix(exactKey, model.Id+":") {
					hasExactMatch = true
					break
				}
			}

			if !hasExactMatch {
				userAvailableModels = append(userAvailableModels, model)
			}
		}
	}

	// Sort models alphabetically for consistent presentation
	sort.Slice(userAvailableModels, func(i, j int) bool {
		return userAvailableModels[i].Id < userAvailableModels[j].Id
	})

	c.JSON(200, gin.H{
		"object": "list",
		"data":   userAvailableModels,
	})
}

func RetrieveModel(c *gin.Context) {
	modelId := c.Param("model")
	if model, ok := modelsMap[modelId]; ok {
		c.JSON(200, model)
	} else {
		Error := relaymodel.Error{
			Message: fmt.Sprintf("The model '%s' does not exist", modelId),
			Type:    "invalid_request_error",
			Param:   "model",
			Code:    "model_not_found",
		}
		c.JSON(200, gin.H{
			"error": Error,
		})
	}
}

func GetUserAvailableModels(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.GetInt(ctxkey.Id)
	userGroup, err := model.CacheGetUserGroup(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	models, err := model.CacheGetGroupModelsV2(ctx, userGroup)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var modelNames []string
	modelsMap := map[string]bool{}
	for _, model := range models {
		modelsMap[model.Model] = true
	}
	for modelName := range modelsMap {
		modelNames = append(modelNames, modelName)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    modelNames,
	})
	return
}
