package controller

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v6"
	gutils "github.com/Laisky/go-utils/v5"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/model"
	relay "github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/apitype"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
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

// Anonymous models display cache (1-minute TTL) to avoid repeated heavy loads.
// Keyed by normalized keyword filter.
var (
	anonymousModelsDisplayCache = gutils.NewExpCache[map[string]ChannelModelsDisplayInfo](context.Background(), time.Minute)
	anonymousModelsDisplayGroup singleflight.Group
)

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
	for i := range apitype.Dummy {
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
	InputPrice       float64 `json:"input_price"`           // Price per 1M input tokens in USD
	CachedInputPrice float64 `json:"cached_input_price"`    // Price per 1M cached input tokens in USD (falls back to input price when unspecified)
	OutputPrice      float64 `json:"output_price"`          // Price per 1M output tokens in USD
	MaxTokens        int32   `json:"max_tokens"`            // Maximum tokens limit, 0 means unlimited
	ImagePrice       float64 `json:"image_price,omitempty"` // USD per image (image models only)
}

// GetModelsDisplay returns models available to the current user grouped by channel/adaptor with pricing information
// This endpoint is designed for the Models display page in the frontend
func GetModelsDisplay(c *gin.Context) {
	// If logged-in, filter by user's allowed models; otherwise, show all supported models grouped by channel type
	userId := c.GetInt(ctxkey.Id)
	keyword := strings.ToLower(strings.TrimSpace(c.Query("keyword")))

	// Helper to build pricing info map for a channel with given model names
	convertRatioToPrice := func(r float64) float64 {
		if r <= 0 {
			return 0
		}
		if r < 0.001 {
			return r * 1_000_000
		}
		return (r * 1_000_000) / ratio.QuotaPerUsd
	}

	buildChannelModels := func(channel *model.Channel, modelNames []string) map[string]ModelDisplayInfo {
		result := make(map[string]ModelDisplayInfo)
		// Get adaptor for this channel type (fallback to OpenAI for unsupported/custom)
		adaptor := relay.GetAdaptor(channeltype.ToAPIType(channel.Type))
		if adaptor == nil {
			adaptor = relay.GetAdaptor(apitype.OpenAI)
			if adaptor == nil {
				return result
			}
		}
		m := &meta.Meta{ChannelType: channel.Type}
		adaptor.Init(m)

		pricing := adaptor.GetDefaultModelPricing()
		modelMapping := channel.GetModelMapping()

		for _, rawName := range modelNames {
			modelName := strings.TrimSpace(rawName)
			if modelName == "" {
				continue
			}
			if !channel.SupportsModel(modelName) {
				continue
			}
			if keyword != "" && !strings.Contains(strings.ToLower(modelName), keyword) {
				continue
			}
			// resolve mapped model for pricing
			actual := modelName
			if modelMapping != nil {
				if mapped, ok := modelMapping[modelName]; ok && mapped != "" {
					actual = mapped
				}
			}

			var inputPrice, cachedInputPrice, outputPrice float64
			var maxTokens int32
			var imagePrice float64

			if cfg, ok := pricing[actual]; ok {
				if cfg.ImagePriceUsd > 0 && cfg.Ratio == 0 {
					result[modelName] = ModelDisplayInfo{
						MaxTokens:        cfg.MaxTokens,
						ImagePrice:       cfg.ImagePriceUsd,
						InputPrice:       0,
						CachedInputPrice: 0,
					}
					continue
				}
				inputPrice = convertRatioToPrice(cfg.Ratio)
				cachedInputPrice = inputPrice
				if cfg.CachedInputRatio != 0 {
					cachedInputPrice = convertRatioToPrice(cfg.CachedInputRatio)
				}
				outputPrice = inputPrice * cfg.CompletionRatio
				maxTokens = cfg.MaxTokens
				imagePrice = cfg.ImagePriceUsd
			} else {
				inRatio := adaptor.GetModelRatio(actual)
				compRatio := adaptor.GetCompletionRatio(actual)
				inputPrice = convertRatioToPrice(inRatio)
				cachedInputPrice = inputPrice
				outputPrice = inputPrice * compRatio
				maxTokens = 0
				imagePrice = 0
			}

			result[modelName] = ModelDisplayInfo{
				InputPrice:       inputPrice,
				CachedInputPrice: cachedInputPrice,
				OutputPrice:      outputPrice,
				MaxTokens:        maxTokens,
				ImagePrice:       imagePrice,
			}
		}
		return result
	}

	// If userId is zero, treat as anonymous: list all channels and their supported models from DB and adaptor
	if userId == 0 {
		// Anonymous path with cache + singleflight to mitigate DB load and thundering herd
		cacheKey := "kw:" + keyword
		if data, ok := anonymousModelsDisplayCache.Load(cacheKey); ok {
			c.JSON(http.StatusOK, ModelsDisplayResponse{Success: true, Message: "", Data: data})
			return
		}

		v, err, _ := anonymousModelsDisplayGroup.Do(cacheKey, func() (any, error) {
			channels, err := model.GetAllEnabledChannels()
			if err != nil {
				return nil, errors.Wrap(err, "get all enabled channels")
			}
			result := make(map[string]ChannelModelsDisplayInfo)
			for _, ch := range channels {
				supported := ch.GetSupportedModelNames()
				if len(supported) == 0 {
					continue
				}
				modelInfos := buildChannelModels(ch, supported)
				if len(modelInfos) == 0 {
					continue
				}
				key := fmt.Sprintf("%s:%s", channeltype.IdToName(ch.Type), ch.Name)
				result[key] = ChannelModelsDisplayInfo{ChannelName: key, ChannelType: ch.Type, Models: modelInfos}
			}
			anonymousModelsDisplayCache.Store(cacheKey, result)
			return result, nil
		})
		if err != nil {
			c.JSON(http.StatusOK, ModelsDisplayResponse{Success: false, Message: "Failed to load channels: " + err.Error()})
			return
		}
		data := v.(map[string]ChannelModelsDisplayInfo)
		c.JSON(http.StatusOK, ModelsDisplayResponse{Success: true, Message: "", Data: data})
		return
	}

	// Logged-in path: show only models allowed for the user group
	ctx := gmw.Ctx(c)
	userGroup, err := model.CacheGetUserGroup(ctx, userId)
	if err != nil {
		c.JSON(http.StatusOK, ModelsDisplayResponse{Success: false, Message: "Failed to get user group: " + err.Error()})
		return
	}
	abilities, err := model.CacheGetGroupModelsV2(ctx, userGroup)
	if err != nil {
		c.JSON(http.StatusOK, ModelsDisplayResponse{Success: false, Message: "Failed to get available models: " + err.Error()})
		return
	}

	result := make(map[string]ChannelModelsDisplayInfo)
	// Group abilities by channel ID and deduplicate models
	ch2models := make(map[int]map[string]struct{})
	for _, ab := range abilities {
		if _, ok := ch2models[ab.ChannelId]; !ok {
			ch2models[ab.ChannelId] = make(map[string]struct{})
		}
		ch2models[ab.ChannelId][ab.Model] = struct{}{}
	}
	for chID, modelSet := range ch2models {
		ch, err := model.GetChannelById(chID, true)
		if err != nil {
			continue
		}
		models := make([]string, 0, len(modelSet))
		for m := range modelSet {
			if ch.SupportsModel(m) {
				models = append(models, m)
			}
		}
		if len(models) == 0 {
			continue
		}
		sort.Strings(models)
		infos := buildChannelModels(ch, models)
		if len(infos) == 0 {
			continue
		}
		key := fmt.Sprintf("%s:%s", channeltype.IdToName(ch.Type), ch.Name)
		result[key] = ChannelModelsDisplayInfo{ChannelName: key, ChannelType: ch.Type, Models: infos}
	}

	c.JSON(http.StatusOK, ModelsDisplayResponse{Success: true, Message: "", Data: result})
}

