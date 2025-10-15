package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

// evaluateResponse inspects a non-streaming response and validates the expected shape.
func evaluateResponse(spec requestSpec, body []byte) (bool, string) {
	if len(body) == 0 {
		return true, ""
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return false, fmt.Sprintf("malformed JSON response: %v", err)
	}

	if errVal, ok := payload["error"]; ok && isMeaningfulErrorValue(errVal) {
		return false, snippet(body)
	}

	switch spec.Type {
	case requestTypeChatCompletion:
		switch spec.Expectation {
		case expectationToolInvocation:
			if choices, ok := payload["choices"].([]any); ok {
				for _, choice := range choices {
					choiceMap, ok := choice.(map[string]any)
					if !ok {
						continue
					}
					if message, ok := choiceMap["message"].(map[string]any); ok {
						if calls, ok := message["tool_calls"].([]any); ok && len(calls) > 0 {
							return true, ""
						}
					}
				}
			}
			return false, "response missing tool_calls"
		default:
			if choices, ok := payload["choices"].([]any); ok && len(choices) > 0 {
				return true, ""
			}
			return false, "response missing choices"
		}
	case requestTypeResponseAPI:
		switch spec.Expectation {
		case expectationToolInvocation:
			if required, ok := payload["required_action"].(map[string]any); ok {
				if stringValue(required, "type") == "submit_tool_outputs" {
					if submit, ok := required["submit_tool_outputs"].(map[string]any); ok {
						if calls, ok := submit["tool_calls"].([]any); ok && len(calls) > 0 {
							return true, ""
						}
					}
				}
			}
			if hasFunctionCallOutput(payload) {
				return true, ""
			}
			return false, "response missing required_action.tool_calls"
		default:
			status := stringValue(payload, "status")
			if status == "failed" {
				return false, snippet(body)
			}
			if output, ok := payload["output"].([]any); ok && len(output) > 0 {
				return true, ""
			}
			if choices, ok := payload["choices"].([]any); ok && len(choices) > 0 {
				return true, ""
			}
			if status == "completed" || status == "in_progress" || status == "requires_action" {
				return true, ""
			}
			if len(payload) == 0 {
				return false, "empty response"
			}
			return false, "response missing output"
		}
	case requestTypeClaudeMessages:
		switch spec.Expectation {
		case expectationToolInvocation:
			if content, ok := payload["content"].([]any); ok {
				for _, entry := range content {
					entryMap, ok := entry.(map[string]any)
					if !ok {
						continue
					}
					if stringValue(entryMap, "type") == "tool_use" {
						return true, ""
					}
				}
			}
			return false, "response missing tool_use block"
		default:
			if content, ok := payload["content"].([]any); ok && len(content) > 0 {
				return true, ""
			}
			if msgType := stringValue(payload, "type"); msgType != "" {
				return true, ""
			}
			if len(payload) == 0 {
				return false, "empty response"
			}
			return true, ""
		}
	default:
		return true, ""
	}
}

// hasFunctionCallOutput reports whether the Response API payload contains a function_call entry.
func hasFunctionCallOutput(payload map[string]any) bool {
	output, ok := payload["output"].([]any)
	if !ok {
		return false
	}
	for _, entry := range output {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if stringValue(entryMap, "type") == "function_call" {
			return true
		}
	}
	return false
}

// evaluateStreamResponse validates streaming SSE payloads for expected content.
func evaluateStreamResponse(spec requestSpec, data []byte) (bool, string) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return false, "empty stream"
	}

	lines := bytes.Split(trimmed, []byte("\n"))
	var (
		hasPayload   bool
		toolDetected bool
	)

	for _, rawLine := range lines {
		line := bytes.TrimSpace(rawLine)
		if len(line) == 0 || !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}

		payload := bytes.TrimSpace(line[len("data:"):])
		if len(payload) == 0 || bytes.Equal(payload, []byte("[DONE]")) {
			continue
		}

		hasPayload = true

		var obj map[string]any
		if err := json.Unmarshal(payload, &obj); err == nil {
			if errVal, ok := obj["error"]; ok && isMeaningfulErrorValue(errVal) {
				return false, snippet(payload)
			}
			if spec.Expectation == expectationToolInvocation && detectToolInvocationInStream(spec, obj) {
				toolDetected = true
			}
		}
	}

	if !hasPayload {
		return false, "empty stream payload"
	}

	if spec.Expectation == expectationToolInvocation && !toolDetected {
		return false, "stream missing tool invocation"
	}

	lower := bytes.ToLower(trimmed)
	if bytes.Contains(lower, []byte("\"error\"")) && !bytes.Contains(lower, []byte("\"error\":null")) {
		return false, snippet(trimmed)
	}

	return true, ""
}

