package middleware

import (
	"strings"

	"github.com/Laisky/errors/v2"
	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
)

// AbortWithError aborts the request with an error message
func AbortWithError(c *gin.Context, statusCode int, err error) {
	logger := gmw.GetLogger(c)
	if ignoreServerError(err) {
		logger.Warn("server abort",
			zap.Int("status_code", statusCode),
			zap.Error(err))
	} else {
		logger.Error("server abort",
			zap.Int("status_code", statusCode),
			zap.Error(err))
	}

	c.JSON(statusCode, gin.H{
		"error": gin.H{
			"message": helper.MessageWithRequestId(err.Error(), c.GetString(helper.RequestIdKey)),
			"type":    "one_api_error",
		},
	})
	c.Abort()
}

func ignoreServerError(err error) bool {
	switch {
	case strings.Contains(err.Error(), "token not found for key:"):
		return true
	default:
		return false
	}
}

func getRequestModel(c *gin.Context) (string, error) {
	// Realtime WS uses model in query string
	if strings.HasPrefix(c.Request.URL.Path, "/v1/realtime") {
		m := c.Query("model")
		if m == "" {
			return "", errors.New("missing required query parameter: model")
		}
		return m, nil
	}

	var modelRequest ModelRequest
	err := common.UnmarshalBodyReusable(c, &modelRequest)
	if err != nil {
		return "", errors.Wrap(err, "common.UnmarshalBodyReusable failed")
	}

	switch {
	case strings.HasPrefix(c.Request.URL.Path, "/v1/moderations"):
		if modelRequest.Model == "" {
			modelRequest.Model = "text-moderation-stable"
		}
	case strings.HasSuffix(c.Request.URL.Path, "embeddings"):
		if modelRequest.Model == "" {
			modelRequest.Model = c.Param("model")
		}
	case strings.HasPrefix(c.Request.URL.Path, "/v1/images/generations"),
		strings.HasPrefix(c.Request.URL.Path, "/v1/images/edits"):
		if modelRequest.Model == "" {
			modelRequest.Model = "dall-e-2"
		}
	case strings.HasPrefix(c.Request.URL.Path, "/v1/audio/transcriptions"),
		strings.HasPrefix(c.Request.URL.Path, "/v1/audio/translations"):
		if modelRequest.Model == "" {
			modelRequest.Model = "whisper-1"
		}
	}

	return modelRequest.Model, nil
}

func isModelInList(modelName string, models string) bool {
	modelList := strings.Split(models, ",")
	for _, model := range modelList {
		if modelName == model {
			return true
		}
	}
	return false
}

// GetTokenKeyParts extracts the token key parts from the Authorization header
//
// key like `sk-{token}[-{channelid}]`
func GetTokenKeyParts(c *gin.Context) []string {
	key := c.Request.Header.Get("Authorization")
	key = strings.TrimPrefix(key, "Bearer ")
	key = strings.TrimPrefix(strings.TrimPrefix(key, "sk-"), "laisky-")
	return strings.Split(key, "-")
}
