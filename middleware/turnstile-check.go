package middleware

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/Laisky/errors/v2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
)

type turnstileCheckResponse struct {
	Success bool `json:"success"`
}

func TurnstileCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.TurnstileCheckEnabled {
			session := sessions.Default(c)
			turnstileChecked := session.Get("turnstile")
			if turnstileChecked != nil {
				c.Next()
				return
			}
			response := c.Query("turnstile")
			if response == "" {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "Turnstile token is empty",
				})
				c.Abort()
				return
			}
			rawRes, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", url.Values{
				"secret":   {config.TurnstileSecretKey},
				"response": {response},
				"remoteip": {c.ClientIP()},
			})
			if err != nil {
				AbortWithError(c, http.StatusOK, errors.Wrap(err, "turnstile check request failed"))
				return
			}
			defer rawRes.Body.Close()
			var res turnstileCheckResponse
			err = json.NewDecoder(rawRes.Body).Decode(&res)
			if err != nil {
				AbortWithError(c, http.StatusOK, errors.Wrap(err, "turnstile response decode failed"))
				return
			}
			if !res.Success {
				AbortWithError(c, http.StatusOK, errors.New("turnstile verification failed"))
				return
			}
			session.Set("turnstile", true)
			err = session.Save()
			if err != nil {
				AbortWithError(c, http.StatusOK, errors.Wrap(err, "unable to save turnstile session information"))
				return
			}
		}
		c.Next()
	}
}
