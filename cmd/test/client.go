package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// performRequest sends a single request variant and returns the execution result.
func performRequest(ctx context.Context, client *http.Client, baseURL, token string, spec requestSpec, model string) (result testResult) {
	start := time.Now()
	result = testResult{
		Model:   model,
		Variant: spec.Variant,
		Label:   spec.Label,
		Type:    spec.Type,
		Stream:  spec.Stream,
	}
	defer func() {
		result.Duration = time.Since(start)
	}()

	payload, err := json.Marshal(spec.Body)
	if err != nil {
		result.ErrorReason = fmt.Sprintf("marshal payload: %v", err)
		return
	}
	result.RequestBody = truncateString(string(payload), maxLoggedBodyBytes)

	endpoint := baseURL + spec.Path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		result.ErrorReason = fmt.Sprintf("build request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "oneapi-test-harness/1.0")

	resp, err := client.Do(req)
	if err != nil {
		result.ErrorReason = fmt.Sprintf("do request: %v", err)
		return
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	if spec.Stream {
		streamData, streamErr := collectStreamBody(resp.Body, maxResponseBodySize)
		if len(streamData) > 0 {
			result.ResponseBody = truncateString(string(streamData), maxLoggedBodyBytes)
		}
		if streamErr != nil {
			result.ErrorReason = fmt.Sprintf("stream read: %v", streamErr)
			return
		}

		success, reason := evaluateStreamResponse(spec, streamData)
		if success {
			result.Success = true
			return
		}

		if isUnsupportedCombination(spec.Type, spec.Stream, resp.StatusCode, streamData, reason) {
			result.Skipped = true
			result.ErrorReason = reason
			return
		}

		if reason == "" {
			reason = snippet(streamData)
		}
		result.ErrorReason = reason
		return
	}

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize))
	if len(body) > 0 {
		result.ResponseBody = truncateString(string(body), maxLoggedBodyBytes)
	}
	if readErr != nil {
		result.ErrorReason = fmt.Sprintf("read response: %v", readErr)
		return
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		success, reason := evaluateResponse(spec, body)
		if success {
			result.Success = true
			return
		}

		if isUnsupportedCombination(spec.Type, spec.Stream, resp.StatusCode, body, reason) {
			result.Skipped = true
			result.ErrorReason = reason
			return
		}

		if reason == "" {
			reason = snippet(body)
		}
		result.ErrorReason = reason
		return
	}

	reason := fmt.Sprintf("status %s: %s", resp.Status, snippet(body))
	if isUnsupportedCombination(spec.Type, spec.Stream, resp.StatusCode, body, reason) {
		result.Skipped = true
		result.ErrorReason = reason
		return
	}

	result.ErrorReason = reason
	return
}

// collectStreamBody reads a streaming response until EOF, blank line, or size limit.
func collectStreamBody(body io.Reader, limit int) ([]byte, error) {
	reader := bufio.NewReader(body)
	buffer := &bytes.Buffer{}

	for buffer.Len() < limit {
		chunk, err := reader.ReadBytes('\n')
		if len(chunk) > 0 {
			if buffer.Len()+len(chunk) > limit {
				chunk = chunk[:limit-buffer.Len()]
			}
			buffer.Write(chunk)
			trimmed := bytes.TrimSpace(chunk)
			if bytes.Equal(trimmed, []byte("data: [DONE]")) || bytes.Equal(trimmed, []byte("[DONE]")) {
				break
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return buffer.Bytes(), err
		}
	}

	if buffer.Len() == 0 {
		return buffer.Bytes(), fmt.Errorf("no stream data received")
	}

	return buffer.Bytes(), nil
}
