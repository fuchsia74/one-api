package openai

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
	relaymeta "github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

func float64PtrRT(v float64) *float64 {
	return &v
}

func TestApplyRequestTransformations_ReasoningDefaults(t *testing.T) {
	adaptor := &Adaptor{}

	meta := &relaymeta.Meta{
		ChannelType:     channeltype.OpenAI,
		ActualModelName: "o1-preview",
	}

	req := &model.GeneralOpenAIRequest{
		Model:     "o1-preview",
		MaxTokens: 1500,
		Messages: []model.Message{
			{Role: "system", Content: "be precise"},
			{Role: "user", Content: "hi"},
		},
		Temperature: float64PtrRT(0.5),
		TopP:        float64PtrRT(0.9),
	}

	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("applyRequestTransformations returned error: %v", err)
	}

	if req.MaxTokens != 0 {
		t.Errorf("expected MaxTokens to be zeroed, got %d", req.MaxTokens)
	}

	if req.MaxCompletionTokens == nil || *req.MaxCompletionTokens != 1500 {
		t.Fatalf("expected MaxCompletionTokens to be set to 1500, got %v", req.MaxCompletionTokens)
	}

	if req.Temperature == nil || *req.Temperature != 1 {
		t.Fatalf("expected Temperature to be forced to 1 for reasoning models, got %v", req.Temperature)
	}

	if req.TopP != nil {
		t.Fatalf("expected TopP to be cleared for reasoning models, got %v", *req.TopP)
	}

	if req.ReasoningEffort == nil || *req.ReasoningEffort != "high" {
		t.Fatalf("expected ReasoningEffort to default to 'high', got %v", req.ReasoningEffort)
	}

	if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
		t.Fatalf("expected system messages to be stripped for reasoning models, got %+v", req.Messages)
	}
}
