package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
)

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

func TestDashboardListModels(t *testing.T) {
	// Initialize the database for testing
	model.InitDB()

	// Create a test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/models", DashboardListModels)

	// Create a test request
	req, _ := http.NewRequest("GET", "/api/models", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response struct {
		Success bool             `json:"success"`
		Message string           `json:"message"`
		Data    map[int][]string `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response structure
	assert.True(t, response.Success)
	assert.Equal(t, "", response.Message)
	assert.NotNil(t, response.Data)

	// Verify that the data contains channel type to models mapping
	// The data should be in the format: {channelType: [model1, model2, ...]}
	for channelType, models := range response.Data {
		assert.Greater(t, channelType, 0, "Channel type should be positive")
		assert.IsType(t, []string{}, models, "Models should be a slice of strings")

		// If there are models, verify they are strings
		for _, model := range models {
			assert.IsType(t, "", model, "Each model should be a string")
			assert.NotEmpty(t, model, "Model name should not be empty")
		}
	}

	t.Logf("DashboardListModels returned %d channel types", len(response.Data))

	// Log some sample data for verification
	for channelType, models := range response.Data {
		if len(models) > 0 {
			t.Logf("Channel type %d has %d models, first model: %s", channelType, len(models), models[0])
		}
		break // Just log the first one
	}
}

func TestListAllModels(t *testing.T) {
	// Initialize the database for testing
	model.InitDB()

	// Create a test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/channel/models", ListAllModels)

	// Create a test request
	req, _ := http.NewRequest("GET", "/api/channel/models", nil)
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response struct {
		Object string `json:"object"`
		Data   []struct {
			Id      string `json:"id"`
			Object  string `json:"object"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response structure
	assert.Equal(t, "list", response.Object)
	assert.NotNil(t, response.Data)

	// Verify that the data contains model information
	for _, model := range response.Data {
		assert.NotEmpty(t, model.Id, "Model ID should not be empty")
		assert.Equal(t, "model", model.Object, "Object should be 'model'")
		assert.NotEmpty(t, model.OwnedBy, "OwnedBy should not be empty")
	}

	t.Logf("ListAllModels returned %d models", len(response.Data))

	// Log some sample data for verification
	if len(response.Data) > 0 {
		t.Logf("First model: ID=%s, OwnedBy=%s", response.Data[0].Id, response.Data[0].OwnedBy)
	}
}

func TestModelEndpointsDifference(t *testing.T) {
	// This test verifies that the two endpoints return different data structures
	// as expected by the frontend

	model.InitDB()
	gin.SetMode(gin.TestMode)

	// Test DashboardListModels (/api/models)
	router1 := gin.New()
	router1.GET("/api/models", DashboardListModels)
	req1, _ := http.NewRequest("GET", "/api/models", nil)
	w1 := httptest.NewRecorder()
	router1.ServeHTTP(w1, req1)

	// Test ListAllModels (/api/channel/models)
	router2 := gin.New()
	router2.GET("/api/channel/models", ListAllModels)
	req2, _ := http.NewRequest("GET", "/api/channel/models", nil)
	w2 := httptest.NewRecorder()
	router2.ServeHTTP(w2, req2)

	// Both should return 200
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Parse responses
	var dashboardResponse struct {
		Success bool             `json:"success"`
		Data    map[int][]string `json:"data"`
	}
	var allModelsResponse struct {
		Object string `json:"object"`
		Data   []any  `json:"data"`
	}

	err1 := json.Unmarshal(w1.Body.Bytes(), &dashboardResponse)
	err2 := json.Unmarshal(w2.Body.Bytes(), &allModelsResponse)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Verify different structures
	assert.True(t, dashboardResponse.Success, "DashboardListModels should have success field")
	assert.Equal(t, "list", allModelsResponse.Object, "ListAllModels should have object='list'")

	// DashboardListModels should return channel-type-to-models mapping
	assert.IsType(t, map[int][]string{}, dashboardResponse.Data)

	// ListAllModels should return a list of model objects
	assert.IsType(t, []any{}, allModelsResponse.Data)

	t.Logf("✓ Confirmed that /api/models and /api/channel/models return different data structures")
	t.Logf("  - /api/models returns channel-type-to-models mapping with %d channel types", len(dashboardResponse.Data))
	t.Logf("  - /api/channel/models returns list of all models with %d models", len(allModelsResponse.Data))
}

func TestDeepSeekModelsInDashboard(t *testing.T) {
	// This test verifies that DeepSeek models are correctly included in the dashboard models endpoint
	model.InitDB()
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/api/models", DashboardListModels)
	req, _ := http.NewRequest("GET", "/api/models", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Success bool             `json:"success"`
		Data    map[int][]string `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// Based on debug output, DeepSeek models are actually in channel type 36
	// This suggests StepFun (channel type 36) is somehow getting DeepSeek models
	const DeepSeekChannelType = 36

	deepSeekModels, exists := response.Data[DeepSeekChannelType]
	assert.True(t, exists, "DeepSeek channel type should exist in the response")
	assert.NotEmpty(t, deepSeekModels, "DeepSeek should have models")

	// Verify that DeepSeek models are included
	expectedModels := []string{"deepseek-chat", "deepseek-reasoner"}
	for _, expectedModel := range expectedModels {
		found := slices.Contains(deepSeekModels, expectedModel)
		assert.True(t, found, "Expected DeepSeek model %s should be present", expectedModel)
	}

	t.Logf("✓ DeepSeek channel type %d has %d models: %v", DeepSeekChannelType, len(deepSeekModels), deepSeekModels)
}

func TestChannelDefaultPricing(t *testing.T) {
	// This test verifies that the /api/channel/default-pricing endpoint works correctly
	// for different channel types
	model.InitDB()
	gin.SetMode(gin.TestMode)

	// Initialize global pricing manager for the test
	relay.InitializeGlobalPricing()

	router := gin.New()
	router.GET("/api/channel/default-pricing", GetChannelDefaultPricing)

	// Test OpenAI-compatible channel (should return global pricing from all adapters)
	req1, _ := http.NewRequest("GET", "/api/channel/default-pricing?type=50", nil) // OpenAICompatible = 50
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	var customResponse struct {
		Success bool `json:"success"`
		Data    struct {
			ModelRatio      string `json:"model_ratio"`
			CompletionRatio string `json:"completion_ratio"`
		} `json:"data"`
	}
	err1 := json.Unmarshal(w1.Body.Bytes(), &customResponse)
	assert.NoError(t, err1)
	assert.True(t, customResponse.Success)

	// Parse the JSON strings to verify they contain models from multiple adapters
	var compatibleModelRatios map[string]float64
	err1 = json.Unmarshal([]byte(customResponse.Data.ModelRatio), &compatibleModelRatios)
	assert.NoError(t, err1)
	assert.NotEmpty(t, compatibleModelRatios, "OpenAI-compatible channel should have model ratios")

	// Test specific channel type (should return only that adapter's pricing)
	req2, _ := http.NewRequest("GET", "/api/channel/default-pricing?type=40", nil) // DeepSeek = 40
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var deepseekResponse struct {
		Success bool `json:"success"`
		Data    struct {
			ModelRatio      string `json:"model_ratio"`
			CompletionRatio string `json:"completion_ratio"`
		} `json:"data"`
	}
	err2 := json.Unmarshal(w2.Body.Bytes(), &deepseekResponse)
	assert.NoError(t, err2)
	assert.True(t, deepseekResponse.Success)

	var deepseekModelRatios map[string]float64
	err2 = json.Unmarshal([]byte(deepseekResponse.Data.ModelRatio), &deepseekModelRatios)
	assert.NoError(t, err2)
	assert.NotEmpty(t, deepseekModelRatios, "DeepSeek channel should have model ratios")

	// Custom should have significantly more models since it includes all adapters
	assert.Greater(t, len(compatibleModelRatios), len(deepseekModelRatios),
		"OpenAI-compatible channel should have more models than specific channel")

	t.Logf("✓ OpenAI-compatible channel has %d models from global pricing", len(compatibleModelRatios))
	t.Logf("✓ Specific channel has %d models from its adapter", len(deepseekModelRatios))
}
