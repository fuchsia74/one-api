package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common/logger"
)

// DebugChannelModelConfigs prints detailed information about a channel's model configuration
func DebugChannelModelConfigs(channelId int) error {
	var channel Channel
	err := DB.Where("id = ?", channelId).First(&channel).Error
	if err != nil {
		return fmt.Errorf("failed to find channel %d: %w", channelId, err)
	}

	logger.Logger.Info(fmt.Sprintf("=== DEBUG CHANNEL %d ===", channelId))
	logger.Logger.Info(fmt.Sprintf("Name: %s", channel.Name))
	logger.Logger.Info(fmt.Sprintf("Type: %d", channel.Type))
	logger.Logger.Info(fmt.Sprintf("Status: %d", channel.Status))

	// Check ModelConfigs
	if channel.ModelConfigs != nil && *channel.ModelConfigs != "" && *channel.ModelConfigs != "{}" {
		logger.Logger.Info(fmt.Sprintf("ModelConfigs (raw): %s", *channel.ModelConfigs))

		// Try to parse as new format
		var newFormatConfigs map[string]ModelConfigLocal
		if err := json.Unmarshal([]byte(*channel.ModelConfigs), &newFormatConfigs); err == nil {
			logger.Logger.Info(fmt.Sprintf("ModelConfigs (new format) - %d models:", len(newFormatConfigs)))
			for modelName, config := range newFormatConfigs {
				logger.Logger.Info(fmt.Sprintf("  %s: ratio=%.6f, completion_ratio=%.2f, max_tokens=%d",
					modelName, config.Ratio, config.CompletionRatio, config.MaxTokens))
			}
		} else {
			// Try to parse as old format
			var oldFormatConfigs map[string]ModelConfig
			if err := json.Unmarshal([]byte(*channel.ModelConfigs), &oldFormatConfigs); err == nil {
				logger.Logger.Info(fmt.Sprintf("ModelConfigs (old format) - %d models:", len(oldFormatConfigs)))
				for modelName, config := range oldFormatConfigs {
					logger.Logger.Info(fmt.Sprintf("  %s: max_tokens=%d", modelName, config.MaxTokens))
				}
			} else {
				logger.Logger.Error("ModelConfigs parsing failed", zap.Error(err))
			}
		}
	} else {
		logger.Logger.Info("ModelConfigs: empty or null")
	}

	// Check ModelRatio
	if channel.ModelRatio != nil && *channel.ModelRatio != "" && *channel.ModelRatio != "{}" {
		logger.Logger.Info(fmt.Sprintf("ModelRatio (raw): %s", *channel.ModelRatio))
		var modelRatios map[string]float64
		if err := json.Unmarshal([]byte(*channel.ModelRatio), &modelRatios); err == nil {
			logger.Logger.Info(fmt.Sprintf("ModelRatio (parsed) - %d models:", len(modelRatios)))
			for modelName, ratio := range modelRatios {
				logger.Logger.Info(fmt.Sprintf("  %s: %.6f", modelName, ratio))
			}
		} else {
			logger.Logger.Error("ModelRatio parsing failed", zap.Error(err))
		}
	} else {
		logger.Logger.Info("ModelRatio: empty or null")
	}

	// Check CompletionRatio
	if channel.CompletionRatio != nil && *channel.CompletionRatio != "" && *channel.CompletionRatio != "{}" {
		logger.Logger.Info(fmt.Sprintf("CompletionRatio (raw): %s", *channel.CompletionRatio))
		var completionRatios map[string]float64
		if err := json.Unmarshal([]byte(*channel.CompletionRatio), &completionRatios); err == nil {
			logger.Logger.Info(fmt.Sprintf("CompletionRatio (parsed) - %d models:", len(completionRatios)))
			for modelName, ratio := range completionRatios {
				logger.Logger.Info(fmt.Sprintf("  %s: %.2f", modelName, ratio))
			}
		} else {
			logger.Logger.Error("CompletionRatio parsing failed", zap.Error(err))
		}
	} else {
		logger.Logger.Info("CompletionRatio: empty or null")
	}

	logger.Logger.Info("=== END DEBUG ===")
	return nil
}

