package utils

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/Laisky/zap"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"

	"github.com/songquanpeng/one-api/common/logger"
)

// CrossRegionConfig holds configuration for cross-region inference
type CrossRegionConfig struct {
	EnableCrossRegion      bool
	PreferredRegions       []string
	MaxRetryAttempts       int
	RetryBackoffMultiplier float64
	CacheEnabled           bool
	CacheTTL               time.Duration
	FallbackTimeout        time.Duration
}

// RegionCapacityTracker tracks regional capacity and performance
type RegionCapacityTracker struct {
	mu              sync.RWMutex
	regionHealth    map[string]RegionHealth
	lastHealthCheck map[string]time.Time
	healthCheckTTL  time.Duration
}

// RegionHealth tracks the health status of a region
type RegionHealth struct {
	IsHealthy   bool
	LastError   error
	ErrorCount  int
	LastSuccess time.Time
	AvgLatency  time.Duration
	SuccessRate float64
}

// InferenceProfileCache caches inference profile information
type InferenceProfileCache struct {
	mu               sync.RWMutex
	profilesByRegion map[string][]string // region -> available profile IDs
	modelToProfile   map[string]string   // model -> profile mapping
	lastUpdated      map[string]time.Time
	ttl              time.Duration
}

var (
	// Global instances
	capacityTracker = &RegionCapacityTracker{
		regionHealth:    make(map[string]RegionHealth),
		lastHealthCheck: make(map[string]time.Time),
		healthCheckTTL:  5 * time.Minute,
	}

	profileCache = &InferenceProfileCache{
		profilesByRegion: make(map[string][]string),
		modelToProfile:   make(map[string]string),
		lastUpdated:      make(map[string]time.Time),
		ttl:              30 * time.Minute,
	}

	// Default configuration following AWS best practices
	DefaultCrossRegionConfig = CrossRegionConfig{
		EnableCrossRegion:      true,
		MaxRetryAttempts:       3,
		RetryBackoffMultiplier: 2.0,
		CacheEnabled:           true,
		CacheTTL:               30 * time.Minute,
		FallbackTimeout:        5 * time.Second,
	}
)

// CrossRegionInferences is a list of model IDs that support cross-region inference.
// This serves as the primary reference for supported models.
//
// https://docs.aws.amazon.com/bedrock/latest/userguide/inference-profiles-support.html
//
// document.querySelectorAll('pre.programlisting code').forEach((e) => {console.log(e.innerHTML)})
var CrossRegionInferences = []string{
	"us.amazon.nova-lite-v1:0",
	"us.amazon.nova-micro-v1:0",
	"us.amazon.nova-premier-v1:0",
	"us.amazon.nova-pro-v1:0",
	"us.anthropic.claude-3-5-haiku-20241022-v1:0",
	"us.anthropic.claude-3-5-sonnet-20240620-v1:0",
	"us.anthropic.claude-3-5-sonnet-20241022-v2:0",
	"us.anthropic.claude-3-7-sonnet-20250219-v1:0",
	"us.anthropic.claude-3-haiku-20240307-v1:0",
	"us.anthropic.claude-3-opus-20240229-v1:0",
	"us.anthropic.claude-3-sonnet-20240229-v1:0",
	"us.anthropic.claude-opus-4-20250514-v1:0",
	"us.anthropic.claude-opus-4-1-20250805-v1:0",
	"us.anthropic.claude-sonnet-4-20250514-v1:0",
	"us.deepseek.r1-v1:0",
	"us.meta.llama3-1-405b-instruct-v1:0",
	"us.meta.llama3-1-70b-instruct-v1:0",
	"us.meta.llama3-1-8b-instruct-v1:0",
	"us.meta.llama3-2-11b-instruct-v1:0",
	"us.meta.llama3-2-1b-instruct-v1:0",
	"us.meta.llama3-2-3b-instruct-v1:0",
	"us.meta.llama3-2-90b-instruct-v1:0",
	"us.meta.llama3-3-70b-instruct-v1:0",
	"us.meta.llama4-maverick-17b-instruct-v1:0",
	"us.meta.llama4-scout-17b-instruct-v1:0",
	"us.mistral.pixtral-large-2502-v1:0",
	"us.writer.palmyra-x4-v1:0",
	"us.writer.palmyra-x5-v1:0",
	"us-gov.anthropic.claude-3-5-sonnet-20240620-v1:0",
	"us-gov.anthropic.claude-3-haiku-20240307-v1:0",
	"eu.amazon.nova-lite-v1:0",
	"eu.amazon.nova-micro-v1:0",
	"eu.amazon.nova-pro-v1:0",
	"eu.anthropic.claude-3-5-sonnet-20240620-v1:0",
	"eu.anthropic.claude-3-7-sonnet-20250219-v1:0",
	"eu.anthropic.claude-3-haiku-20240307-v1:0",
	"eu.anthropic.claude-3-sonnet-20240229-v1:0",
	"eu.anthropic.claude-sonnet-4-20250514-v1:0",
	"eu.meta.llama3-2-1b-instruct-v1:0",
	"eu.meta.llama3-2-3b-instruct-v1:0",
	"eu.mistral.pixtral-large-2502-v1:0",
	"apac.amazon.nova-lite-v1:0",
	"apac.amazon.nova-micro-v1:0",
	"apac.amazon.nova-pro-v1:0",
	"apac.anthropic.claude-3-5-sonnet-20240620-v1:0",
	"apac.anthropic.claude-3-5-sonnet-20241022-v2:0",
	"apac.anthropic.claude-3-7-sonnet-20250219-v1:0",
	"apac.anthropic.claude-3-haiku-20240307-v1:0",
	"apac.anthropic.claude-3-sonnet-20240229-v1:0",
	"apac.anthropic.claude-sonnet-4-20250514-v1:0",
}

