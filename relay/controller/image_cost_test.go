package controller

import (
	"testing"

	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func TestGetImageCostRatio_Dalle3Tiers(t *testing.T) {
	// standard 1024x1024 -> 1x
	r := &relaymodel.ImageRequest{Model: "dall-e-3", Size: "1024x1024", Quality: "standard"}
	v, err := getImageCostRatio(r)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if v != 1 {
		t.Fatalf("expected 1, got %v", v)
	}

	// standard 1024x1792 -> 2x
	r = &relaymodel.ImageRequest{Model: "dall-e-3", Size: "1024x1792", Quality: "standard"}
	v, err = getImageCostRatio(r)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if v != 2 {
		t.Fatalf("expected 2, got %v", v)
	}

	// hd 1024x1024 -> 2x
	r = &relaymodel.ImageRequest{Model: "dall-e-3", Size: "1024x1024", Quality: "hd"}
	v, err = getImageCostRatio(r)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if v != 2 {
		t.Fatalf("expected 2, got %v", v)
	}

	// hd 1024x1792 -> 3x
	r = &relaymodel.ImageRequest{Model: "dall-e-3", Size: "1024x1792", Quality: "hd"}
	v, err = getImageCostRatio(r)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if v != 3 {
		t.Fatalf("expected 3, got %v", v)
	}
}
