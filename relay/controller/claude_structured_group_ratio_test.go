package controller

import (
	"math"
	"testing"

	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func TestClaudeStructuredOutputCost_ScalesWithGroupRatio(t *testing.T) {
	completionTokens := 1000
	modelRatio := 0.25

	req := &ClaudeMessagesRequest{
		Tools:     []relaymodel.ClaudeTool{{Name: "t", Description: "d"}},
		Model:     "claude-3-haiku-20240307",
		MaxTokens: 128,
	}

	// groupRatio=1
	costBase := calculateClaudeStructuredOutputCost(req, completionTokens, modelRatio, 1.0)
	// groupRatio=2 should roughly double the surcharge
	costDouble := calculateClaudeStructuredOutputCost(req, completionTokens, modelRatio, 2.0)

	// Expected values
	expBase := int64(math.Ceil(float64(completionTokens) * 0.25 * (modelRatio * 1.0)))
	expDouble := int64(math.Ceil(float64(completionTokens) * 0.25 * (modelRatio * 2.0)))

	if costBase != expBase {
		t.Fatalf("base group ratio cost mismatch: got %d want %d", costBase, expBase)
	}
	if costDouble != expDouble {
		t.Fatalf("double group ratio cost mismatch: got %d want %d", costDouble, expDouble)
	}
	if costDouble < costBase*2-1 || costDouble > costBase*2+1 {
		t.Fatalf("cost does not scale ~2x with group ratio: base=%d double=%d", costBase, costDouble)
	}
}
