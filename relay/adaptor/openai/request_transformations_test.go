package openai

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
	relaymeta "github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func float64PtrRT(v float64) *float64 {
	return &v
}

func stringPtrRT(s string) *string {
	return &s
}

func TestApplyRequestTransformations_ReasoningDefaults(t *testing.T) {
	adaptor := &Adaptor{}

	cases := []struct {
		name          string
		channelType   int
		expectNilTemp bool
	}{
		{name: "OpenAI Responses", channelType: channeltype.OpenAI, expectNilTemp: true},
		{name: "Azure Chat", channelType: channeltype.Azure, expectNilTemp: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			meta := &relaymeta.Meta{
				ChannelType:     tc.channelType,
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

			if tc.expectNilTemp {
				if req.Temperature != nil {
					t.Fatalf("expected Temperature to be removed, got %v", req.Temperature)
				}
			} else {
				if req.Temperature == nil || *req.Temperature != 1 {
					t.Fatalf("expected Temperature to be forced to 1, got %v", req.Temperature)
				}
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
		})
	}
}

func TestApplyRequestTransformations_DeepResearchAddsWebSearchTool(t *testing.T) {
	adaptor := &Adaptor{}

	meta := &relaymeta.Meta{
		ChannelType:     channeltype.OpenAI,
		ActualModelName: "o3-deep-research",
	}

	req := &model.GeneralOpenAIRequest{
		Model: "o3-deep-research",
		Messages: []model.Message{
			{Role: "user", Content: "summarize the news"},
		},
	}

	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("applyRequestTransformations returned error: %v", err)
	}

	count := 0
	for _, tool := range req.Tools {
		if tool.Type == "web_search" {
			count++
		}
	}

	if count != 1 {
		t.Fatalf("expected exactly one web_search tool after transformation, got %d", count)
	}

	// Running transformations again should not duplicate the tool
	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("second applyRequestTransformations returned error: %v", err)
	}

	count = 0
	for _, tool := range req.Tools {
		if tool.Type == "web_search" {
			count++
		}
	}

	if count != 1 {
		t.Fatalf("expected web_search tool count to remain 1 after second pass, got %d", count)
	}
}

func TestApplyRequestTransformations_WebSearchOptionsAddsWebSearchTool(t *testing.T) {
	adaptor := &Adaptor{}

	meta := &relaymeta.Meta{
		ChannelType:     channeltype.OpenAI,
		ActualModelName: "gpt-4o-search-preview",
	}

	req := &model.GeneralOpenAIRequest{
		Model:            "gpt-4o-search-preview",
		WebSearchOptions: &model.WebSearchOptions{},
	}

	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("applyRequestTransformations returned error: %v", err)
	}

	count := 0
	for _, tool := range req.Tools {
		if tool.Type == "web_search" {
			count++
		}
	}

	if count != 1 {
		t.Fatalf("expected exactly one web_search tool when web_search_options provided, got %d", count)
	}

	// Running the transformation again should not duplicate the tool
	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("second applyRequestTransformations returned error: %v", err)
	}

	count = 0
	for _, tool := range req.Tools {
		if tool.Type == "web_search" {
			count++
		}
	}

	if count != 1 {
		t.Fatalf("expected web_search tool count to remain 1 after second pass, got %d", count)
	}
}

func TestApplyRequestTransformations_DeepResearchReasoningEffort(t *testing.T) {
	adaptor := &Adaptor{}

	meta := &relaymeta.Meta{
		ChannelType:     channeltype.OpenAI,
		ActualModelName: "o4-mini-deep-research",
	}

	req := &model.GeneralOpenAIRequest{
		Model: "o4-mini-deep-research",
		Messages: []model.Message{
			{Role: "user", Content: "Summarize the latest research on fusion"},
		},
	}

	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("applyRequestTransformations returned error: %v", err)
	}

	if req.ReasoningEffort == nil || *req.ReasoningEffort != "medium" {
		t.Fatalf("expected ReasoningEffort to default to 'medium', got %v", req.ReasoningEffort)
	}

	// User-provided unsupported effort should be normalized to medium
	req = &model.GeneralOpenAIRequest{
		Model:           "o4-mini-deep-research",
		ReasoningEffort: stringPtrRT("high"),
		Messages:        []model.Message{{Role: "user", Content: "analyze"}},
	}

	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("applyRequestTransformations returned error: %v", err)
	}

	if req.ReasoningEffort == nil || *req.ReasoningEffort != "medium" {
		t.Fatalf("expected ReasoningEffort to be normalized to 'medium', got %v", req.ReasoningEffort)
	}
}

