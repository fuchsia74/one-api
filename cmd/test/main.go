package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sort"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/Laisky/errors/v2"
	glog "github.com/Laisky/go-utils/v5/log"
	"github.com/Laisky/zap"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/sync/errgroup"

	cfg "github.com/songquanpeng/one-api/common/config"
)

type requestType string

const (
	requestTypeChatCompletion requestType = "chat_completion"
	requestTypeResponseAPI    requestType = "response_api"
	requestTypeClaudeMessages requestType = "claude_messages"

	defaultAPIBase      = "https://oneapi.laisky.com"
	defaultTestModels   = "gpt-4o-mini,gpt-5-mini,claude-3.5-haiku,gemini-2.5-flash,openai/gpt-oss-20b,deepseek-chat"
	defaultMaxTokens    = 2048
	defaultTemperature  = 0.7
	defaultTopP         = 0.9
	defaultTopK         = 40
	maxResponseBodySize = 1 << 20 // 1 MiB
	maxLoggedBodyBytes  = 2048
)

type requestVariant struct {
	Key         string
	Header      string
	Type        requestType
	Path        string
	Stream      bool
	Expectation expectation
}

// expectation describes what a request variant should validate in a response.
type expectation int

const (
	expectationDefault expectation = iota
	expectationToolInvocation
)

