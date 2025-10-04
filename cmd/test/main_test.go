package main

import (
	"net/http"
	"testing"
)

func TestParseModels(t *testing.T) {
	cases := map[string][]string{
		"gpt-4":                     {"gpt-4"},
		"gpt-4,claude-3":            {"gpt-4", "claude-3"},
		"gpt-4; claude-3 \n gemini": {"gpt-4", "claude-3", "gemini"},
		"  gpt-4  ,  claude-3   ":   {"gpt-4", "claude-3"},
		"gpt-4\n\nclaude-3":         {"gpt-4", "claude-3"},
		"gpt-4 claude-3":            {"gpt-4", "claude-3"},
	}

	for input, want := range cases {
		got, err := parseModels(input)
		if err != nil {
			t.Fatalf("parseModels(%q) returned error: %v", input, err)
		}
		if len(got) != len(want) {
			t.Fatalf("parseModels(%q) length = %d, want %d", input, len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("parseModels(%q)[%d] = %q, want %q", input, i, got[i], want[i])
			}
		}
	}
}

func TestParseModelsEmpty(t *testing.T) {
	got, err := parseModels("   ")
	if err != nil {
		t.Fatalf("parseModels empty error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("parseModels empty length = %d, want 0", len(got))
	}
}

func TestEvaluateResponseChatCompletionSuccess(t *testing.T) {
	body := []byte(`{"choices":[{"message":{"role":"assistant","content":"hello"}}]}`)
	success, reason := evaluateResponse(requestTypeChatCompletion, body)
	if !success {
		t.Fatalf("expected success, got failure: %s", reason)
	}
}

func TestEvaluateResponseIgnoresEmptyErrorObject(t *testing.T) {
	body := []byte(`{"choices":[{"message":{"role":"assistant","content":"hi"}}],"error":{"message":"","type":"","param":"","code":null}}`)
	success, reason := evaluateResponse(requestTypeChatCompletion, body)
	if !success {
		t.Fatalf("expected success despite empty error object, got: %s", reason)
	}
}

func TestEvaluateResponseResponseAPIChoicesFallback(t *testing.T) {
	body := []byte(`{"choices":[{"message":{"role":"assistant","content":"hi"}}],"object":"chat.completion"}`)
	success, reason := evaluateResponse(requestTypeResponseAPI, body)
	if !success {
		t.Fatalf("expected Response API fallback success, got: %s", reason)
	}
}

func TestIsUnsupportedCombinationResponse(t *testing.T) {
	body := []byte("{\"error\":{\"message\":\"unknown field `messages`\"}}")
	if !isUnsupportedCombination(requestTypeResponseAPI, false, http.StatusBadRequest, body, "") {
		t.Fatalf("expected combination to be marked unsupported")
	}
}

func TestEvaluateStreamResponseSuccess(t *testing.T) {
	data := []byte("data: {\"id\":\"resp_123\",\"error\":null}\n\n")
	success, reason := evaluateStreamResponse(requestTypeResponseAPI, data)
	if !success {
		t.Fatalf("expected stream success, got failure: %s", reason)
	}
}

func TestIsUnsupportedCombinationStream(t *testing.T) {
	body := []byte("streaming is not supported")
	if !isUnsupportedCombination(requestTypeChatCompletion, true, http.StatusBadRequest, body, "") {
		t.Fatalf("expected streaming combination to be marked unsupported")
	}
}
