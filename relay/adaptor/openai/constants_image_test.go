package openai

import (
	"testing"
)

func TestGptImage1HasPerImagePricing(t *testing.T) {
	cfg, ok := ModelRatios["gpt-image-1"]
	if !ok {
		t.Fatalf("gpt-image-1 not found in ModelRatios")
	}
	if cfg.Ratio != 0 {
		t.Fatalf("expected Ratio=0 for per-image model, got %v", cfg.Ratio)
	}
	if cfg.ImagePriceUsd <= 0 {
		t.Fatalf("expected ImagePriceUsd > 0 for gpt-image-1, got %v", cfg.ImagePriceUsd)
	}
}