var requestVariants = []requestVariant{
	{Key: "chat_stream_false", Header: "Chat (stream=false)", Type: requestTypeChatCompletion, Path: "/v1/chat/completions", Stream: false, Expectation: expectationDefault},
	{Key: "chat_stream_true", Header: "Chat (stream=true)", Type: requestTypeChatCompletion, Path: "/v1/chat/completions", Stream: true, Expectation: expectationDefault},
	{Key: "chat_tools", Header: "Chat Tools", Type: requestTypeChatCompletion, Path: "/v1/chat/completions", Stream: false, Expectation: expectationToolInvocation},
	{Key: "response_stream_false", Header: "Response (stream=false)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: false, Expectation: expectationDefault},
	{Key: "response_stream_true", Header: "Response (stream=true)", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: true, Expectation: expectationDefault},
	{Key: "response_tools", Header: "Response Tools", Type: requestTypeResponseAPI, Path: "/v1/responses", Stream: false, Expectation: expectationToolInvocation},
	{Key: "claude_stream_false", Header: "Claude (stream=false)", Type: requestTypeClaudeMessages, Path: "/v1/messages", Stream: false, Expectation: expectationDefault},
	{Key: "claude_stream_true", Header: "Claude (stream=true)", Type: requestTypeClaudeMessages, Path: "/v1/messages", Stream: true, Expectation: expectationDefault},
	{Key: "claude_tools", Header: "Claude Tools", Type: requestTypeClaudeMessages, Path: "/v1/messages", Stream: false, Expectation: expectationToolInvocation},
}

type testResult struct {
	Model        string
	Variant      string
	Label        string
	Type         requestType
	Stream       bool
	Success      bool
	Skipped      bool
	StatusCode   int
	Duration     time.Duration
	ErrorReason  string
	RequestBody  string
	ResponseBody string
}

type requestSpec struct {
	Variant     string
	Label       string
	Type        requestType
	Path        string
	Body        any
	Stream      bool
	Expectation expectation
}

func main() {
	logger, err := glog.NewConsoleWithName("oneapi-test", glog.LevelInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %+v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, logger); err != nil {
		logger.Error("test run failed", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("all tests passed")
}

func run(ctx context.Context, logger glog.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return errors.Wrap(err, "load config")
	}

	variantLabels := make([]string, 0, len(cfg.Variants))
	for _, v := range cfg.Variants {
		variantLabels = append(variantLabels, v.Header)
	}
	logger.Info("starting API regression sweep",
		zap.String("base_url", cfg.APIBase),
		zap.Int("model_count", len(cfg.Models)),
		zap.Int("variant_count", len(cfg.Variants)),
		zap.Strings("variants", variantLabels),
	)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resultsCh := make(chan testResult, len(cfg.Models)*len(cfg.Variants))

	var (
		results   []testResult
		collectWg sync.WaitGroup
	)

	collectWg.Go(func() {
		for res := range resultsCh {
			results = append(results, res)
			switch {
			case res.Success:
				logger.Info("request succeeded",
					zap.String("model", res.Model),
					zap.String("variant", res.Label),
					zap.String("type", string(res.Type)),
					zap.Bool("stream", res.Stream),
					zap.Duration("duration", res.Duration),
					zap.Int("status", res.StatusCode),
				)
			case res.Skipped:
				logger.Info("request skipped",
					zap.String("model", res.Model),
					zap.String("variant", res.Label),
					zap.String("type", string(res.Type)),
					zap.Bool("stream", res.Stream),
					zap.Int("status", res.StatusCode),
					zap.String("reason", res.ErrorReason),
				)
			default:
				logger.Warn("request failed",
					zap.String("model", res.Model),
					zap.String("variant", res.Label),
					zap.String("type", string(res.Type)),
					zap.Bool("stream", res.Stream),
					zap.Duration("duration", res.Duration),
					zap.Int("status", res.StatusCode),
					zap.String("error", res.ErrorReason),
					zap.String("request_body", res.RequestBody),
					zap.String("response_body", res.ResponseBody),
				)
			}
		}
	})

	grp, grpCtx := errgroup.WithContext(ctx)
	for _, modelName := range cfg.Models {
		model := modelName
		grp.Go(func() error {
			executeModelSweep(grpCtx, httpClient, cfg, model, resultsCh)
			return nil
		})
	}

	_ = grp.Wait()
	close(resultsCh)
	collectWg.Wait()

	report := buildReport(cfg.Models, cfg.Variants, results)
	renderReport(report)

	if report.failedCount > 0 {
		return errors.Errorf("%d of %d requests failed", report.failedCount, report.totalRequests)
	}

	return nil
}

type config struct {
	APIBase  string
	Token    string
	Models   []string
	Variants []requestVariant
}

func loadConfig() (config, error) {
	base := strings.TrimSpace(cfg.OneAPITestAPIBase)
	if base == "" {
		base = defaultAPIBase
	}

	token := strings.TrimSpace(cfg.OneAPITestToken)
	if token == "" {
		return config{}, errors.Errorf("API_TOKEN must be set")
	}

	modelsRaw := cfg.OneAPITestModels
	models, err := parseModels(modelsRaw)
	if err != nil {
		return config{}, errors.Wrap(err, "parse models")
	}
	if len(models) == 0 {
		models, err = parseModels(defaultTestModels)
		if err != nil {
			return config{}, errors.Wrap(err, "parse default models")
		}
	}

	variantsRaw := cfg.OneAPITestVariants
	variants, err := parseVariants(variantsRaw)
	if err != nil {
		return config{}, errors.Wrap(err, "parse variants")
	}

	return config{
		APIBase:  strings.TrimSuffix(base, "/"),
		Token:    token,
		Models:   models,
		Variants: variants,
	}, nil
}

func parseModels(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	separators := []string{",", ";", "\n", "\r"}
	normalized := raw
	for _, sep := range separators {
		normalized = strings.ReplaceAll(normalized, sep, ",")
	}

	parts := strings.Split(normalized, ",")
	if len(parts) == 1 && !strings.Contains(raw, ",") && !strings.Contains(raw, ";") && !strings.Contains(raw, "\n") {
		parts = strings.Fields(raw)
	}

	var models []string
	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		models = append(models, candidate)
	}

	return models, nil
}

func parseVariants(raw string) ([]requestVariant, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return requestVariants, nil
	}

	separators := []string{",", ";", "\n", "\r"}
	normalized := raw
	for _, sep := range separators {
		normalized = strings.ReplaceAll(normalized, sep, ",")
	}
	parts := strings.Split(normalized, ",")
	if len(parts) == 1 && !strings.Contains(raw, ",") && !strings.Contains(raw, ";") && !strings.Contains(raw, "\n") {
		parts = strings.Fields(raw)
	}

	selected := make([]requestVariant, 0, len(requestVariants))
	seen := make(map[string]bool, len(requestVariants))
	typeGroups := map[string]requestType{
		"chat":            requestTypeChatCompletion,
		"chat_completion": requestTypeChatCompletion,
		"response":        requestTypeResponseAPI,
		"responses":       requestTypeResponseAPI,
		"response_api":    requestTypeResponseAPI,
		"claude":          requestTypeClaudeMessages,
		"claude_messages": requestTypeClaudeMessages,
	}

	for _, part := range parts {
		candidate := strings.TrimSpace(part)
		if candidate == "" {
			continue
		}
		lower := strings.ToLower(candidate)

		matched := false
		for _, variant := range requestVariants {
			if strings.EqualFold(candidate, variant.Key) || strings.EqualFold(candidate, variant.Header) {
				if !seen[variant.Key] {
					selected = append(selected, variant)
					seen[variant.Key] = true
				}
				matched = true
				break
			}
		}
		if matched {
			continue
		}

		if groupType, ok := typeGroups[lower]; ok {
			for _, variant := range requestVariants {
				if variant.Type == groupType && !seen[variant.Key] {
					selected = append(selected, variant)
					seen[variant.Key] = true
				}
			}
			matched = true
		}

		if !matched {
			return nil, errors.Errorf("unknown variant or api format %q", candidate)
		}
	}

	if len(selected) == 0 {
		return nil, errors.New("no variants selected")
	}

	return selected, nil
}

