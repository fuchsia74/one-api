package main

import "strings"

// shorten trims whitespace and clamps the string to the provided rune length.
func shorten(text string, limit int) string {
	text = strings.TrimSpace(text)
	if limit <= 0 || len(text) <= limit {
		return text
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "…"
}

// truncateString clamps long strings while preserving rune safety.
func truncateString(text string, limit int) string {
	if limit <= 0 || len(text) <= limit {
		return text
	}
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "…"
}

// snippet trims response bodies for logging without exceeding 256 characters.
func snippet(body []byte) string {
	const maxLen = 256
	cleaned := strings.TrimSpace(string(body))
	if len(cleaned) <= maxLen {
		return cleaned
	}
	return cleaned[:maxLen] + "…"
}