// detectToolInvocationInStream inspects streamed JSON fragments for tool calls.
func detectToolInvocationInStream(spec requestSpec, obj map[string]any) bool {
	switch spec.Type {
	case requestTypeChatCompletion:
		if choices, ok := obj["choices"].([]any); ok {
			for _, choice := range choices {
				choiceMap, ok := choice.(map[string]any)
				if !ok {
					continue
				}
				if delta, ok := choiceMap["delta"].(map[string]any); ok {
					if calls, ok := delta["tool_calls"].([]any); ok && len(calls) > 0 {
						return true
					}
				}
				if message, ok := choiceMap["message"].(map[string]any); ok {
					if calls, ok := message["tool_calls"].([]any); ok && len(calls) > 0 {
						return true
					}
				}
			}
		}
	case requestTypeResponseAPI:
		if choices, ok := obj["choices"].([]any); ok {
			for _, choice := range choices {
				choiceMap, ok := choice.(map[string]any)
				if !ok {
					continue
				}
				if delta, ok := choiceMap["delta"].(map[string]any); ok {
					if calls, ok := delta["tool_calls"].([]any); ok && len(calls) > 0 {
						return true
					}
				}
				if message, ok := choiceMap["message"].(map[string]any); ok {
					if calls, ok := message["tool_calls"].([]any); ok && len(calls) > 0 {
						return true
					}
				}
			}
		}
		if responseObj, ok := obj["response"].(map[string]any); ok {
			if required, ok := responseObj["required_action"].(map[string]any); ok {
				if stringValue(required, "type") == "submit_tool_outputs" {
					if submit, ok := required["submit_tool_outputs"].(map[string]any); ok {
						if calls, ok := submit["tool_calls"].([]any); ok && len(calls) > 0 {
							return true
						}
					}
				}
			}
			if output, ok := responseObj["output"].([]any); ok {
				for _, entry := range output {
					entryMap, ok := entry.(map[string]any)
					if !ok {
						continue
					}
					if stringValue(entryMap, "type") == "function_call" {
						return true
					}
				}
			}
			if delta, ok := responseObj["delta"].(map[string]any); ok {
				if calls, ok := delta["tool_calls"].([]any); ok && len(calls) > 0 {
					return true
				}
			}
		}
		if required, ok := obj["required_action"].(map[string]any); ok {
			if stringValue(required, "type") == "submit_tool_outputs" {
				if submit, ok := required["submit_tool_outputs"].(map[string]any); ok {
					if calls, ok := submit["tool_calls"].([]any); ok && len(calls) > 0 {
						return true
					}
				}
			}
		}
		if output, ok := obj["output"].([]any); ok {
			for _, entry := range output {
				entryMap, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				if stringValue(entryMap, "type") == "function_call" {
					return true
				}
			}
		}
		if delta, ok := obj["delta"].(map[string]any); ok {
			if calls, ok := delta["tool_calls"].([]any); ok && len(calls) > 0 {
				return true
			}
		}
	case requestTypeClaudeMessages:
		if contentBlock, ok := obj["content_block"].(map[string]any); ok {
			if stringValue(contentBlock, "type") == "tool_use" {
				return true
			}
		}
		if content, ok := obj["content"].([]any); ok {
			for _, entry := range content {
				entryMap, ok := entry.(map[string]any)
				if !ok {
					continue
				}
				if stringValue(entryMap, "type") == "tool_use" {
					return true
				}
			}
		}
	}
	return false
}

func stringValue(data map[string]any, key string) string {
	if raw, ok := data[key]; ok {
		if s, ok := raw.(string); ok {
			return s
		}
	}
	return ""
}

func isMeaningfulErrorValue(val any) bool {
	switch v := val.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(v) != ""
	case map[string]any:
		if len(v) == 0 {
			return false
		}
		for _, nested := range v {
			if isMeaningfulErrorValue(nested) {
				return true
			}
		}
		return false
	case []any:
		return slices.ContainsFunc(v, isMeaningfulErrorValue)
	case bool:
		return v
	case float64:
		return v != 0
	default:
		return true
	}
}

func isUnsupportedCombination(reqType requestType, stream bool, statusCode int, body []byte, reason string) bool {
	text := reason
	if text == "" {
		text = snippet(body)
	}
	lower := strings.ToLower(text)

	switch reqType {
	case requestTypeResponseAPI:
		if strings.Contains(lower, "unknown field `messages`") ||
			strings.Contains(lower, "does not support responses") ||
			strings.Contains(lower, "response api is not available") {
			return true
		}
	case requestTypeChatCompletion:
		if strings.Contains(lower, "only supports response") ||
			strings.Contains(lower, "chat completions unsupported") {
			return true
		}
	case requestTypeClaudeMessages:
		if strings.Contains(lower, "does not support claude") ||
			strings.Contains(lower, "claude messages unsupported") {
			return true
		}
	}

	if stream && (strings.Contains(lower, "streaming is not supported") ||
		strings.Contains(lower, "stream parameter is not supported") ||
		strings.Contains(lower, "stream currently disabled")) {
		return true
	}

	if statusCode == http.StatusNotFound || statusCode == http.StatusMethodNotAllowed {
		return true
	}

	return false
}