func executeModelSweep(ctx context.Context, client *http.Client, cfg config, model string, results chan<- testResult) {
	specs := buildRequestSpecs(model, cfg.Variants)

	innerGrp, innerCtx := errgroup.WithContext(ctx)
	for _, spec := range specs {
		innerGrp.Go(func() error {
			res := performRequest(innerCtx, client, cfg.APIBase, cfg.Token, spec, model)
			select {
			case results <- res:
			case <-innerCtx.Done():
			}
			return nil
		})
	}

	_ = innerGrp.Wait()
}

func buildRequestSpecs(model string, variants []requestVariant) []requestSpec {
	specs := make([]requestSpec, 0, len(variants))
	for _, variant := range variants {
		var body any
		switch variant.Type {
		case requestTypeChatCompletion:
			body = chatCompletionPayload(model, variant.Stream, variant.Expectation)
		case requestTypeResponseAPI:
			body = responseAPIPayload(model, variant.Stream, variant.Expectation)
		case requestTypeClaudeMessages:
			body = claudeMessagesPayload(model, variant.Stream, variant.Expectation)
		default:
			body = nil
		}

		specs = append(specs, requestSpec{
			Variant:     variant.Key,
			Label:       variant.Header,
			Type:        variant.Type,
			Path:        variant.Path,
			Body:        body,
			Stream:      variant.Stream,
			Expectation: variant.Expectation,
		})
	}

	return specs
}

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
			result.Skipped = false
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
			result.Skipped = false
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
		result.Skipped = false
		result.ErrorReason = reason
		return
	}

	result.ErrorReason = reason
	return
}

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
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return buffer.Bytes(), err
		}
		if len(strings.TrimSpace(string(chunk))) == 0 && buffer.Len() > 0 {
			break
		}
	}

	if buffer.Len() == 0 {
		return buffer.Bytes(), fmt.Errorf("no stream data received")
	}

	return buffer.Bytes(), nil
}

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
					message, ok := choiceMap["message"].(map[string]any)
					if !ok {
						continue
					}
					if calls, ok := message["tool_calls"].([]any); ok && len(calls) > 0 {
						return true, ""
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

// hasFunctionCallOutput reports whether the Response API payload contains a function_call entry
// inside the output array. Recent OpenAI responses can surface tool instructions directly in
// output when the request specifies a concrete tool_choice, bypassing required_action.
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

func evaluateStreamResponse(spec requestSpec, data []byte) (bool, string) {
	_ = spec
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return false, "empty stream"
	}

	lines := bytes.SplitSeq(trimmed, []byte("\n"))
	for line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if bytes.HasPrefix(line, []byte("data:")) {
			payload := bytes.TrimSpace(line[len("data:"):])
			if len(payload) == 0 {
				continue
			}
			var obj map[string]any
			if err := json.Unmarshal(payload, &obj); err == nil {
				if errVal, ok := obj["error"]; ok && isMeaningfulErrorValue(errVal) {
					return false, snippet(payload)
				}
			}
		}
	}

	lower := bytes.ToLower(trimmed)
	if bytes.Contains(lower, []byte("\"error\"")) && !bytes.Contains(lower, []byte("\"error\":null")) {
		return false, snippet(trimmed)
	}

	return true, ""
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

