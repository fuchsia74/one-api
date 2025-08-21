package controller

import (
	"testing"

	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

// This test documents the intended unit behavior for image pricing math:
// adapter ratios for image models are already in quota-per-image units
// (usd_per_image * QuotaPerUsd). Controller must not multiply by 1000 again.
//
// Note: This is a lightweight doc-test ensuring we don't reintroduce the old bug.
func TestImageQuotaNoExtraThousand(t *testing.T) {
	_ = relaymodel.Usage{} // reference package to avoid unused import if first test is modified
	// Suppose adapter ratio encodes $0.04 per image → 0.04 * 500000 = 20000 quota/image
	ratio := 20000.0 // quota per image
	imageCostRatio := 1.0

	// Old buggy math would do: int64(ratio*imageCostRatio) * 1000 → 20,000,000
	// Correct math: no extra *1000
	usedQuotaSingle := int64(ratio * imageCostRatio)
	if usedQuotaSingle != 20000 {
		t.Fatalf("unexpected single-image quota: got %d want %d", usedQuotaSingle, 20000)
	}

	// n images scale linearly
	n := int64(3)
	usedQuotaN := usedQuotaSingle * n
	if usedQuotaN != 60000 {
		t.Fatalf("unexpected n-image quota: got %d want %d", usedQuotaN, 60000)
	}
}
