package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/pricing"
)

func GetAllChannels(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	// Get page size from query parameter, default to config value
	size, _ := strconv.Atoi(c.Query("size"))
	if size <= 0 {
		size = config.DefaultItemsPerPage
	}
	if size > config.MaxItemsPerPage {
		size = config.MaxItemsPerPage
	}

	sortBy := c.Query("sort")
	sortOrder := c.Query("order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	channels, err := model.GetAllChannels(p*size, size, "limited", sortBy, sortOrder)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Get total count for pagination
	totalCount, err := model.GetChannelCount()
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
		"data":    channels,
		"total":   totalCount,
	})
	return
}

func SearchChannels(c *gin.Context) {
	keyword := c.Query("keyword")
	sortBy := c.Query("sort")
	sortOrder := c.Query("order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	channels, err := model.SearchChannels(keyword, sortBy, sortOrder)
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
		"data":    channels,
	})
	return
}

func GetChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, false)
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
		"data":    channel,
	})
	return
}

func AddChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Disallow empty channel name
	if strings.TrimSpace(channel.Name) == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Channel name is required",
		})
		return
	}

	// Validate inference profile ARN map if provided
	if channel.InferenceProfileArnMap != nil && *channel.InferenceProfileArnMap != "" {
		err = model.ValidateInferenceProfileArnMapJSON(*channel.InferenceProfileArnMap)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Invalid inference profile ARN map: " + err.Error(),
			})
			return
		}
	}

	channel.CreatedTime = helper.GetTimestamp()
	// Sanitize testing model at creation: only keep if present in models list
	if channel.TestingModel != nil {
		tm := strings.TrimSpace(*channel.TestingModel)
		if tm == "" {
			channel.TestingModel = nil
		} else {
			ok := false
			for name := range strings.SplitSeq(channel.Models, ",") {
				if strings.TrimSpace(name) == tm {
					ok = true
					break
				}
			}
			if !ok {
				channel.TestingModel = nil
			}
		}
	}
	keys := strings.Split(channel.Key, "\n")
	channels := make([]model.Channel, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		localChannel := channel
		localChannel.Key = key
		// Auto-populate default BaseURL on creation if blank and default exists
		if (localChannel.BaseURL == nil || *localChannel.BaseURL == "") && localChannel.Type >= 0 {
			// Defensive bounds check against channeltype.ChannelBaseURLs
			if localChannel.Type < len(channeltype.ChannelBaseURLs) {
				def := channeltype.ChannelBaseURLs[localChannel.Type]
				if strings.TrimSpace(def) != "" {
					v := strings.TrimRight(def, "/")
					localChannel.BaseURL = &v
				}
			}
		}
		channels = append(channels, localChannel)
	}
	err = model.BatchInsertChannels(channels)
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

func DeleteChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	channel := model.Channel{Id: id}
	err := channel.Delete()
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

func DeleteDisabledChannel(c *gin.Context) {
	rows, err := model.DeleteDisabledChannel()
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
		"data":    rows,
	})
	return
}

func UpdateChannel(c *gin.Context) {
	statusOnly := c.Query("status_only")
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Validate inference profile ARN map if provided
	if channel.InferenceProfileArnMap != nil && *channel.InferenceProfileArnMap != "" {
		err = model.ValidateInferenceProfileArnMapJSON(*channel.InferenceProfileArnMap)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Invalid inference profile ARN map: " + err.Error(),
			})
			return
		}
	}

	if statusOnly != "" {
		// Only update status safely
		if channel.Id == 0 {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "Channel id is required"})
			return
		}
		model.UpdateChannelStatusById(channel.Id, channel.Status)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": ""})
		return
	}

	// Disallow empty name on full update
	if strings.TrimSpace(channel.Name) == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Channel name cannot be empty",
		})
		return
	}

	err = channel.Update()
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
		"data":    channel,
	})
	return
}

func GetChannelPricing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	channel, err := model.GetChannelById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Get from unified ModelConfigs only (after migration)
	modelRatio := channel.GetModelRatioFromConfigs()
	completionRatio := channel.GetCompletionRatioFromConfigs()

	// Also get the unified ModelConfigs
	modelConfigs := channel.GetModelPriceConfigs()

	// Debug logging to help identify data issues
	if modelConfigs != nil && len(modelConfigs) > 0 {
		var modelNames []string
		for modelName := range modelConfigs {
			modelNames = append(modelNames, modelName)
		}
		logger.Logger.Info("Channel returning model configs", zap.Int("id", channel.Id), zap.Int("type", channel.Type), zap.Any("models", modelNames))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"model_ratio":      modelRatio,
			"completion_ratio": completionRatio,
			"model_configs":    modelConfigs,
		},
	})
	return
}