// DebugAllChannelModelConfigs prints summary information about all channels
func DebugAllChannelModelConfigs() error {
	var channels []Channel
	err := DB.Select("id, name, type, status").Find(&channels).Error
	if err != nil {
		return fmt.Errorf("failed to fetch channels: %w", err)
	}

	logger.Logger.Info("=== ALL CHANNELS SUMMARY ===")
	for _, channel := range channels {
		// Get full channel data
		var fullChannel Channel
		err := DB.Where("id = ?", channel.Id).First(&fullChannel).Error
		if err != nil {
			logger.Logger.Error("Failed to load channel", zap.Int("channel_id", channel.Id), zap.Error(err))
			continue
		}

		hasModelConfigs := fullChannel.ModelConfigs != nil && *fullChannel.ModelConfigs != "" && *fullChannel.ModelConfigs != "{}"
		hasModelRatio := fullChannel.ModelRatio != nil && *fullChannel.ModelRatio != "" && *fullChannel.ModelRatio != "{}"
		hasCompletionRatio := fullChannel.CompletionRatio != nil && *fullChannel.CompletionRatio != "" && *fullChannel.CompletionRatio != "{}"

		status := "EMPTY"
		if hasModelConfigs {
			status = "UNIFIED"
		} else if hasModelRatio || hasCompletionRatio {
			status = "LEGACY"
		}

		logger.Logger.Info(fmt.Sprintf("Channel %d (%s) Type=%d Status=%s",
			channel.Id, channel.Name, channel.Type, status))

		if hasModelConfigs {
			// Count models in unified format
			var configs map[string]ModelConfigLocal
			if err := json.Unmarshal([]byte(*fullChannel.ModelConfigs), &configs); err == nil {
				var modelNames []string
				for modelName := range configs {
					modelNames = append(modelNames, modelName)
				}
				logger.Logger.Info(fmt.Sprintf("  Unified models (%d): %s", len(configs), strings.Join(modelNames, ", ")))
			}
		}

		if hasModelRatio {
			// Count models in legacy format
			var ratios map[string]float64
			if err := json.Unmarshal([]byte(*fullChannel.ModelRatio), &ratios); err == nil {
				var modelNames []string
				for modelName := range ratios {
					modelNames = append(modelNames, modelName)
				}
				logger.Logger.Info(fmt.Sprintf("  Legacy models (%d): %s", len(ratios), strings.Join(modelNames, ", ")))
			}
		}
	}
	logger.Logger.Info("=== END SUMMARY ===")
	return nil
}

// FixChannelModelConfigs attempts to fix a specific channel's model configuration
func FixChannelModelConfigs(channelId int) error {
	var channel Channel
	err := DB.Where("id = ?", channelId).First(&channel).Error
	if err != nil {
		return fmt.Errorf("failed to find channel %d: %w", channelId, err)
	}

	logger.Logger.Info(fmt.Sprintf("=== FIXING CHANNEL %d ===", channelId))

	// First, debug current state
	DebugChannelModelConfigs(channelId)

	// Clear any mixed model data and regenerate from adapter defaults
	logger.Logger.Info("Clearing mixed model data and regenerating from adapter defaults...")

	// Clear existing model configs
	emptyConfigs := "{}"
	channel.ModelConfigs = &emptyConfigs

	// Clear legacy data
	channel.ModelRatio = &emptyConfigs
	channel.CompletionRatio = &emptyConfigs

	// Get default pricing for this channel type from adapter
	logger.Logger.Info(fmt.Sprintf("Loading default pricing for channel type %d", channel.Type))
	defaultPricing := getChannelDefaultPricing(channel.Type)

	if defaultPricing != "" {
		logger.Logger.Info(fmt.Sprintf("Setting default model configs: %s", defaultPricing))
		channel.ModelConfigs = &defaultPricing
	} else {
		logger.Logger.Info("No default pricing available for this channel type")
	}

	// Save changes
	err = DB.Model(&channel).Updates(map[string]interface{}{
		"model_configs":    channel.ModelConfigs,
		"model_ratio":      channel.ModelRatio,
		"completion_ratio": channel.CompletionRatio,
	}).Error
	if err != nil {
		logger.Logger.Error("Failed to save fixed data", zap.Error(err))
		return err
	}
	logger.Logger.Info("Fixed data saved to database")

	// Debug final state
	logger.Logger.Info("Final state after fix:")
	DebugChannelModelConfigs(channelId)

	logger.Logger.Info("=== FIX COMPLETED ===")
	return nil
}