func chatCompletionPayload(model string, stream bool, exp expectation) any {
	base := map[string]any{
		"model":       model,
		"max_tokens":  defaultMaxTokens,
		"temperature": defaultTemperature,
		"top_p":       defaultTopP,
		"stream":      stream,
	}

	if exp == expectationToolInvocation {
		base["messages"] = []map[string]any{
			{
				"role":    "system",
				"content": "You are a weather assistant that must call tools when asked for weather information.",
			},
			{
				"role":    "user",
				"content": "What is the weather in San Francisco, CA right now? Use the tool to find out.",
			},
		}
		base["tools"] = []map[string]any{chatWeatherToolDefinition()}
		base["tool_choice"] = map[string]any{
			"type": "function",
			"function": map[string]string{
				"name": "get_weather",
			},
		}
		return base
	}

	base["messages"] = []map[string]any{
		{
			"role":    "user",
			"content": "Say hello in one sentence.",
		},
	}
	return base
}

func responseAPIPayload(model string, stream bool, exp expectation) any {
	base := map[string]any{
		"model":             model,
		"max_output_tokens": defaultMaxTokens,
		"temperature":       defaultTemperature,
		"top_p":             defaultTopP,
		"stream":            stream,
	}

	if exp == expectationToolInvocation {
		base["input"] = []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "input_text",
						"text": "Please call get_weather for San Francisco, CA in celsius and report the findings.",
					},
				},
			},
		}
		base["tools"] = []map[string]any{responseWeatherToolDefinition()}
		base["tool_choice"] = map[string]any{
			"type": "tool",
			"name": "get_weather",
		}
		return base
	}

	base["input"] = []map[string]any{
		{
			"role": "user",
			"content": []map[string]any{
				{
					"type": "input_text",
					"text": "Say hello in one sentence.",
				},
			},
		},
	}
	return base
}

func claudeMessagesPayload(model string, stream bool, exp expectation) any {
	base := map[string]any{
		"model":       model,
		"max_tokens":  defaultMaxTokens,
		"temperature": defaultTemperature,
		"top_p":       defaultTopP,
		"top_k":       defaultTopK,
		"stream":      stream,
	}

	if exp == expectationToolInvocation {
		base["messages"] = []map[string]any{
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type": "text",
						"text": "Use the get_weather tool to retrieve today's weather in San Francisco, CA.",
					},
				},
			},
		}
		base["tools"] = []map[string]any{claudeWeatherToolDefinition()}
		base["tool_choice"] = map[string]any{
			"type": "tool",
			"name": "get_weather",
		}
		return base
	}

	base["messages"] = []map[string]any{
		{
			"role": "user",
			"content": []map[string]any{
				{
					"type": "text",
					"text": "Say hello in one sentence.",
				},
			},
		},
	}
	return base
}

func chatWeatherToolDefinition() map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        "get_weather",
			"description": "Get the current weather for a location",
			"parameters":  weatherFunctionSchema(),
		},
	}
}

func responseWeatherToolDefinition() map[string]any {
	return map[string]any{
		"type":        "function",
		"name":        "get_weather",
		"description": "Get the current weather for a location",
		"parameters":  weatherFunctionSchema(),
	}
}

func claudeWeatherToolDefinition() map[string]any {
	return map[string]any{
		"name":         "get_weather",
		"description":  "Get the current weather for a location",
		"input_schema": weatherFunctionSchema(),
	}
}

func weatherFunctionSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{
				"type":        "string",
				"description": "City and region to look up (example: San Francisco, CA)",
			},
			"unit": map[string]any{
				"type":        "string",
				"description": "Temperature unit to use",
				"enum":        []string{"celsius", "fahrenheit"},
			},
		},
		"required": []string{"location"},
	}
}

type report struct {
	models         []string
	variants       []requestVariant
	resultsByModel map[string]map[string]testResult
	totalRequests  int
	failedCount    int
	skippedCount   int
}

