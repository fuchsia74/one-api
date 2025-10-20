package utils

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/stretchr/testify/require"
)

func TestGetRegionPrefix(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		expected string
	}{
		{"US East 1", "us-east-1", "us"},
		{"US West 2", "us-west-2", "us"},
		{"Canada Central", "ca-central-1", "us"},
		{"EU West 1", "eu-west-1", "eu"},
		{"EU Central", "eu-central-1", "eu"},
		{"Asia Pacific Southeast", "ap-southeast-1", "apac"},
		{"Asia Pacific Northeast", "ap-northeast-1", "jp"},
		{"US Government East", "us-gov-east-1", "us-gov"},
		{"US Government West", "us-gov-west-1", "us-gov"},
		{"South America", "sa-east-1", "us"},
		{"Unknown region", "unknown-region-1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRegionPrefix(tt.region)
			require.Equalf(t, tt.expected, result, "getRegionPrefix(%s)", tt.region)
		})
	}
}

func TestConvertModelID2CrossRegionProfile(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		model    string
		region   string
		expected string
	}{
		{
			name:     "US region with supported model",
			model:    "anthropic.claude-3-haiku-20240307-v1:0",
			region:   "us-east-1",
			expected: "us.anthropic.claude-3-haiku-20240307-v1:0",
		},
		{
			name:     "EU region with supported model",
			model:    "anthropic.claude-3-sonnet-20240229-v1:0",
			region:   "eu-west-1",
			expected: "eu.anthropic.claude-3-sonnet-20240229-v1:0",
		},
		{
			name:     "APAC region with supported model",
			model:    "anthropic.claude-3-5-sonnet-20240620-v1:0",
			region:   "ap-southeast-1",
			expected: "apac.anthropic.claude-3-5-sonnet-20240620-v1:0",
		},
		{
			name:     "Japan region prefers JP profile when available",
			model:    "anthropic.claude-haiku-4-5-20251001-v1:0",
			region:   "ap-northeast-1",
			expected: "global.anthropic.claude-haiku-4-5-20251001-v1:0",
		},
		{
			name:     "Global profile when source region is allowed",
			model:    "anthropic.claude-sonnet-4-20250514-v1:0",
			region:   "us-west-2",
			expected: "global.anthropic.claude-sonnet-4-20250514-v1:0",
		},
		{
			name:     "Global profile falls back to geography when source region unsupported",
			model:    "anthropic.claude-sonnet-4-20250514-v1:0",
			region:   "eu-central-1",
			expected: "eu.anthropic.claude-sonnet-4-20250514-v1:0",
		},
		{
			name:     "Australian region prefers AU prefix when available",
			model:    "anthropic.claude-sonnet-4-5-20250929-v1:0",
			region:   "ap-southeast-2",
			expected: "global.anthropic.claude-sonnet-4-5-20250929-v1:0",
		},
		{
			name:     "Unsupported model returns original",
			model:    "unsupported.model-v1:0",
			region:   "us-east-1",
			expected: "unsupported.model-v1:0",
		},
		{
			name:     "Unsupported region returns original",
			model:    "anthropic.claude-3-haiku-20240307-v1:0",
			region:   "unknown-region-1",
			expected: "anthropic.claude-3-haiku-20240307-v1:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertModelID2CrossRegionProfile(ctx, tt.model, tt.region)
			require.Equalf(t, tt.expected, result, "ConvertModelID2CrossRegionProfile(%s, %s)", tt.model, tt.region)
		})
	}
}

func TestUpdateRegionHealthMetrics(t *testing.T) {
	region := "us-east-1"

	// Test successful operation
	UpdateRegionHealthMetrics(region, true, 100*time.Millisecond, nil)
	health := GetRegionHealth(region)

	require.True(t, health.IsHealthy, "region should be healthy after successful operation")
	require.Zero(t, health.ErrorCount, "error count should remain zero after success")
	require.Equal(t, 100*time.Millisecond, health.AvgLatency, "average latency should match successful operation")

	// Test failed operation
	testErr := errors.New("test error")
	UpdateRegionHealthMetrics(region, false, 0, testErr)
	health = GetRegionHealth(region)

	require.Equal(t, 1, health.ErrorCount, "error count should increment after failure")
	require.NotNil(t, health.LastError, "last error should be recorded")

	// Test multiple failures to trigger unhealthy status
	for range 3 {
		UpdateRegionHealthMetrics(region, false, 0, testErr)
	}
	health = GetRegionHealth(region)

	require.False(t, health.IsHealthy, "region should transition to unhealthy after repeated failures")
}

func TestConvertModelID2CrossRegionProfileWithFallback(t *testing.T) {
	ctx := context.Background()
	model := "anthropic.claude-3-haiku-20240307-v1:0"
	region := "us-east-1"

	// Test with nil client (should return cross-region profile for best effort)
	result := ConvertModelID2CrossRegionProfileWithFallback(ctx, model, region, nil)
	expected := "us.anthropic.claude-3-haiku-20240307-v1:0"
	require.Equal(t, expected, result, "fallback conversion should return cross-region profile when available")

	// Test that static conversion works independently
	staticResult := ConvertModelID2CrossRegionProfile(ctx, model, region)
	require.Equal(t, expected, staticResult, "static conversion should match fallback conversion")
}

func TestRegionMapping(t *testing.T) {
	// Test that all regions in RegionMapping have valid prefixes
	for region, prefixes := range RegionMapping {
		require.NotEmptyf(t, prefixes, "RegionMapping entry for %s should define at least one prefix", region)

		actualPrefix := getRegionPrefix(region)
		require.Equalf(t, prefixes[0], actualPrefix, "primary prefix mismatch for region %s", region)

		for _, prefix := range prefixes {
			require.NotEmptyf(t, prefix, "RegionMapping entry for %s contains an empty prefix", region)
		}
	}
}

func TestCrossRegionInferencesValidation(t *testing.T) {
	// Test that all cross-region inference models have valid prefixes
	validPrefixes := map[string]bool{
		"us":     true,
		"us-gov": true,
		"eu":     true,
		"apac":   true,
		"global": true,
		"ca":     true,
		"jp":     true,
		"au":     true,
	}

	for _, modelID := range CrossRegionInferences {
		parts := strings.SplitN(modelID, ".", 2)
		require.Lenf(t, parts, 2, "invalid cross-region model ID format: %s", modelID)

		prefix := parts[0]
		require.Truef(t, validPrefixes[prefix], "invalid prefix %s in model ID: %s", prefix, modelID)
	}
}

func BenchmarkConvertModelID2CrossRegionProfile(b *testing.B) {
	ctx := context.Background()
	model := "anthropic.claude-3-haiku-20240307-v1:0"
	region := "us-east-1"

	for b.Loop() {
		ConvertModelID2CrossRegionProfile(ctx, model, region)
	}
}

func BenchmarkGetRegionPrefix(b *testing.B) {
	region := "us-east-1"

	for b.Loop() {
		getRegionPrefix(region)
	}
}