// ListModels lists all models available to the user.
func ListModels(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)

	ctx := gmw.Ctx(c)
	userGroup, err := model.CacheGetUserGroup(ctx, userId)
	if err != nil {
		middleware.AbortWithError(c, http.StatusBadRequest, err)
		return
	}

	// Get available models with their channel names
	availableAbilities, err := model.CacheGetGroupModelsV2(ctx, userGroup)
	if err != nil {
		middleware.AbortWithError(c, http.StatusBadRequest, err)
		return
	}

	// fix(#39): Previously, to fix #31, I concatenated model_name with adaptor name to return models.
	// But this caused an issue with OpenAI-compatible channels (including legacy custom entries), where the returned adaptor is "openai",
	// resulting in adaptor name and ownedBy field mismatches when matching against allModels.
	// For deepseek example, the adaptor is "openai" but ownedBy is "deepseek", causing mismatch.
	// Our current solution: for models from OpenAI-compatible channels, don't concatenate adaptor name,
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
		msg := fmt.Sprintf("The model '%s' does not exist", modelId)
		Error := relaymodel.Error{Message: msg, Type: "invalid_request_error", Param: "model", Code: "model_not_found", RawError: errors.New(msg)}
		c.JSON(200, gin.H{
			"error": Error,
		})
	}
}

func GetUserAvailableModels(c *gin.Context) {
	ctx := gmw.Ctx(c)
	id := c.GetInt(ctxkey.Id)
	userGroup, err := model.CacheGetUserGroup(ctx, id)
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
}

// GetAvailableModelsByToken get available models by API token
func GetAvailableModelsByToken(c *gin.Context) {
	// Get token information to determine status
	tokenID := c.GetInt(ctxkey.TokenId)
	userID := c.GetInt(ctxkey.Id)
	token, err := model.GetTokenByIds(tokenID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
			"data": gin.H{
				"available": nil,
				"enabled":   false,
			},
		})
		return
	}

	// Determine if token is enabled
	statusToken := token.Status == model.TokenStatusEnabled

	// Check if the token has specific model restrictions
	if availableModels, exists := c.Get(ctxkey.AvailableModels); exists {
		// Token has model restrictions, use those models
		modelsString := availableModels.(string)
		if modelsString != "" {
			modelNames := strings.Split(modelsString, ",")
			// Trim whitespace from each model name
			for i := range modelNames {
				modelNames[i] = strings.TrimSpace(modelNames[i])
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"available": modelNames,
					"enabled":   statusToken,
				},
			})
			return
		}
	}

	// Token has no model restrictions, return error instead of fallback
	c.JSON(http.StatusOK, gin.H{
		"success": false,
		"message": "the token has no available models",
		"data": gin.H{
			"available": nil,
			"enabled":   statusToken,
		},
	})
}
