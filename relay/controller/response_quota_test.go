package controller

import (
	"math"
	"testing"

	"github.com/songquanpeng/one-api/common/config"
)

func TestCalculateResponseAPIPreconsumeQuotaBackground(t *testing.T) {
	t.Parallel()

	maxOutput := 1000
	inputRatio := 1.0
	completionMultiplier := 2.0
	outputRatio := inputRatio * completionMultiplier

	quota := calculateResponseAPIPreconsumeQuota(200, &maxOutput, inputRatio, outputRatio, true)

	expectedMin := int64(math.Ceil(float64(config.PreconsumeTokenForBackgroundRequest) * outputRatio))
	if quota < expectedMin {
		t.Fatalf("expected quota to be at least %d when background is enabled, got %d", expectedMin, quota)
	}
}

func TestCalculateResponseAPIPreconsumeQuotaForeground(t *testing.T) {
	t.Parallel()

	maxOutput := 500
	inputRatio := 1.0
	outputRatio := inputRatio

	quota := calculateResponseAPIPreconsumeQuota(200, &maxOutput, inputRatio, outputRatio, false)

	expected := int64(float64(200+maxOutput) * inputRatio)
	if quota != expected {
		t.Fatalf("expected quota %d for foreground request, got %d", expected, quota)
	}
}

func TestCalculateResponseAPIPreconsumeQuotaBackgroundLargeEstimate(t *testing.T) {
	t.Parallel()

	maxOutput := 55000
	inputRatio := 1.0
	completionMultiplier := 0.5
	outputRatio := inputRatio * completionMultiplier

	quota := calculateResponseAPIPreconsumeQuota(100, &maxOutput, inputRatio, outputRatio, true)

	expectedBase := int64(float64(100+maxOutput) * inputRatio)
	expectedMin := int64(math.Ceil(float64(config.PreconsumeTokenForBackgroundRequest) * outputRatio))
	if quota != expectedBase {
		if expectedBase < expectedMin {
			t.Fatalf("expected quota to match background floor %d, got %d", expectedMin, quota)
		}
		t.Fatalf("expected quota to remain base estimate %d when it exceeds background floor %d, got %d", expectedBase, expectedMin, quota)
	}
}
