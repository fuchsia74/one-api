package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/i18n"
	"github.com/songquanpeng/one-api/model"
)

func GetOptions(c *gin.Context) {
	var options []*model.Option
	config.OptionMapRWMutex.Lock()
	for k, v := range config.OptionMap {
		if strings.HasSuffix(k, "Token") || strings.HasSuffix(k, "Secret") {
			continue
		}
		options = append(options, &model.Option{
			Key:   k,
			Value: helper.Interface2String(v),
		})
	}
	config.OptionMapRWMutex.Unlock()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    options,
	})
}

func UpdateOption(c *gin.Context) {
	var option model.Option
	err := json.NewDecoder(c.Request.Body).Decode(&option)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": i18n.Translate(c, "invalid_parameter"),
		})
		return
	}
	switch option.Key {
	case "Theme":
		if !config.ValidThemes[option.Value] {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "invalid theme",
			})
			return
		}
	case "GitHubOAuthEnabled":
		if option.Value == "true" && config.GitHubClientId == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Unable to enable GitHub OAuth, please fill in the GitHub Client Id and GitHub Client Secret first!",
			})
			return
		}
	case "EmailDomainRestrictionEnabled":
		if option.Value == "true" && len(config.EmailDomainWhitelist) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Unable to enable email domain restriction, please fill in the restricted email domains first!",
			})
			return
		}
	case "WeChatAuthEnabled":
		if option.Value == "true" && config.WeChatServerAddress == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Unable to enable WeChat login, please fill in the relevant configuration information for WeChat login first!",
			})
			return
		}
	case "TurnstileCheckEnabled":
		if option.Value == "true" && config.TurnstileSiteKey == "" {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "Unable to enable Turnstile verification, please fill in the relevant configuration information for Turnstile verification first!",
			})
			return
		}
	}
	err = model.UpdateOption(option.Key, option.Value)
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
}
