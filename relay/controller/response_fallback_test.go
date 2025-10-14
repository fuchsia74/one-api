package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/adaptor/openai_compatible"
	"github.com/songquanpeng/one-api/relay/channeltype"
	metalib "github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

const (
	fallbackUserID              = 99001
	fallbackTokenID             = 99002
	fallbackChannelID           = 99003
	fallbackCompatibleChannelID = 99004
)

func TestRenderChatResponseAsResponseAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Set(ctxkey.RequestId, "req_123")

	textResp := &openai_compatible.SlimTextResponse{
		Choices: []openai_compatible.TextResponseChoice{
			{
				Message:      relaymodel.Message{Role: "assistant", Content: "Hello there"},
				FinishReason: "stop",
			},
		},
		Usage: relaymodel.Usage{PromptTokens: 12, CompletionTokens: 8, TotalTokens: 20},
	}

	parallel := true
	request := &openai.ResponseAPIRequest{ParallelToolCalls: &parallel}
	meta := &metalib.Meta{ActualModelName: "gpt-fallback"}

	if err := renderChatResponseAsResponseAPI(c, http.StatusOK, textResp, request, meta); err != nil {
		t.Fatalf("unexpected error rendering response: %v", err)
	}

	var resp openai.ResponseAPIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if resp.Model != "gpt-fallback" {
		t.Fatalf("expected model gpt-fallback, got %s", resp.Model)
	}
	if resp.Status != "completed" {
		t.Fatalf("expected status completed, got %s", resp.Status)
	}
	if len(resp.Output) != 1 {
		t.Fatalf("expected single output item, got %d", len(resp.Output))
	}
	if resp.Output[0].Type != "message" {
		t.Fatalf("expected message output, got %s", resp.Output[0].Type)
	}
	if len(resp.Output[0].Content) == 0 || resp.Output[0].Content[0].Text != "Hello there" {
		t.Fatalf("expected message content preserved, got %#v", resp.Output[0].Content)
	}
	if resp.Usage == nil || resp.Usage.TotalTokens != 20 {
		t.Fatalf("expected usage to be carried over, got %#v", resp.Usage)
	}
	if !resp.ParallelToolCalls {
		t.Fatalf("expected parallel tool calls to be true")
	}
}

