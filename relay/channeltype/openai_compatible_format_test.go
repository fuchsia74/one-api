package channeltype

import "testing"

func TestNormalizeOpenAICompatibleAPIFormat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty defaults to chat", "", OpenAICompatibleAPIFormatChatCompletion},
		{"whitespace defaults", "   ", OpenAICompatibleAPIFormatChatCompletion},
		{"chat-completion alias", "Chat-Completion", OpenAICompatibleAPIFormatChatCompletion},
		{"chat shorthand", "chat", OpenAICompatibleAPIFormatChatCompletion},
		{"response canonical", "response", OpenAICompatibleAPIFormatResponse},
		{"response alias", "Response_API", OpenAICompatibleAPIFormatResponse},
		{"response plural", "responses", OpenAICompatibleAPIFormatResponse},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := NormalizeOpenAICompatibleAPIFormat(tc.input)
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestUseOpenAICompatibleResponseAPI(t *testing.T) {
	t.Parallel()

	if UseOpenAICompatibleResponseAPI("response") != true {
		t.Fatal("expected response format to enable Response API")
	}

	if UseOpenAICompatibleResponseAPI("chat") {
		t.Fatal("expected chat format to disable Response API")
	}
}