// RegionMapping defines the mapping between AWS regions and their cross-region inference prefixes
// Following AWS best practices for region grouping
var RegionMapping = map[string]string{
	// US regions
	"us-east-1":    "us",
	"us-east-2":    "us",
	"us-west-1":    "us",
	"us-west-2":    "us",
	"ca-central-1": "us", // Canada uses US inference profiles
	"sa-east-1":    "us", // South America uses US inference profiles

	// US Government regions
	"us-gov-east-1": "us-gov",
	"us-gov-west-1": "us-gov",

	// EU regions
	"eu-west-1":    "eu",
	"eu-west-2":    "eu",
	"eu-west-3":    "eu",
	"eu-central-1": "eu",
	"eu-north-1":   "eu",
	"eu-south-1":   "eu",
	"eu-central-2": "eu",

	// Asia Pacific regions
	"ap-southeast-1": "apac",
	"ap-southeast-2": "apac",
	"ap-northeast-1": "apac",
	"ap-northeast-2": "apac",
	"ap-south-1":     "apac",
	"ap-southeast-3": "apac",
	"ap-southeast-4": "apac",
	"ap-south-2":     "apac",
	"ap-northeast-3": "apac",
}

// getRegionPrefix returns the cross-region inference prefix for a given AWS region
func getRegionPrefix(region string) string {
	if prefix, exists := RegionMapping[region]; exists {
		return prefix
	}

	// Fallback to parsing logic for regions not in the map
	if strings.HasPrefix(region, "us-gov-") {
		return "us-gov"
	}

	parts := strings.Split(region, "-")
	if len(parts) == 0 {
		return ""
	}

	switch parts[0] {
	case "us":
		return "us"
	case "eu":
		return "eu"
	case "ap":
		return "apac"
	case "ca", "sa":
		// Canadian and South American regions typically use US inference profiles
		return "us"
	default:
		logger.Logger.Debug("unknown region prefix", zap.String("region", region))
		return ""
	}
}

// TestInferenceProfileAvailability tests if a cross-region inference profile is available
// by attempting a lightweight operation with the runtime client
func TestInferenceProfileAvailability(ctx context.Context, client *bedrockruntime.Client, profileID string) bool {
	profileCache.mu.RLock()
	cacheKey := client.Options().Region + ":" + profileID
	if availability, exists := profileCache.modelToProfile[cacheKey]; exists {
		lastUpdated := profileCache.lastUpdated[client.Options().Region]
		if time.Since(lastUpdated) < profileCache.ttl {
			profileCache.mu.RUnlock()
			return availability == "available"
		}
	}
	profileCache.mu.RUnlock()

	// Test with a minimal request to check availability
	// We'll try to invoke the model with minimal input to see if it's available
	available := testModelAvailability(ctx, client, profileID)

	// Update cache
	profileCache.mu.Lock()
	if available {
		profileCache.modelToProfile[cacheKey] = "available"
	} else {
		profileCache.modelToProfile[cacheKey] = "unavailable"
	}
	profileCache.lastUpdated[client.Options().Region] = time.Now()
	profileCache.mu.Unlock()

	return available
}