func TestApplyRequestTransformations_ResponseAPIRemovesSampling(t *testing.T) {
	adaptor := &Adaptor{}

	meta := &relaymeta.Meta{
		ChannelType:     channeltype.OpenAI,
		ActualModelName: "gpt-5-mini",
		Mode:            relaymode.ResponseAPI,
	}

	req := &model.GeneralOpenAIRequest{
		Model: "gpt-5-mini",
		Messages: []model.Message{
			{Role: "user", Content: "hello"},
		},
		Temperature: float64PtrRT(0.3),
		TopP:        float64PtrRT(0.2),
	}

	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("applyRequestTransformations returned error: %v", err)
	}

	if req.Temperature != nil {
		t.Fatalf("expected Temperature to be removed for Response API reasoning models, got %v", req.Temperature)
	}

	if req.TopP != nil {
		t.Fatalf("expected TopP to be removed for Response API reasoning models, got %v", req.TopP)
	}
}

func TestApplyRequestTransformations_ValidDataURLImage(t *testing.T) {
	adaptor := &Adaptor{}

	meta := &relaymeta.Meta{
		ChannelType:     channeltype.OpenAI,
		ActualModelName: "gpt-5-codex",
	}

	dataURL := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

	req := &model.GeneralOpenAIRequest{
		Model: "gpt-5-codex",
		Messages: []model.Message{
			{
				Role: "user",
				Content: []model.MessageContent{
					{
						Type: model.ContentTypeText,
						Text: stringPtrRT("Describe the image"),
					},
					{
						Type:     model.ContentTypeImageURL,
						ImageURL: &model.ImageURL{Url: dataURL},
					},
				},
			},
		},
	}

	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("applyRequestTransformations returned error for valid data URL: %v", err)
	}
}

func TestApplyRequestTransformations_InvalidDataURLImage(t *testing.T) {
	adaptor := &Adaptor{}

	meta := &relaymeta.Meta{
		ChannelType:     channeltype.OpenAI,
		ActualModelName: "gpt-5-codex",
	}

	req := &model.GeneralOpenAIRequest{
		Model: "gpt-5-codex",
		Messages: []model.Message{
			{
				Role: "user",
				Content: []model.MessageContent{
					{
						Type: model.ContentTypeText,
						Text: stringPtrRT("Describe the image"),
					},
					{
						Type:     model.ContentTypeImageURL,
						ImageURL: &model.ImageURL{Url: "data:image/png;base64,not-an-image"},
					},
				},
			},
		},
	}

	if err := adaptor.applyRequestTransformations(meta, req); err == nil {
		t.Fatalf("expected error for invalid data URL image, got nil")
	}
}

func TestApplyRequestTransformations_PopulatesMetaActualModel(t *testing.T) {
	adaptor := &Adaptor{}

	meta := &relaymeta.Meta{
		ChannelType:  channeltype.OpenAI,
		ModelMapping: map[string]string{"gpt-x": "gpt-x-mapped"},
	}

	req := &model.GeneralOpenAIRequest{
		Model: "gpt-x",
		Messages: []model.Message{
			{Role: "user", Content: "hi"},
		},
	}

	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("applyRequestTransformations returned error: %v", err)
	}

	if meta.OriginModelName != "gpt-x" {
		t.Fatalf("expected OriginModelName to be populated, got %q", meta.OriginModelName)
	}

	if meta.ActualModelName != "gpt-x-mapped" {
		t.Fatalf("expected ActualModelName to use mapping, got %q", meta.ActualModelName)
	}
}

func TestApplyRequestTransformations_NormalizesToolChoice(t *testing.T) {
	adaptor := &Adaptor{}

	meta := &relaymeta.Meta{
		ChannelType:     channeltype.OpenAI,
		ActualModelName: "gpt-4o-mini",
	}

	req := &model.GeneralOpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []model.Message{
			{Role: "user", Content: "Call the weather tool"},
		},
		ToolChoice: map[string]any{
			"type": "tool",
			"name": "get_weather",
		},
	}

	if err := adaptor.applyRequestTransformations(meta, req); err != nil {
		t.Fatalf("applyRequestTransformations returned error: %v", err)
	}

	toolChoice, ok := req.ToolChoice.(map[string]any)
	if !ok {
		t.Fatalf("expected tool_choice to be map after normalization, got %T", req.ToolChoice)
	}

	if typ := toolChoice["type"]; typ != "function" {
		t.Fatalf("expected normalized tool_choice type 'function', got %v", typ)
	}

	if name := toolChoice["name"]; name != "get_weather" {
		t.Fatalf("expected top-level name 'get_weather', got %v", name)
	}

	if _, exists := toolChoice["function"]; exists {
		t.Fatalf("function block should be stripped for OpenAI upstream requests")
	}
}