func TestRelayResponseAPIHelper_FallbackAzure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ensureResponseFallbackFixtures(t)

	prevRedis := common.IsRedisEnabled()
	common.SetRedisEnabled(false)
	t.Cleanup(func() { common.SetRedisEnabled(prevRedis) })

	prevLogConsume := config.IsLogConsumeEnabled()
	config.SetLogConsumeEnabled(false)
	t.Cleanup(func() { config.SetLogConsumeEnabled(prevLogConsume) })

	upstreamCalled := false
	var upstreamPath string
	var upstreamBody []byte

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		upstreamPath = r.URL.Path
		if r.URL.RawQuery != "" {
			upstreamPath += "?" + r.URL.RawQuery
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read upstream body: %v", err)
		}
		upstreamBody = body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
		  "id": "chatcmpl-123",
		  "object": "chat.completion",
		  "created": 1741036800,
		  "model": "gpt-4o-mini",
		  "choices": [
		    {
		      "index": 0,
		      "message": {"role": "assistant", "content": "Hi there!"},
		      "finish_reason": "stop"
		    }
		  ],
		  "usage": {"prompt_tokens": 5, "completion_tokens": 8, "total_tokens": 13}
		}`))
	}))
	defer upstream.Close()

	prevClient := client.HTTPClient
	client.HTTPClient = upstream.Client()
	t.Cleanup(func() { client.HTTPClient = prevClient })

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	requestPayload := `{"model":"gpt-4o-mini","stream":false,"instructions":"You are helpful.","input":[{"role":"user","content":[{"type":"input_text","text":"Hello via response API"}]}],"parallel_tool_calls":true}`
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", bytes.NewBufferString(requestPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer azure-key")
	c.Request = req

	gmw.SetLogger(c, logger.Logger)

	c.Set(ctxkey.Channel, channeltype.Azure)
	c.Set(ctxkey.ChannelId, fallbackChannelID)
	c.Set(ctxkey.TokenId, fallbackTokenID)
	c.Set(ctxkey.TokenName, "fallback-token")
	c.Set(ctxkey.Id, fallbackUserID)
	c.Set(ctxkey.Group, "default")
	c.Set(ctxkey.ModelMapping, map[string]string{})
	c.Set(ctxkey.ChannelRatio, 1.0)
	c.Set(ctxkey.RequestModel, "gpt-4o-mini")
	c.Set(ctxkey.BaseURL, upstream.URL)
	c.Set(ctxkey.ContentType, "application/json")
	c.Set(ctxkey.RequestId, "req_fallback")
	c.Set(ctxkey.TokenQuotaUnlimited, true)
	c.Set(ctxkey.TokenQuota, int64(0))
	c.Set(ctxkey.Username, "response-fallback")
	c.Set(ctxkey.UserQuota, int64(1_000_000))
	c.Set(ctxkey.ChannelModel, &model.Channel{Id: fallbackChannelID, Type: channeltype.Azure})
	c.Set(ctxkey.Config, model.ChannelConfig{APIVersion: "2024-02-15-preview"})

	if err := RelayResponseAPIHelper(c); err != nil {
		t.Fatalf("RelayResponseAPIHelper returned error: %v", err)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
	if !upstreamCalled {
		t.Fatalf("expected upstream to be called")
	}
	if !strings.Contains(upstreamPath, "/openai/deployments/gpt-4o-mini/chat/completions") {
		t.Fatalf("unexpected upstream path: %s", upstreamPath)
	}
	if !strings.Contains(upstreamPath, "api-version=") {
		t.Fatalf("expected api-version query parameter in upstream path: %s", upstreamPath)
	}

	var fallbackResp openai.ResponseAPIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &fallbackResp); err != nil {
		t.Fatalf("failed to unmarshal fallback response body: %v", err)
	}
	if fallbackResp.Status != "completed" {
		t.Fatalf("expected response status completed, got %s", fallbackResp.Status)
	}
	if len(fallbackResp.Output) != 1 {
		t.Fatalf("expected single output item, got %d", len(fallbackResp.Output))
	}
	output := fallbackResp.Output[0]
	if output.Type != "message" {
		t.Fatalf("expected message output type, got %s", output.Type)
	}
	if len(output.Content) == 0 || output.Content[0].Text != "Hi there!" {
		t.Fatalf("unexpected output content: %#v", output.Content)
	}
	if fallbackResp.Usage == nil || fallbackResp.Usage.TotalTokens != 13 {
		t.Fatalf("unexpected usage: %#v", fallbackResp.Usage)
	}
	if fallbackResp.RequiredAction != nil {
		t.Fatalf("did not expect required_action for non-tool response, got %#v", fallbackResp.RequiredAction)
	}
	if !fallbackResp.ParallelToolCalls {
		t.Fatalf("expected parallel tool calls to remain true")
	}

	var chatReq relaymodel.GeneralOpenAIRequest
	if err := json.Unmarshal(upstreamBody, &chatReq); err != nil {
		t.Fatalf("failed to unmarshal upstream chat request: %v", err)
	}
	if chatReq.Model != "gpt-4o-mini" {
		t.Fatalf("expected chat request model gpt-4o-mini, got %s", chatReq.Model)
	}
	if len(chatReq.Messages) != 2 {
		t.Fatalf("expected two messages (system + user), got %d", len(chatReq.Messages))
	}
	if chatReq.Messages[0].Role != "system" || chatReq.Messages[0].StringContent() != "You are helpful." {
		t.Fatalf("system message not preserved: %#v", chatReq.Messages[0])
	}
	if chatReq.Messages[1].Role != "user" || chatReq.Messages[1].StringContent() != "Hello via response API" {
		t.Fatalf("user message not preserved: %#v", chatReq.Messages[1])
	}
}

func TestRelayResponseAPIHelper_FallbackStreaming(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ensureResponseFallbackFixtures(t)

	prevRedis := common.IsRedisEnabled()
	common.SetRedisEnabled(false)
	t.Cleanup(func() { common.SetRedisEnabled(prevRedis) })

	prevLogConsume := config.IsLogConsumeEnabled()
	config.SetLogConsumeEnabled(false)
	t.Cleanup(func() { config.SetLogConsumeEnabled(prevLogConsume) })

	upstreamCalled := false
	var upstreamPath string
	var upstreamBody []byte

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalled = true
		upstreamPath = r.URL.Path
		if r.URL.RawQuery != "" {
			upstreamPath += "?" + r.URL.RawQuery
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read upstream body: %v", err)
		}
		upstreamBody = body

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatalf("response writer does not support flushing")
		}

		chunks := []string{
			`{"id":"chatcmpl-123","object":"chat.completion.chunk","created":1741036800,"model":"gpt-4o-mini","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}`,
			`{"id":"chatcmpl-123","object":"chat.completion.chunk","created":1741036800,"model":"gpt-4o-mini","choices":[{"index":0,"delta":{"content":" world!"},"finish_reason":null}]}`,
			`{"id":"chatcmpl-123","object":"chat.completion.chunk","created":1741036800,"model":"gpt-4o-mini","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":7,"total_tokens":12}}`,
		}
		for _, chunk := range chunks {
			fmt.Fprintf(w, "data: %s\n\n", chunk)
			flusher.Flush()
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer upstream.Close()

	prevClient := client.HTTPClient
	client.HTTPClient = upstream.Client()
	t.Cleanup(func() { client.HTTPClient = prevClient })

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	requestPayload := `{"model":"gpt-4o-mini","stream":true,"instructions":"You are helpful.","input":[{"role":"user","content":[{"type":"input_text","text":"Hello via response API stream"}]}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", strings.NewReader(requestPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer compat-key")
	c.Request = req

	gmw.SetLogger(c, logger.Logger)

	c.Set(ctxkey.Channel, channeltype.OpenAICompatible)
	c.Set(ctxkey.ChannelId, fallbackCompatibleChannelID)
	c.Set(ctxkey.TokenId, fallbackTokenID)
	c.Set(ctxkey.TokenName, "fallback-token")
	c.Set(ctxkey.Id, fallbackUserID)
	c.Set(ctxkey.Group, "default")
	c.Set(ctxkey.ModelMapping, map[string]string{})
	c.Set(ctxkey.ChannelRatio, 1.0)
	c.Set(ctxkey.RequestModel, "gpt-4o-mini")
	c.Set(ctxkey.BaseURL, upstream.URL)
	c.Set(ctxkey.ContentType, "application/json")
	c.Set(ctxkey.RequestId, "req_fallback_stream")
	c.Set(ctxkey.TokenQuotaUnlimited, true)
	c.Set(ctxkey.TokenQuota, int64(0))
	c.Set(ctxkey.Username, "response-fallback")
	c.Set(ctxkey.UserQuota, int64(1_000_000))
	c.Set(ctxkey.ChannelModel, &model.Channel{Id: fallbackCompatibleChannelID, Type: channeltype.OpenAICompatible})
	c.Set(ctxkey.Config, model.ChannelConfig{})

	if err := RelayResponseAPIHelper(c); err != nil {
		t.Fatalf("RelayResponseAPIHelper returned error: %v", err)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
	if !upstreamCalled {
		t.Fatalf("expected upstream to be called")
	}
	if !(strings.Contains(upstreamPath, "/v1/chat/completions") || strings.Contains(upstreamPath, "/chat/completions")) {
		t.Fatalf("unexpected upstream path for streaming fallback: %s", upstreamPath)
	}

	var chatReq relaymodel.GeneralOpenAIRequest
	if err := json.Unmarshal(upstreamBody, &chatReq); err != nil {
		t.Fatalf("failed to unmarshal upstream chat request: %v", err)
	}
	if !chatReq.Stream {
		t.Fatalf("expected chat request stream flag to be true")
	}
	if len(chatReq.Messages) == 0 || chatReq.Messages[0].StringContent() == "" {
		t.Fatalf("expected user message in chat request: %#v", chatReq.Messages)
	}

	events := parseSSEEvents(recorder.Body.String())
	t.Logf("raw SSE: %s", recorder.Body.String())
	t.Logf("parsed SSE events: %+v", events)
	if len(events) == 0 {
		t.Fatalf("expected SSE events, got none")
	}

	var (
		seenCreated   bool
		seenCompleted bool
		deltaCount    int
		finalResponse *openai.ResponseAPIResponse
	)

	for idx, ev := range events {
		if idx == len(events)-1 {
			if ev.event != "" || strings.TrimSpace(ev.data) != "[DONE]" {
				t.Fatalf("expected final SSE chunk to be [DONE], got event=%q data=%q", ev.event, ev.data)
			}
			continue
		}
		switch ev.event {
		case "response.created":
			seenCreated = true
		case "response.output_text.delta":
			deltaCount++
		case "response.completed":
			seenCompleted = true
			var streamEvent openai.ResponseAPIStreamEvent
			if err := json.Unmarshal([]byte(ev.data), &streamEvent); err != nil {
				t.Fatalf("failed to unmarshal response.completed event: %v", err)
			}
			if streamEvent.Response == nil {
				t.Fatalf("expected response payload in response.completed event")
			}
			finalResponse = streamEvent.Response
		}
	}

	if !seenCreated {
		t.Fatalf("missing response.created event")
	}
	if deltaCount < 2 {
		t.Fatalf("expected at least two delta events, got %d", deltaCount)
	}
	if !seenCompleted || finalResponse == nil {
		t.Fatalf("missing response.completed event")
	}
	if finalResponse.Status != "completed" {
		t.Fatalf("expected final status completed, got %s", finalResponse.Status)
	}
	if len(finalResponse.Output) == 0 || len(finalResponse.Output[0].Content) == 0 {
		t.Fatalf("expected output message in final response: %#v", finalResponse.Output)
	}
	if text := finalResponse.Output[0].Content[0].Text; text != "Hello world!" {
		t.Fatalf("unexpected final response text: %q", text)
	}
	if finalResponse.Usage == nil || finalResponse.Usage.TotalTokens != 12 {
		t.Fatalf("unexpected usage in final response: %#v", finalResponse.Usage)
	}
}

func TestNormalizeResponseAPIRawBody_RemovesUnsupportedParams(t *testing.T) {
	temp := 0.7
	topP := 0.9
	req := &openai.ResponseAPIRequest{Model: "gpt-5-mini", Temperature: &temp, TopP: &topP}

	sanitizeResponseAPIRequest(req)
	if req.Temperature != nil {
		t.Fatalf("expected temperature pointer to be nil after sanitization")
	}
	if req.TopP != nil {
		t.Fatalf("expected top_p pointer to be nil after sanitization")
	}

	raw := []byte(`{"model":"gpt-5-mini","temperature":0.7,"top_p":0.9}`)
	patched, err := normalizeResponseAPIRawBody(raw, req)
	if err != nil {
		t.Fatalf("normalizeResponseAPIRawBody failed: %v", err)
	}
	if bytes.Contains(patched, []byte("\"temperature\"")) {
		t.Fatalf("temperature should have been removed: %s", patched)
	}
	if bytes.Contains(patched, []byte("\"top_p\"")) {
		t.Fatalf("top_p should have been removed: %s", patched)
	}
}

type parsedSSE struct {
	event string
	data  string
}

func parseSSEEvents(raw string) []parsedSSE {
	var events []parsedSSE
	remaining := raw
	for len(remaining) > 0 {
		chunkEnd := strings.Index(remaining, "\n\n")
		var chunk string
		if chunkEnd == -1 {
			chunk = remaining
			remaining = ""
		} else {
			chunk = remaining[:chunkEnd]
			remaining = remaining[chunkEnd+2:]
		}
		chunk = strings.Trim(chunk, "\n")
		if chunk == "" {
			continue
		}

		lines := strings.Split(chunk, "\n")
		var ev parsedSSE
		for _, line := range lines {
			line = strings.TrimRight(line, "\r")
			if strings.HasPrefix(line, "event: ") {
				ev.event = strings.TrimSpace(line[len("event: "):])
			} else if strings.HasPrefix(line, "data: ") {
				dataLine := line[len("data: "):]
				if ev.data != "" {
					ev.data += "\n"
				}
				ev.data += dataLine
			}
		}
		if ev.event != "" || ev.data != "" {
			events = append(events, ev)
		}
	}
	return events
}

func ensureResponseFallbackFixtures(t *testing.T) {
	t.Helper()
	ensureResponseFallbackDB(t)

	if err := model.DB.AutoMigrate(&model.User{}, &model.Token{}, &model.Channel{}, &model.UserRequestCost{}, &model.Log{}, &model.Trace{}); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}

	if err := model.DB.Where("id = ?", fallbackUserID).Delete(&model.User{}).Error; err != nil {
		t.Fatalf("failed to clean user fixture: %v", err)
	}
	user := &model.User{Id: fallbackUserID, Username: "response-fallback", Quota: 1_000_000, Status: model.UserStatusEnabled}
	if err := model.DB.Create(user).Error; err != nil {
		t.Fatalf("failed to create user fixture: %v", err)
	}

	if err := model.DB.Where("id = ?", fallbackTokenID).Delete(&model.Token{}).Error; err != nil {
		t.Fatalf("failed to clean token fixture: %v", err)
	}
	token := &model.Token{
		Id:             fallbackTokenID,
		UserId:         fallbackUserID,
		Key:            "fallback-token-key",
		Name:           "fallback-token",
		Status:         model.TokenStatusEnabled,
		UnlimitedQuota: true,
		RemainQuota:    0,
	}
	if err := model.DB.Create(token).Error; err != nil {
		t.Fatalf("failed to create token fixture: %v", err)
	}

	if err := model.DB.Where("id = ?", fallbackChannelID).Delete(&model.Channel{}).Error; err != nil {
		t.Fatalf("failed to clean channel fixture: %v", err)
	}
	channel := &model.Channel{Id: fallbackChannelID, Type: channeltype.Azure, Name: "azure-fallback", Status: model.ChannelStatusEnabled}
	if err := model.DB.Create(channel).Error; err != nil {
		t.Fatalf("failed to create channel fixture: %v", err)
	}

	if err := model.DB.Where("id = ?", fallbackCompatibleChannelID).Delete(&model.Channel{}).Error; err != nil {
		t.Fatalf("failed to clean openai-compatible channel fixture: %v", err)
	}
	compatibleChannel := &model.Channel{Id: fallbackCompatibleChannelID, Type: channeltype.OpenAICompatible, Name: "compatible-fallback", Status: model.ChannelStatusEnabled}
	if err := model.DB.Create(compatibleChannel).Error; err != nil {
		t.Fatalf("failed to create openai-compatible channel fixture: %v", err)
	}
}

func ensureResponseFallbackDB(t *testing.T) {
	t.Helper()
	if model.DB != nil {
		if model.LOG_DB == nil {
			model.LOG_DB = model.DB
		}
		return
	}
	db, err := gorm.Open(sqlite.Open("file:response_fallback_tests?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}
	model.DB = db
	model.LOG_DB = db
}
