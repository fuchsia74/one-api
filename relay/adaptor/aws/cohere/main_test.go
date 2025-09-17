package aws

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/gin-gonic/gin"
)

// Test that convertConverseResponseToCohere maps usage to relaymodel.Usage
// with JSON fields: prompt_tokens, completion_tokens, total_tokens.
func TestConvertConverseResponseToCohereUsageMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	in := int32(11)
	out := int32(22)
	total := int32(33)

	converse := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{Value: types.Message{Role: types.ConversationRole("assistant")}},
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(in),
			OutputTokens: aws.Int32(out),
			TotalTokens:  aws.Int32(total),
		},
	}

	resp := convertConverseResponseToCohere(c, converse, "cohere.command-r-v1:0")

	// Marshal to JSON to assert field names
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal cohere response: %v", err)
	}

	// Quick JSON string contains checks for usage keys and values
	js := string(b)
	if want := "\"prompt_tokens\":11"; !strings.Contains(js, want) {
		t.Fatalf("expected %s in json, got %s", want, js)
	}
	if want := "\"completion_tokens\":22"; !strings.Contains(js, want) {
		t.Fatalf("expected %s in json, got %s", want, js)
	}
	if want := "\"total_tokens\":33"; !strings.Contains(js, want) {
		t.Fatalf("expected %s in json, got %s", want, js)
	}
}
