package streamfinalizer

import (
	"encoding/json"
	"testing"

	"github.com/Laisky/zap"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"

	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

type capturedRender struct {
	payloads [][]byte
	allow    bool
}

func (c *capturedRender) render(b []byte) bool {
	if !c.allow {
		return false
	}
	cp := append([]byte(nil), b...)
	c.payloads = append(c.payloads, cp)
	return true
}

func TestFinalizerEmitsAfterStopAndMetadata(t *testing.T) {
	usage := relaymodel.Usage{}
	cap := &capturedRender{allow: true}
	f := NewFinalizer("test-model", 123, &usage, zap.NewNop(), cap.render)
	f.SetID("chatcmpl-1")

	stop := "stop"
	if !f.RecordStop(&stop) {
		t.Fatalf("record stop returned false")
	}
	if len(cap.payloads) != 0 {
		t.Fatalf("expected no emission before metadata, got %d", len(cap.payloads))
	}

	meta := &types.TokenUsage{
		InputTokens:  aws.Int32(10),
		OutputTokens: aws.Int32(20),
		TotalTokens:  aws.Int32(30),
	}
	if !f.RecordMetadata(meta) {
		t.Fatalf("record metadata returned false")
	}
	if len(cap.payloads) != 1 {
		t.Fatalf("expected one final chunk, got %d", len(cap.payloads))
	}

	var payload struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Choices []struct {
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(cap.payloads[0], &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Usage.PromptTokens != 10 || payload.Usage.CompletionTokens != 20 || payload.Usage.TotalTokens != 30 {
		t.Fatalf("unexpected usage: %+v", payload.Usage)
	}
	if len(payload.Choices) != 1 || payload.Choices[0].FinishReason == nil || *payload.Choices[0].FinishReason != "stop" {
		t.Fatalf("unexpected finish reason: %+v", payload.Choices)
	}
}

func TestFinalizerMetadataBeforeStop(t *testing.T) {
	usage := relaymodel.Usage{}
	cap := &capturedRender{allow: true}
	f := NewFinalizer("test-model", 123, &usage, zap.NewNop(), cap.render)
	f.SetID("chatcmpl-2")

	meta := &types.TokenUsage{}
	if !f.RecordMetadata(meta) {
		t.Fatalf("record metadata returned false")
	}
	if len(cap.payloads) != 0 {
		t.Fatalf("expected no chunk before stop, got %d", len(cap.payloads))
	}

	reason := "length"
	if !f.RecordStop(&reason) {
		t.Fatalf("record stop returned false")
	}
	if len(cap.payloads) != 1 {
		t.Fatalf("expected final chunk after stop, got %d", len(cap.payloads))
	}
}

func TestFinalizerFinalizeOnCloseWithoutMetadata(t *testing.T) {
	usage := relaymodel.Usage{}
	cap := &capturedRender{allow: true}
	f := NewFinalizer("test-model", 123, &usage, zap.NewNop(), cap.render)
	f.SetID("chatcmpl-3")

	reason := "stop"
	if !f.RecordStop(&reason) {
		t.Fatalf("record stop returned false")
	}
	if len(cap.payloads) != 0 {
		t.Fatalf("expected no chunk until close, got %d", len(cap.payloads))
	}

	if !f.FinalizeOnClose() {
		t.Fatalf("finalize on close returned false")
	}
	if len(cap.payloads) != 1 {
		t.Fatalf("expected one chunk on close, got %d", len(cap.payloads))
	}

	var payload struct {
		Usage *struct{} `json:"usage"`
	}
	if err := json.Unmarshal(cap.payloads[0], &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if payload.Usage != nil {
		t.Fatalf("expected no usage when metadata missing")
	}
}

func TestFinalizerFinalizeWithoutStop(t *testing.T) {
	usage := relaymodel.Usage{}
	cap := &capturedRender{allow: true}
	f := NewFinalizer("test-model", 123, &usage, zap.NewNop(), cap.render)
	f.SetID("chatcmpl-4")

	if !f.FinalizeOnClose() {
		t.Fatalf("finalize on close returned false")
	}
	if len(cap.payloads) != 1 {
		t.Fatalf("expected chunk even without stop, got %d", len(cap.payloads))
	}

	var payload struct {
		Choices []struct {
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(cap.payloads[0], &payload); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(payload.Choices) != 1 || payload.Choices[0].FinishReason != nil {
		t.Fatalf("expected nil finish reason, got %+v", payload.Choices)
	}
}

func TestFinalizerNotDuplicate(t *testing.T) {
	usage := relaymodel.Usage{}
	cap := &capturedRender{allow: true}
	f := NewFinalizer("test-model", 123, &usage, zap.NewNop(), cap.render)
	f.SetID("chatcmpl-5")

	reason := "stop"
	f.RecordStop(&reason)
	meta := &types.TokenUsage{}
	f.RecordMetadata(meta)
	if len(cap.payloads) != 1 {
		t.Fatalf("expected one chunk after first emit, got %d", len(cap.payloads))
	}

	if !f.FinalizeOnClose() {
		t.Fatalf("finalize on close returned false")
	}
	if len(cap.payloads) != 1 {
		t.Fatalf("expected no additional chunks, got %d", len(cap.payloads))
	}
}
