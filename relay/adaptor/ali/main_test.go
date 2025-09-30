package ali

import (
	"math"
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

func float64PtrAli(v float64) *float64 {
	return &v
}

func TestConvertRequestClampsTopP(t *testing.T) {
	req := model.GeneralOpenAIRequest{
		Model: "qwen-plus-internet",
		TopP:  float64PtrAli(1.5),
	}

	converted := ConvertRequest(req)
	if converted.Parameters.TopP == nil {
		t.Fatal("expected TopP to be populated")
	}

	if diff := math.Abs(*converted.Parameters.TopP - 0.9999); diff > 1e-9 {
		t.Fatalf("expected TopP to be clamped to 0.9999, got %v", *converted.Parameters.TopP)
	}
}

func TestConvertRequestLeavesNilTopPUnchanged(t *testing.T) {
	req := model.GeneralOpenAIRequest{
		Model: "qwen-plus",
	}

	converted := ConvertRequest(req)
	if converted.Parameters.TopP != nil {
		t.Fatalf("expected TopP to remain nil when not provided, got %v", *converted.Parameters.TopP)
	}
}