func buildReport(models []string, variants []requestVariant, results []testResult) report {
	byModel := make(map[string]map[string]testResult, len(models))
	for _, model := range models {
		byModel[model] = make(map[string]testResult)
	}

	failed := 0
	skipped := 0
	for _, res := range results {
		if res.Model == "" {
			continue
		}
		modelMap, ok := byModel[res.Model]
		if !ok {
			modelMap = make(map[string]testResult)
			byModel[res.Model] = modelMap
		}
		modelMap[res.Variant] = res
		if res.Skipped {
			skipped++
			continue
		}
		if !res.Success {
			failed++
		}
	}

	return report{
		models:         models,
		variants:       variants,
		resultsByModel: byModel,
		totalRequests:  len(results),
		failedCount:    failed,
		skippedCount:   skipped,
	}
}

func renderReport(rep report) {
	if len(rep.models) == 0 {
		fmt.Println("no models to report")
		return
	}
	if len(rep.variants) == 0 {
		fmt.Println("no api formats selected")
		return
	}

	fmt.Println()
	fmt.Println("=== One-API Regression Matrix ===")
	fmt.Println()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "Variant")
	for _, model := range rep.models {
		fmt.Fprintf(tw, "\t%s", model)
	}
	fmt.Fprintln(tw)

	for _, variant := range rep.variants {
		fmt.Fprintf(tw, "%s", variant.Header)
		for _, model := range rep.models {
			entry := rep.resultsByModel[model]
			cell := formatMatrixCell(entry[variant.Key])
			fmt.Fprintf(tw, "\t%s", cell)
		}
		fmt.Fprintln(tw)
	}
	_ = tw.Flush()

	fmt.Println()

	passed := rep.totalRequests - rep.failedCount - rep.skippedCount
	fmt.Printf("Totals  | Requests: %d | Passed: %d | Failed: %d | Skipped: %d\n",
		rep.totalRequests,
		passed,
		rep.failedCount,
		rep.skippedCount,
	)

	failures, skips := gatherOutcomes(rep)
	if len(failures) > 0 {
		fmt.Println()
		fmt.Println("Failures:")
		for _, res := range failures {
			fmt.Printf("- %s · %s → %s\n", res.Model, res.Label, shorten(res.ErrorReason, 200))
		}
	}
	if len(skips) > 0 {
		fmt.Println()
		fmt.Println("Skipped (unsupported combinations):")
		for _, res := range skips {
			fmt.Printf("- %s · %s → %s\n", res.Model, res.Label, shorten(res.ErrorReason, 200))
		}
	}

	fmt.Println()
}

func formatMatrixCell(res testResult) string {
	if res.Model == "" {
		return "—"
	}

	duration := res.Duration.Truncate(10 * time.Millisecond)
	switch {
	case res.Success:
		return fmt.Sprintf("PASS %.2fs", duration.Seconds())
	case res.Skipped:
		reason := res.ErrorReason
		if reason == "" {
			reason = "skipped"
		}
		return fmt.Sprintf("SKIP %s", shorten(reason, 32))
	default:
		reason := res.ErrorReason
		if reason == "" {
			reason = duration.String()
		}
		return fmt.Sprintf("FAIL %s", shorten(reason, 32))
	}
}

func gatherOutcomes(rep report) (failures, skips []testResult) {
	for _, model := range rep.models {
		entry := rep.resultsByModel[model]
		for _, variant := range rep.variants {
			res, ok := entry[variant.Key]
			if !ok || res.Model == "" {
				continue
			}
			if res.Skipped {
				skips = append(skips, res)
				continue
			}
			if !res.Success {
				failures = append(failures, res)
			}
		}
	}

	sort.Slice(failures, func(i, j int) bool {
		if failures[i].Model == failures[j].Model {
			return failures[i].Label < failures[j].Label
		}
		return failures[i].Model < failures[j].Model
	})
	sort.Slice(skips, func(i, j int) bool {
		if skips[i].Model == skips[j].Model {
			return skips[i].Label < skips[j].Label
		}
		return skips[i].Model < skips[j].Model
	})

	return failures, skips
}

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

func snippet(body []byte) string {
	const maxLen = 256
	cleaned := strings.TrimSpace(string(body))
	if len(cleaned) <= maxLen {
		return cleaned
	}
	return cleaned[:maxLen] + "…"
}
