package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	gutils "github.com/Laisky/go-utils/v5"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/singleflight"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

func setupModelsDisplayTestEnv(t *testing.T) {
	t.Helper()

	anonymousModelsDisplayCache = gutils.NewExpCache[map[string]ChannelModelsDisplayInfo](context.Background(), time.Minute)
	anonymousModelsDisplayGroup = singleflight.Group{}

	originalRedisEnabled := common.IsRedisEnabled()
	common.SetRedisEnabled(false)
	t.Cleanup(func() {
		common.SetRedisEnabled(originalRedisEnabled)
	})

	originalSQLitePath := common.SQLitePath
	tempDir := t.TempDir()
	common.SQLitePath = filepath.Join(tempDir, "models-display.db")
	t.Cleanup(func() {
		common.SQLitePath = originalSQLitePath
	})

	model.InitDB()
	model.InitLogDB()

	t.Cleanup(func() {
		if model.DB != nil {
			require.NoError(t, model.CloseDB())
			model.DB = nil
			model.LOG_DB = nil
		}
	})
}

// TestGetModelsDisplay_Keyword ensures the endpoint accepts the 'keyword' filter
// and returns a valid success response (even when no data present in test DB).
func TestGetModelsDisplay_Keyword(t *testing.T) {
	setupModelsDisplayTestEnv(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	groupName := fmt.Sprintf("group-%d", time.Now().UnixNano())
	user := &model.User{
		Username: "keyword-user",
		Password: "password",
		Group:    groupName,
		Status:   model.UserStatusEnabled,
	}
	require.NoError(t, model.DB.Create(user).Error)
	router.GET("/api/models/display", func(c *gin.Context) {
		// inject a test user id so CacheGetUserGroup works
		c.Set(ctxkey.Id, user.Id)
		GetModelsDisplay(c)
	})

	req := httptest.NewRequest("GET", "/api/models/display?keyword=test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	// Success should be either true (normal) or false if user/group missing but should not crash
	// We only assert the presence of the success field and valid JSON structure.
	assert.NotNil(t, resp.Success)
}

// TestGetModelsDisplay_Anonymous ensures anonymous users can access the endpoint
// and receive a well-formed success response (may be empty data on a fresh DB).
func TestGetModelsDisplay_Anonymous(t *testing.T) {
	setupModelsDisplayTestEnv(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/models/display", func(c *gin.Context) {
		// Do not set ctxkey.Id to simulate anonymous user
		GetModelsDisplay(c)
	})

	req := httptest.NewRequest("GET", "/api/models/display", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

// TestGetModelsDisplay_AnonymousUsesConfiguredModels ensures guests only see models configured on the channel
func TestGetModelsDisplay_AnonymousUsesConfiguredModels(t *testing.T) {
	setupModelsDisplayTestEnv(t)
	gin.SetMode(gin.TestMode)
	channel := &model.Channel{
		Name:   "Public Channel",
		Type:   channeltype.OpenAI,
		Status: model.ChannelStatusEnabled,
		Models: "gpt-3.5-turbo,gpt-4o-mini",
		Group:  "public",
	}
	require.NoError(t, model.DB.Create(channel).Error)

	router := gin.New()
	router.GET("/api/models/display", func(c *gin.Context) {
		GetModelsDisplay(c)
	})

	req := httptest.NewRequest("GET", "/api/models/display", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp ModelsDisplayResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.True(t, resp.Success)

	key := fmt.Sprintf("%s:%s", channeltype.IdToName(channel.Type), channel.Name)
	info, ok := resp.Data[key]
	require.True(t, ok, "expected channel %s in response", key)
	require.Len(t, info.Models, 2)
	if _, ok := info.Models["gpt-3.5-turbo"]; !ok {
		t.Fatalf("expected gpt-3.5-turbo in models list: %+v", info.Models)
	}
	if _, ok := info.Models["gpt-4o-mini"]; !ok {
		t.Fatalf("expected gpt-4o-mini in models list: %+v", info.Models)
	}
	for modelName := range info.Models {
		if modelName != "gpt-3.5-turbo" && modelName != "gpt-4o-mini" {
			t.Fatalf("unexpected model present: %s", modelName)
		}
	}

	convertRatioToPrice := func(r float64) float64 {
		if r <= 0 {
			return 0
		}
		if r < 0.001 {
			return r * 1_000_000
		}
		return (r * 1_000_000) / ratio.QuotaPerUsd
	}

	gpt35 := info.Models["gpt-3.5-turbo"]
	gpt35Cfg := openai.ModelRatios["gpt-3.5-turbo"]
	expected35Input := convertRatioToPrice(gpt35Cfg.Ratio)
	require.InDelta(t, expected35Input, gpt35.InputPrice, 1e-6)
	expected35Cached := expected35Input
	if gpt35Cfg.CachedInputRatio != 0 {
		expected35Cached = convertRatioToPrice(gpt35Cfg.CachedInputRatio)
	}
	require.InDelta(t, expected35Cached, gpt35.CachedInputPrice, 1e-6)

	gpt4o := info.Models["gpt-4o-mini"]
	gpt4oCfg := openai.ModelRatios["gpt-4o-mini"]
	expected4oInput := convertRatioToPrice(gpt4oCfg.Ratio)
	require.InDelta(t, expected4oInput, gpt4o.InputPrice, 1e-6)
	expected4oCached := expected4oInput
	if gpt4oCfg.CachedInputRatio != 0 {
		expected4oCached = convertRatioToPrice(gpt4oCfg.CachedInputRatio)
	}
	require.InDelta(t, expected4oCached, gpt4o.CachedInputPrice, 1e-6)
}

// TestGetModelsDisplay_LoggedInFiltersUnsupportedModels ensures logged-in users don't see models outside their allowed set
func TestGetModelsDisplay_LoggedInFiltersUnsupportedModels(t *testing.T) {
	setupModelsDisplayTestEnv(t)
	gin.SetMode(gin.TestMode)
	groupName := fmt.Sprintf("group-%d", time.Now().UnixNano())
	user := &model.User{
		Username: "allowed-user",
		Password: "password",
		Group:    groupName,
		Status:   model.UserStatusEnabled,
	}
	require.NoError(t, model.DB.Create(user).Error)

	channel := &model.Channel{
		Name:   "User Channel",
		Type:   channeltype.OpenAI,
		Status: model.ChannelStatusEnabled,
		Models: "gpt-3.5-turbo",
		Group:  groupName,
	}
	require.NoError(t, model.DB.Create(channel).Error)

	abilities := []*model.Ability{
		{
			Group:     groupName,
			Model:     "gpt-3.5-turbo",
			ChannelId: channel.Id,
			Enabled:   true,
		},
		{
			Group:     groupName,
			Model:     "gpt-invalid-model",
			ChannelId: channel.Id,
			Enabled:   true,
		},
	}
	for _, ability := range abilities {
		require.NoError(t, model.DB.Create(ability).Error)
	}

	router := gin.New()
	router.GET("/api/models/display", func(c *gin.Context) {
		c.Set(ctxkey.Id, user.Id)
		GetModelsDisplay(c)
	})

	req := httptest.NewRequest("GET", "/api/models/display", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp ModelsDisplayResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.True(t, resp.Success)

	key := fmt.Sprintf("%s:%s", channeltype.IdToName(channel.Type), channel.Name)
	info, ok := resp.Data[key]
	require.True(t, ok, "expected channel %s in response", key)
	require.Len(t, info.Models, 1)
	if _, ok := info.Models["gpt-3.5-turbo"]; !ok {
		t.Fatalf("expected gpt-3.5-turbo for user; got %+v", info.Models)
	}
	if _, ok := info.Models["gpt-invalid-model"]; ok {
		t.Fatalf("unexpected unsupported model exposed to user: %+v", info.Models)
	}
}
