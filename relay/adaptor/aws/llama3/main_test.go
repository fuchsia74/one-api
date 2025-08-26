package aws_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	aws "github.com/songquanpeng/one-api/relay/adaptor/aws/llama3"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func TestRenderPrompt(t *testing.T) {
	messages := []relaymodel.Message{
		{
			Role:    "user",
			Content: "What's your name?",
		},
	}
	// Test with legacy model (Llama 3) - should use <|eot_id|>
	prompt := aws.RenderPrompt(messages, "llama3-8b-8192")
	expected := `<|begin_of_text|><|start_header_id|>user<|end_header_id|>What's your name?<|eot_id|><|start_header_id|>assistant<|end_header_id|>`
	require.Equal(t, expected, prompt)

	// Test with Llama 4 model - should use <|eot|>
	prompt = aws.RenderPrompt(messages, "llama4-scout-17b-3.5m")
	expected = `<|begin_of_text|><|start_header_id|>user<|end_header_id|>What's your name?<|eot|><|start_header_id|>assistant<|end_header_id|>`
	require.Equal(t, expected, prompt)

	messages = []relaymodel.Message{
		{
			Role:    "system",
			Content: "Your name is Kat. You are a detective.",
		},
		{
			Role:    "user",
			Content: "What's your name?",
		},
		{
			Role:    "assistant",
			Content: "Kat",
		},
		{
			Role:    "user",
			Content: "What's your job?",
		},
	}
	// Test multi-message with legacy model
	prompt = aws.RenderPrompt(messages, "llama3-70b-8192")
	expected = `<|begin_of_text|><|start_header_id|>system<|end_header_id|>Your name is Kat. You are a detective.<|eot_id|><|start_header_id|>user<|end_header_id|>What's your name?<|eot_id|><|start_header_id|>assistant<|end_header_id|>Kat<|eot_id|><|start_header_id|>user<|end_header_id|>What's your job?<|eot_id|><|start_header_id|>assistant<|end_header_id|>`
	require.Equal(t, expected, prompt)

	// Test multi-message with Llama 4 model
	prompt = aws.RenderPrompt(messages, "llama4-maverick-17b-1m")
	expected = `<|begin_of_text|><|start_header_id|>system<|end_header_id|>Your name is Kat. You are a detective.<|eot|><|start_header_id|>user<|end_header_id|>What's your name?<|eot|><|start_header_id|>assistant<|end_header_id|>Kat<|eot|><|start_header_id|>user<|end_header_id|>What's your job?<|eot|><|start_header_id|>assistant<|end_header_id|>`
	require.Equal(t, expected, prompt)
}