// testModelAvailability performs a lightweight test to check if the model is available
func testModelAvailability(ctx context.Context, client *bedrockruntime.Client, modelID string) bool {
	// Create a timeout context for the test
	testCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Use a minimal test request that should fail quickly if the model doesn't exist
	// but succeed (or fail with a different error) if it does exist
	testInput := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        []byte(`{"max_tokens": 1}`), // Minimal test payload
	}

	_, err := client.InvokeModel(testCtx, testInput)

	// If the error indicates the model doesn't exist, return false
	// If there's no error or a different error (like validation), the model likely exists
	if err != nil {
		errStr := err.Error()
		// AWS returns specific error codes for non-existent models
		if strings.Contains(errStr, "ValidationException") &&
			(strings.Contains(errStr, "model") || strings.Contains(errStr, "not found")) {
			return false
		}
	}

	return true
}

// Deprecated: File-based ARN configuration has been replaced with channel-specific configuration
// These functions are kept for backward compatibility but should not be used in new code

// ConvertModelID2CrossRegionProfile converts the model ID to a cross-region profile ID.
// Enhanced version that uses aws-sdk-go-v2 patterns and includes availability testing.
func ConvertModelID2CrossRegionProfile(ctx context.Context, model, region string) string {
	lg := gmw.GetLogger(ctx)

	regionPrefix := getRegionPrefix(region)
	if regionPrefix == "" {
		lg.Debug("unsupported region for cross-region inference", zap.String("region", region))
		return model
	}

	newModelID := regionPrefix + "." + model
	if slices.Contains(CrossRegionInferences, newModelID) {
		lg.Debug("convert model to cross-region profile", zap.String("model", model), zap.String("cross_region_profile", newModelID))
		return newModelID
	}

	lg.Debug("no cross-region profile found", zap.String("model", model), zap.String("region", region))
	return model
}

// ConvertModelID2CrossRegionProfileWithFallback provides enhanced conversion with runtime availability testing
func ConvertModelID2CrossRegionProfileWithFallback(ctx context.Context, model, region string, client *bedrockruntime.Client) string {
	lg := gmw.GetLogger(ctx)

	// Try cross-region profile first
	crossRegionModel := ConvertModelID2CrossRegionProfile(ctx, model, region)

	// If we got a cross-region profile and have a client, test availability
	if crossRegionModel != model && client != nil {
		if TestInferenceProfileAvailability(ctx, client, crossRegionModel) {
			lg.Debug("cross-region profile available", zap.String("cross_region_profile", crossRegionModel))
			return crossRegionModel
		}
		lg.Debug("cross-region profile not available, falling back", zap.String("cross_region_profile", crossRegionModel))
		return model
	}

	// If no client provided, return the cross-region profile anyway (best effort)
	// This allows the system to still attempt cross-region inference
	if crossRegionModel != model {
		lg.Debug("no client for availability testing, using cross-region profile", zap.String("cross_region_profile", crossRegionModel))
		return crossRegionModel
	}

	return model
}

// UpdateRegionHealthMetrics updates health metrics for a region based on operation results
func UpdateRegionHealthMetrics(region string, success bool, latency time.Duration, err error) {
	capacityTracker.mu.Lock()
	defer capacityTracker.mu.Unlock()

	health, exists := capacityTracker.regionHealth[region]
	if !exists {
		health = RegionHealth{
			IsHealthy:   true,
			SuccessRate: 1.0,
		}
	}

	now := time.Now()

	if success {
		health.LastSuccess = now
		health.ErrorCount = 0
		health.LastError = nil
		health.IsHealthy = true

		// Update average latency (simple moving average)
		if health.AvgLatency == 0 {
			health.AvgLatency = latency
		} else {
			health.AvgLatency = (health.AvgLatency + latency) / 2
		}

		// Update success rate
		health.SuccessRate = (health.SuccessRate + 1.0) / 2.0
	} else {
		health.ErrorCount++
		health.LastError = err
		health.SuccessRate = health.SuccessRate * 0.9 // Decay success rate

		// Mark as unhealthy if too many consecutive errors
		if health.ErrorCount >= 3 {
			health.IsHealthy = false
		}
	}

	capacityTracker.regionHealth[region] = health
	capacityTracker.lastHealthCheck[region] = now
}

// GetRegionHealth returns the current health status of a region
func GetRegionHealth(region string) RegionHealth {
	capacityTracker.mu.RLock()
	defer capacityTracker.mu.RUnlock()

	if health, exists := capacityTracker.regionHealth[region]; exists {
		return health
	}

	// Return default healthy status for unknown regions
	return RegionHealth{
		IsHealthy:   true,
		SuccessRate: 1.0,
	}
}