// CleanAllMixedModelData cleans all channels that have mixed model data
func CleanAllMixedModelData() error {
	logger.Logger.Info("=== CLEANING ALL MIXED MODEL DATA ===")

	var channels []Channel
	err := DB.Find(&channels).Error
	if err != nil {
		return fmt.Errorf("failed to fetch channels: %w", err)
	}

	cleanedCount := 0
	for _, channel := range channels {
		if channel.ModelConfigs != nil && *channel.ModelConfigs != "" && *channel.ModelConfigs != "{}" {
			// Check if this channel has mixed model data
			configs := channel.GetModelPriceConfigs()
			if configs != nil && len(configs) > 0 {
				hasMixedData := false
				channelTypeModels := getExpectedModelsForChannelType(channel.Type)

				for modelName := range configs {
					if !contains(channelTypeModels, modelName) {
						logger.Logger.Info(fmt.Sprintf("Channel %d (type %d) has unexpected model: %s", channel.Id, channel.Type, modelName))
						hasMixedData = true
						break
					}
				}

				if hasMixedData {
					logger.Logger.Info(fmt.Sprintf("Cleaning mixed data for channel %d", channel.Id))
					err := FixChannelModelConfigs(channel.Id)
					if err != nil {
						logger.Logger.Error("Failed to clean channel", zap.Int("channel_id", channel.Id), zap.Error(err))
					} else {
						cleanedCount++
					}
				}
			}
		}
	}

	logger.Logger.Info(fmt.Sprintf("Cleaned %d channels with mixed model data", cleanedCount))
	logger.Logger.Info("=== CLEANING COMPLETED ===")
	return nil
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to get expected models for a channel type
func getExpectedModelsForChannelType(channelType int) []string {
	// This is a simplified version - in practice, you'd want to get this from the adapters
	switch channelType {
	case 1: // OpenAI
		return []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini", "text-embedding-ada-002", "text-embedding-3-small", "text-embedding-3-large"}
	case 8: // Claude
		return []string{"claude-instant-1.2", "claude-2", "claude-2.0", "claude-2.1", "claude-3-opus-20240229", "claude-3-sonnet-20240229", "claude-3-haiku-20240307", "claude-3-5-haiku-20241022", "claude-3-5-sonnet-20240620", "claude-3-5-sonnet-20241022"}
	default:
		return []string{} // Allow all models for unknown types
	}
}

// Helper function to get default pricing for a channel type
func getChannelDefaultPricing(channelType int) string {
	// Generate appropriate default model configs for the channel type
	switch channelType {
	case 1: // OpenAI
		return `{
  "gpt-3.5-turbo": {
    "ratio": 0.0015,
    "completion_ratio": 2.0,
    "max_tokens": 16385
  },
  "gpt-4": {
    "ratio": 0.03,
    "completion_ratio": 2.0,
    "max_tokens": 8192
  },
  "gpt-4-turbo": {
    "ratio": 0.01,
    "completion_ratio": 3.0,
    "max_tokens": 128000
  },
  "gpt-4o": {
    "ratio": 0.005,
    "completion_ratio": 3.0,
    "max_tokens": 128000
  },
  "gpt-4o-mini": {
    "ratio": 0.00015,
    "completion_ratio": 4.0,
    "max_tokens": 128000
  }
}`
	case 8: // Claude
		return `{
  "claude-3-haiku-20240307": {
    "ratio": 0.00000025,
    "completion_ratio": 5.0,
    "max_tokens": 200000
  },
  "claude-3-sonnet-20240229": {
    "ratio": 0.000003,
    "completion_ratio": 5.0,
    "max_tokens": 200000
  },
  "claude-3-opus-20240229": {
    "ratio": 0.000015,
    "completion_ratio": 5.0,
    "max_tokens": 200000
  },
  "claude-3-5-sonnet-20240620": {
    "ratio": 0.000003,
    "completion_ratio": 5.0,
    "max_tokens": 200000
  },
  "claude-3-5-sonnet-20241022": {
    "ratio": 0.000003,
    "completion_ratio": 5.0,
    "max_tokens": 200000
  }
}`
	default:
		return "" // No default pricing for unknown types
	}
}

// ValidateAllChannelModelConfigs validates all channels and reports issues
func ValidateAllChannelModelConfigs() error {
	var channels []Channel
	err := DB.Find(&channels).Error
	if err != nil {
		return fmt.Errorf("failed to fetch channels: %w", err)
	}

	logger.Logger.Info("=== VALIDATION REPORT ===")

	validCount := 0
	issueCount := 0
	emptyCount := 0

	for _, channel := range channels {
		hasModelConfigs := channel.ModelConfigs != nil && *channel.ModelConfigs != "" && *channel.ModelConfigs != "{}"
		hasLegacyData := (channel.ModelRatio != nil && *channel.ModelRatio != "" && *channel.ModelRatio != "{}") ||
			(channel.CompletionRatio != nil && *channel.CompletionRatio != "" && *channel.CompletionRatio != "{}")

		if !hasModelConfigs && !hasLegacyData {
			emptyCount++
			continue
		}

		if hasModelConfigs {
			// Validate unified format
			var configs map[string]ModelConfigLocal
			if err := json.Unmarshal([]byte(*channel.ModelConfigs), &configs); err != nil {
				logger.Logger.Error("Channel: Invalid ModelConfigs JSON", zap.Int("channel_id", channel.Id), zap.Error(err))
				issueCount++
				continue
			}

			// Validate each model config
			if err := channel.validateModelPriceConfigs(configs); err != nil {
				logger.Logger.Error("Channel: Invalid ModelConfigs data", zap.Int("channel_id", channel.Id), zap.Error(err))
				issueCount++
				continue
			}

			validCount++
		} else if hasLegacyData {
			logger.Logger.Info(fmt.Sprintf("Channel %d: Has legacy data, needs migration", channel.Id))
			issueCount++
		}
	}

	logger.Logger.Info(fmt.Sprintf("Validation Summary: %d valid, %d issues, %d empty, %d total",
		validCount, issueCount, emptyCount, len(channels)))
	logger.Logger.Info("=== END VALIDATION ===")

	return nil
}
