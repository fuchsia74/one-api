package controller

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
	metalib "github.com/songquanpeng/one-api/relay/meta"
)

func TestSupportsNativeResponseAPIOpenAICompatible(t *testing.T) {
	t.Parallel()

	metaInfo := &metalib.Meta{
		ChannelType: channeltype.OpenAICompatible,
		Config:      model.ChannelConfig{APIFormat: channeltype.OpenAICompatibleAPIFormatResponse},
	}
	require.True(t, supportsNativeResponseAPI(metaInfo))

	metaInfo.Config.APIFormat = channeltype.OpenAICompatibleAPIFormatChatCompletion
	require.False(t, supportsNativeResponseAPI(metaInfo))
}

func TestSupportsNativeResponseAPIDeepSeekForcesFallback(t *testing.T) {
	t.Parallel()

	metaInfo := &metalib.Meta{
		ChannelType:     channeltype.OpenAICompatible,
		Config:          model.ChannelConfig{APIFormat: channeltype.OpenAICompatibleAPIFormatResponse},
		ActualModelName: "deepseek-chat",
	}
	require.False(t, supportsNativeResponseAPI(metaInfo))

	metaInfo.ActualModelName = ""
	metaInfo.OriginModelName = "DeepSeek-Coder"
	require.False(t, supportsNativeResponseAPI(metaInfo))
}