func UpdateChannelPricing(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	var request struct {
		ModelRatio      map[string]float64                `json:"model_ratio"`
		CompletionRatio map[string]float64                `json:"completion_ratio"`
		ModelConfigs    map[string]model.ModelConfigLocal `json:"model_configs"`
	}

	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	channel, err := model.GetChannelById(id, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Handle both old format (separate model_ratio and completion_ratio) and new format (unified model_configs)
	if request.ModelConfigs != nil && len(request.ModelConfigs) > 0 {
		// New unified format - preferred approach
		err = channel.SetModelPriceConfigs(request.ModelConfigs)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Failed to set model configs: " + err.Error(),
			})
			return
		}
	} else if (request.ModelRatio != nil && len(request.ModelRatio) > 0) || (request.CompletionRatio != nil && len(request.CompletionRatio) > 0) {
		// Old format - convert to unified format automatically
		modelConfigs := make(map[string]model.ModelConfigLocal)

		// Collect all model names from both ratios
		allModelNames := make(map[string]bool)
		if request.ModelRatio != nil {
			for modelName := range request.ModelRatio {
				allModelNames[modelName] = true
			}
		}
		if request.CompletionRatio != nil {
			for modelName := range request.CompletionRatio {
				allModelNames[modelName] = true
			}
		}

		// Create ModelPriceLocal entries for each model
		for modelName := range allModelNames {
			config := model.ModelConfigLocal{}

			if request.ModelRatio != nil {
				if ratio, exists := request.ModelRatio[modelName]; exists {
					config.Ratio = ratio
				}
			}

			if request.CompletionRatio != nil {
				if completionRatio, exists := request.CompletionRatio[modelName]; exists {
					config.CompletionRatio = completionRatio
				}
			}

			// Only add if we have some pricing data
			if config.Ratio != 0 || config.CompletionRatio != 0 {
				modelConfigs[modelName] = config
			}
		}

		// Save to unified ModelConfigs only
		err = channel.SetModelPriceConfigs(modelConfigs)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Failed to set model configs: " + err.Error(),
			})
			return
		}
	}

	err = channel.Update()
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

func GetChannelDefaultPricing(c *gin.Context) {
	channelType, err := strconv.Atoi(c.Query("type"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Invalid channel type: " + err.Error(),
		})
		return
	}

	var defaultPricing map[string]adaptor.ModelConfig

	// For Custom channels and OpenAI-compatible channels, use global pricing from all adapters
	// This gives users access to pricing for all supported models
	if channelType == channeltype.Custom || channelType == channeltype.OpenAICompatible {
		// Use global pricing manager to get pricing from all adapters
		defaultPricing = pricing.GetGlobalModelPricing()
	} else {
		// For specific channel types, use their adapter's default pricing
		// Convert channel type to API type first
		apiType := channeltype.ToAPIType(channelType)
		adaptor := relay.GetAdaptor(apiType)
		if adaptor == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Unsupported channel type",
			})
			return
		}
		defaultPricing = adaptor.GetDefaultModelPricing()
	}

	// Separate model ratios and completion ratios for UI compatibility
	modelRatios := make(map[string]float64)
	completionRatios := make(map[string]float64)

	for model, price := range defaultPricing {
		modelRatios[model] = price.Ratio
		// Include all completion ratios, including 0 (which is valid pricing info)
		completionRatios[model] = price.CompletionRatio
	}

	// Create unified model configs format
	modelConfigs := make(map[string]model.ModelConfigLocal)
	for modelName, price := range defaultPricing {
		modelConfigs[modelName] = model.ModelConfigLocal{
			Ratio:           price.Ratio,
			CompletionRatio: price.CompletionRatio,
			MaxTokens:       price.MaxTokens,
		}
	}

	// Convert to JSON
	modelRatioJSON, err := json.Marshal(modelRatios)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to serialize model ratios: " + err.Error(),
		})
		return
	}

	completionRatioJSON, err := json.Marshal(completionRatios)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to serialize completion ratios: " + err.Error(),
		})
		return
	}

	modelConfigsJSON, err := json.Marshal(modelConfigs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Failed to serialize model configs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"model_ratio":      string(modelRatioJSON),
			"completion_ratio": string(completionRatioJSON),
			"model_configs":    string(modelConfigsJSON),
		},
	})
}
