package aws

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/gin-gonic/gin"

	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func TestConvertConverseResponseToQwenUsageMapping(t *testing.T) {
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

	resp := convertConverseResponseToQwen(c, converse, "qwen.qwen3-coder-480b-a35b-v1:0")

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal qwen response: %v", err)
	}

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

func TestConvertRequestMapsToolsAndReasoning(t *testing.T) {
	temp := 0.5
	topP := 0.9
	reasoning := "high"
	stop := []any{"done", "halt"}

	req := relaymodel.GeneralOpenAIRequest{
		Messages: []relaymodel.Message{
			{Role: "system", Content: "you are helpful"},
			{Role: "user", Content: "hi"},
		},
		Model:           "qwen3-32b",
		MaxTokens:       4096,
		Temperature:     &temp,
		TopP:            &topP,
		Stop:            stop,
		ReasoningEffort: &reasoning,
		Tools: []relaymodel.Tool{
			{
				Type: "function",
				Function: &relaymodel.Function{
					Name:        "calculate",
					Description: "perform calculation",
					Parameters: map[string]any{
						"type": "object",
					},
				},
			},
		},
		ToolChoice: map[string]any{
			"type": "function",
			"function": map[string]any{
				"name": "calculate",
			},
		},
	}

	converted := ConvertRequest(req)
	if converted == nil {
		t.Fatalf("expected non-nil converted request")
	}

	if converted.MaxTokens != 4096 {
		t.Fatalf("unexpected max tokens: %d", converted.MaxTokens)
	}

	if converted.Temperature == nil || *converted.Temperature != temp {
		t.Fatalf("temperature not preserved: %+v", converted.Temperature)
	}

	if converted.TopP == nil || *converted.TopP != topP {
		t.Fatalf("top_p not preserved: %+v", converted.TopP)
	}

	if !reflect.DeepEqual(converted.Stop, []string{"done", "halt"}) {
		t.Fatalf("unexpected stop sequences: %+v", converted.Stop)
	}

	if converted.ReasoningEffort == nil || *converted.ReasoningEffort != reasoning {
		t.Fatalf("reasoning effort not preserved: %+v", converted.ReasoningEffort)
	}

	if len(converted.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(converted.Tools))
	}

	if converted.Tools[0].Function.Name != "calculate" {
		t.Fatalf("unexpected tool name: %s", converted.Tools[0].Function.Name)
	}

	toolChoice, ok := converted.ToolChoice.(map[string]any)
	if !ok {
		t.Fatalf("expected tool choice map, got %T", converted.ToolChoice)
	}
	if toolChoice["type"] != "function" {
		t.Fatalf("unexpected tool choice type: %v", toolChoice["type"])
	}
	if fn, ok := toolChoice["function"].(map[string]any); !ok || fn["name"] != "calculate" {
		t.Fatalf("unexpected tool choice function: %+v", toolChoice["function"])
	}
}

func TestConvertMessagesMarshalsNonStringArguments(t *testing.T) {
	args := map[string]any{"foo": "bar"}
	messages := []relaymodel.Message{
		{
			Role: "assistant",
			ToolCalls: []relaymodel.Tool{
				{
					Id:   "call-1",
					Type: "function",
					Function: &relaymodel.Function{
						Name:      "do",
						Arguments: args,
					},
				},
			},
		},
	}

	converted := ConvertMessages(messages)
	if len(converted) != 1 {
		t.Fatalf("expected 1 message, got %d", len(converted))
	}

	if len(converted[0].ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(converted[0].ToolCalls))
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(converted[0].ToolCalls[0].Function.Arguments), &decoded); err != nil {
		t.Fatalf("arguments should be json: %v", err)
	}

	if decoded["foo"] != "bar" {
		t.Fatalf("unexpected tool arguments: %+v", decoded)
	}
}

func TestConvertQwenToConverseRequestIncludesReasoningConfig(t *testing.T) {
	reasoning := "medium"
	req := &Request{
		Messages: []Message{
			{Role: "system", Content: "guide"},
			{Role: "user", Content: "hello"},
		},
		ReasoningEffort: &reasoning,
		Tools: []QwenTool{
			{
				Type: "function",
				Function: QwenToolSpec{
					Name:        "calculate",
					Description: "math helper",
					Parameters: map[string]any{
						"type": "object",
					},
				},
			},
		},
		ToolChoice: map[string]any{
			"type": "function",
			"function": map[string]any{
				"name": "calculate",
			},
		},
	}

	converseReq, err := convertQwenToConverseRequest(req, "qwen3-test")
	if err != nil {
		t.Fatalf("convert request: %v", err)
	}

	if converseReq.ToolConfig == nil {
		t.Fatalf("expected tool config to be set")
	}
	if len(converseReq.ToolConfig.Tools) != 1 {
		t.Fatalf("expected one tool specification, got %d", len(converseReq.ToolConfig.Tools))
	}

	toolSpec, ok := converseReq.ToolConfig.Tools[0].(*types.ToolMemberToolSpec)
	if !ok {
		t.Fatalf("unexpected tool type: %T", converseReq.ToolConfig.Tools[0])
	}
	if toolSpec.Value.Name == nil || *toolSpec.Value.Name != "calculate" {
		t.Fatalf("unexpected tool name: %v", toolSpec.Value.Name)
	}

	if converseReq.AdditionalModelRequestFields == nil {
		t.Fatalf("expected reasoning config to be included")
	}

	b, err := json.Marshal(converseReq.AdditionalModelRequestFields)
	if err != nil {
		t.Fatalf("marshal additional fields: %v", err)
	}
	if len(b) == 0 {
		t.Fatalf("expected additional fields to marshal to json, got empty payload")
	}
}

func TestConvertQwenToConverseRequestInvalidToolArguments(t *testing.T) {
	req := &Request{
		Messages: []Message{
			{
				Role: "assistant",
				ToolCalls: []QwenToolCall{
					{
						ID:   "call-1",
						Type: "function",
						Function: QwenToolFunction{
							Name:      "calc",
							Arguments: "{not-json",
						},
					},
				},
			},
		},
	}

	_, err := convertQwenToConverseRequest(req, "qwen3-test")
	if err == nil {
		t.Fatalf("expected error converting invalid tool arguments")
	}
}
