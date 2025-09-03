package openai

import (
	"testing"

	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
)

func TestGetRequestURL_AzureRequiresModel(t *testing.T) {
	a := &Adaptor{}
	m := &meta.Meta{ChannelType: channeltype.Azure, BaseURL: "https://example.openai.azure.com", RequestURLPath: "/v1/chat/completions"}
	if _, err := a.GetRequestURL(m); err == nil {
		t.Fatalf("expected error when ActualModelName is empty for Azure, got nil")
	}

	m.ActualModelName = "gpt-4o-mini"
	if _, err := a.GetRequestURL(m); err != nil {
		t.Fatalf("unexpected error building Azure URL with model: %v", err)
	}
}
