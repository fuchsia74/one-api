package openai

import "testing"

func TestDallE3HasPerImagePricing(t *testing.T) {
	cfg, ok := ModelRatios["dall-e-3"]
	if !ok {
		t.Fatalf("dall-e-3 not found in ModelRatios")
	}
	if cfg.Ratio != 0 {
		t.Fatalf("expected Ratio=0 for per-image model, got %v", cfg.Ratio)
	}
	if cfg.ImagePriceUsd <= 0 {
		t.Fatalf("expected ImagePriceUsd > 0 for dall-e-3, got %v", cfg.ImagePriceUsd)
	}
}
