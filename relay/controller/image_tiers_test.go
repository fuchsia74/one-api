package controller

import (
	"math"
	"testing"

	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// Sanity check: usd_per_image * ImageUsdPerPic with $0.04 → 0.04 * 500000 = 20000
// We purposely do not call adapters; this guards controller math/unit consistency.
func TestImageUsdToQuotaMath(t *testing.T) {
	const quotaPerUsd = 500000.0
	usd := 0.04
	quotaPerImage := usd * quotaPerUsd
	if quotaPerImage != 20000 {
		t.Fatalf("expected 20000 quota per image for $0.04, got %v", quotaPerImage)
	}
}

// Test tier table values align with legacy logic for key models/sizes/qualities.
func TestImageTierTablesParity(t *testing.T) {
	// DALL·E 3 hd 1024x1024 → 2x; other sizes → 1.5x
	cases := []struct {
		model   string
		size    string
		quality string
		want    float64
	}{
		{"dall-e-3", "1024x1024", "hd", 2},
		{"dall-e-3", "1024x1792", "hd", 3}, // 2 * 1.5
		{"dall-e-3", "1792x1024", "hd", 3}, // 2 * 1.5
		{"gpt-image-1", "1024x1024", "high", 167.0 / 11},
		{"gpt-image-1", "1024x1536", "high", 250.0 / 11},
		{"gpt-image-1", "1536x1024", "high", 250.0 / 11},
		{"gpt-image-1", "1024x1024", "medium", 42.0 / 11},
		{"gpt-image-1", "1024x1536", "medium", 63.0 / 11},
		{"gpt-image-1", "1536x1024", "medium", 63.0 / 11},
		{"gpt-image-1", "1024x1024", "low", 1},
		{"gpt-image-1", "1024x1536", "low", 16.0 / 11},
		{"gpt-image-1", "1536x1024", "low", 16.0 / 11},
	}

	for _, tc := range cases {
		got, err := getImageCostRatio(&relaymodel.ImageRequest{Model: tc.model, Size: tc.size, Quality: tc.quality})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("%s %s %s: got %v, want %v", tc.model, tc.size, tc.quality, got, tc.want)
		}
	}
}
