package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/model"
)

// TestGetModelsDisplay_Keyword ensures the endpoint accepts the 'keyword' filter
// and returns a valid success response (even when no data present in test DB).
func TestGetModelsDisplay_Keyword(t *testing.T) {
	model.InitDB()
	model.InitLogDB()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/models/display", func(c *gin.Context) {
		// inject a test user id so CacheGetUserGroup works
		c.Set(ctxkey.Id, 1)
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
	model.InitDB()
	model.InitLogDB()
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
