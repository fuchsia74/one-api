package meta

import (
	"strings"
	"time"

	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

type Meta struct {
	Mode         int
	ChannelType  int
	ChannelId    int
	TokenId      int
	TokenName    string
	UserId       int
	Group        string
	ModelMapping map[string]string
	// BaseURL is the proxy url set in the channel config
	BaseURL  string
	APIKey   string
	APIType  int
	Config   model.ChannelConfig
	IsStream bool
	// OriginModelName is the model name from the raw user request
	OriginModelName string
	// ActualModelName is the model name after mapping
	ActualModelName    string
	RequestURLPath     string
	PromptTokens       int // only for DoResponse
	ChannelRatio       float64
	ForcedSystemPrompt string
	StartTime          time.Time
}

// GetMappedModelName returns the mapped model name and a bool indicating if the model name is mapped
func GetMappedModelName(modelName string, mapping map[string]string) string {
	if mapping == nil {
		return modelName
	}

	mappedModelName := mapping[modelName]
	if mappedModelName != "" {
		return mappedModelName
	}

	return modelName
}

func GetByContext(c *gin.Context) *Meta {
	if v, ok := c.Get(ctxkey.Meta); ok {
		existingMeta := v.(*Meta)
		// Check if channel information has changed (indicating a retry with new channel)
		currentChannelId := c.GetInt(ctxkey.ChannelId)
		if existingMeta.ChannelId != currentChannelId && currentChannelId != 0 {
			// Channel has changed, update the cached meta with new channel information
			logger.Logger.Info("Channel changed during retry", zap.Int("from", existingMeta.ChannelId), zap.Int("to", currentChannelId), zap.String("action", "updating meta"))
			existingMeta.ChannelType = c.GetInt(ctxkey.Channel)
			existingMeta.ChannelId = currentChannelId
			existingMeta.BaseURL = c.GetString(ctxkey.BaseURL)
			existingMeta.APIKey = strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
			existingMeta.ChannelRatio = c.GetFloat64(ctxkey.ChannelRatio)
			existingMeta.ModelMapping = c.GetStringMapString(ctxkey.ModelMapping)
			existingMeta.ForcedSystemPrompt = c.GetString(ctxkey.SystemPrompt)

			// Update config
			if cfg, ok := c.Get(ctxkey.Config); ok {
				existingMeta.Config = cfg.(model.ChannelConfig)
			}

			// Update BaseURL fallback if needed
			if existingMeta.BaseURL == "" {
				existingMeta.BaseURL = channeltype.ChannelBaseURLs[existingMeta.ChannelType]
			}

			// Update API type and actual model name
			existingMeta.APIType = channeltype.ToAPIType(existingMeta.ChannelType)
			existingMeta.ActualModelName = GetMappedModelName(existingMeta.OriginModelName, existingMeta.ModelMapping)

			// Update the cached meta in context
			Set2Context(c, existingMeta)
		}
		return existingMeta
	}

	meta := Meta{
		Mode:               relaymode.GetByPath(c.Request.URL.Path),
		ChannelType:        c.GetInt(ctxkey.Channel),
		ChannelId:          c.GetInt(ctxkey.ChannelId),
		TokenId:            c.GetInt(ctxkey.TokenId),
		TokenName:          c.GetString(ctxkey.TokenName),
		UserId:             c.GetInt(ctxkey.Id),
		Group:              c.GetString(ctxkey.Group),
		ModelMapping:       c.GetStringMapString(ctxkey.ModelMapping),
		OriginModelName:    c.GetString(ctxkey.RequestModel),
		ActualModelName:    c.GetString(ctxkey.RequestModel),
		BaseURL:            c.GetString(ctxkey.BaseURL),
		APIKey:             strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer "),
		RequestURLPath:     c.Request.URL.String(),
		ChannelRatio:       c.GetFloat64(ctxkey.ChannelRatio), // add by Laisky
		ForcedSystemPrompt: c.GetString(ctxkey.SystemPrompt),
		StartTime:          time.Now(),
	}
	cfg, ok := c.Get(ctxkey.Config)
	if ok {
		meta.Config = cfg.(model.ChannelConfig)
	}
	if meta.BaseURL == "" {
		meta.BaseURL = channeltype.ChannelBaseURLs[meta.ChannelType]
	}
	meta.APIType = channeltype.ToAPIType(meta.ChannelType)

	meta.ActualModelName = GetMappedModelName(meta.OriginModelName, meta.ModelMapping)

	Set2Context(c, &meta)
	return &meta
}

func Set2Context(c *gin.Context, meta *Meta) {
	c.Set(ctxkey.Meta, meta)
}
