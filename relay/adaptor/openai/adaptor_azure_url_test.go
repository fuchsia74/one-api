package openai

import (
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

func TestAzureGetRequestURLRequiresModel(t *testing.T) {
	m := &meta.Meta{
		Mode:            relaymode.ChatCompletions,
		ChannelType:     channeltype.Azure,
		BaseURL:         "https://example.azure.com",
		ActualModelName: "",
		RequestURLPath:  "/v1/chat/completions",
		Config:          model.ChannelConfig{APIVersion: "2024-06-01"},
	}

	a := &Adaptor{}
	a.Init(m)

	if _, err := a.GetRequestURL(m); err == nil {
		t.Fatalf("expected error when ActualModelName is empty for Azure, got nil")
	}

	m.ActualModelName = "gpt-4o-mini"
	url, err := a.GetRequestURL(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(url, "/openai/deployments/gpt-4o-mini/chat/completions?api-version=") {
		t.Fatalf("unexpected azure request url: %s", url)
	}
}
