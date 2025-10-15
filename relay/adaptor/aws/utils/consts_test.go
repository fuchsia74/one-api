package utils

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Laisky/errors/v2"
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
		{"Asia Pacific Northeast", "ap-northeast-1", "apac"},
		{"US Government East", "us-gov-east-1", "us-gov"},
		{"US Government West", "us-gov-west-1", "us-gov"},
		{"South America", "sa-east-1", "us"},
		{"Unknown region", "unknown-region-1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getRegionPrefix(tt.region)
			if result != tt.expected {
				t.Errorf("getRegionPrefix(%s) = %s, want %s", tt.region, result, tt.expected)
			}
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
			if result != tt.expected {
				t.Errorf("ConvertModelID2CrossRegionProfile(%s, %s) = %s, want %s",
					tt.model, tt.region, result, tt.expected)
			}
		})
	}
}

func TestUpdateRegionHealthMetrics(t *testing.T) {
	region := "us-east-1"

	// Test successful operation
	UpdateRegionHealthMetrics(region, true, 100*time.Millisecond, nil)
	health := GetRegionHealth(region)

	if !health.IsHealthy {
		t.Error("Expected region to be healthy after successful operation")
	}
	if health.ErrorCount != 0 {
		t.Errorf("Expected error count to be 0, got %d", health.ErrorCount)
	}
	if health.AvgLatency != 100*time.Millisecond {
		t.Errorf("Expected average latency to be 100ms, got %v", health.AvgLatency)
	}

	// Test failed operation
	testErr := errors.New("test error")
	UpdateRegionHealthMetrics(region, false, 0, testErr)
	health = GetRegionHealth(region)

	if health.ErrorCount != 1 {
		t.Errorf("Expected error count to be 1, got %d", health.ErrorCount)
	}
	if health.LastError == nil {
		t.Error("Expected LastError to be set")
	}

	// Test multiple failures to trigger unhealthy status
	for range 3 {
		UpdateRegionHealthMetrics(region, false, 0, testErr)
	}
	health = GetRegionHealth(region)

	if health.IsHealthy {
		t.Error("Expected region to be unhealthy after multiple failures")
	}
}

func TestConvertModelID2CrossRegionProfileWithFallback(t *testing.T) {
	ctx := context.Background()
	model := "anthropic.claude-3-haiku-20240307-v1:0"
	region := "us-east-1"

	// Test with nil client (should return cross-region profile for best effort)
	result := ConvertModelID2CrossRegionProfileWithFallback(ctx, model, region, nil)
	expected := "us.anthropic.claude-3-haiku-20240307-v1:0"
	if result != expected {
		t.Errorf("ConvertModelID2CrossRegionProfileWithFallback with nil client = %s, want %s",
			result, expected)
	}

	// Test that static conversion works independently
	staticResult := ConvertModelID2CrossRegionProfile(ctx, model, region)
	if staticResult != expected {
		t.Errorf("Static conversion = %s, want %s", staticResult, expected)
	}
}

func TestRegionMapping(t *testing.T) {
	// Test that all regions in RegionMapping have valid prefixes
	for region, expectedPrefix := range RegionMapping {
		actualPrefix := getRegionPrefix(region)
		if actualPrefix != expectedPrefix {
			t.Errorf("RegionMapping inconsistency: region %s maps to %s but getRegionPrefix returns %s",
				region, expectedPrefix, actualPrefix)
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
	}

	for _, modelID := range CrossRegionInferences {
		parts := strings.SplitN(modelID, ".", 2)
		if len(parts) != 2 {
			t.Errorf("Invalid cross-region model ID format: %s", modelID)
			continue
		}

		prefix := parts[0]
		if !validPrefixes[prefix] {
			t.Errorf("Invalid prefix %s in model ID: %s", prefix, modelID)
		}
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
