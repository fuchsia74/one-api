package controller

import (
	"testing"

	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func TestClaudeStructuredOutputCost_NoSurcharge(t *testing.T) {
	completionTokens := 1000
	modelRatio := 0.25

	req := &ClaudeMessagesRequest{
		Tools:     []relaymodel.ClaudeTool{{Name: "t", Description: "d"}},
		Model:     "claude-3-haiku-20240307",
		MaxTokens: 128,
	}

	costBase := calculateClaudeStructuredOutputCost(req, completionTokens, modelRatio, 1.0)
	costDouble := calculateClaudeStructuredOutputCost(req, completionTokens, modelRatio, 2.0)

	if costBase != 0 || costDouble != 0 {
		t.Fatalf("expected no structured output surcharge, got base=%d double=%d", costBase, costDouble)
	}
}
