package replicate

import "testing"

// Ensure key Replicate image models have non-zero per-image pricing
func TestReplicateImageModelPrices(t *testing.T) {
	cases := []string{
		"black-forest-labs/flux-schnell",
		"black-forest-labs/flux-pro",
		"stability-ai/stable-diffusion-3",
	}
	for _, model := range cases {
		cfg, ok := ModelRatios[model]
		if !ok {
			t.Fatalf("model %s not found in ModelRatios", model)
		}
		if cfg.ImagePriceUsd <= 0 {
			t.Fatalf("expected ImagePriceUsd > 0 for %s, got %v", model, cfg.ImagePriceUsd)
		}
	}
}
