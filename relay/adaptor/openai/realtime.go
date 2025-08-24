package openai

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/songquanpeng/one-api/common/logger"
	rmeta "github.com/songquanpeng/one-api/relay/meta"
	rmodel "github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

// RealtimeHandler proxies a WebSocket session to the upstream OpenAI Realtime endpoint.
// It preserves text/binary frames and mirrors the `Sec-WebSocket-Protocol` when present.
func RealtimeHandler(c *gin.Context, meta *rmeta.Meta) (*rmodel.ErrorWithStatusCode, *rmodel.Usage) {
	if meta.Mode != relaymode.Realtime {
		return &rmodel.ErrorWithStatusCode{
			Error: rmodel.Error{
				Message: "invalid mode for realtime handler",
				Type:    "one_api_error",
				Code:    "invalid_mode",
			},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	// Upgrade downstream connection
	upgrader := websocket.Upgrader{
		CheckOrigin:      func(r *http.Request) bool { return true },
		HandshakeTimeout: 10 * time.Second,
	}

	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return &rmodel.ErrorWithStatusCode{
			Error:      rmodel.Error{Message: "websocket upgrade failed: " + err.Error(), Type: "one_api_error", Code: "ws_upgrade_failed"},
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	// Ensure close on exit
	defer func() { _ = clientConn.Close() }()

	// Build upstream URL
	base := meta.BaseURL
	if base == "" {
		base = "https://api.openai.com" // fallback
	}
	// Preserve query but ensure model uses mapped ActualModelName
	rawQuery := c.Request.URL.RawQuery
	u, _ := url.Parse(base)
	u.Scheme = strings.Replace(u.Scheme, "http", "ws", 1) // http->ws, https->wss
	if u.Scheme == "" || u.Scheme == "http" {
		u.Scheme = "wss"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	}
	u.Path = "/v1/realtime"
	// Override model query with mapped model if provided
	q, _ := url.ParseQuery(rawQuery)
	if meta.ActualModelName != "" {
		q.Set("model", meta.ActualModelName)
	}
	u.RawQuery = q.Encode()

	// Prepare headers and subprotocols
	requestHeader := http.Header{}
	if sp := c.GetHeader("Sec-WebSocket-Protocol"); sp != "" {
		requestHeader.Set("Sec-WebSocket-Protocol", sp)
	}
	if beta := c.GetHeader("OpenAI-Beta"); beta != "" {
		requestHeader.Set("OpenAI-Beta", beta)
	} else {
		// Default beta header required by OpenAI Realtime during beta period
		requestHeader.Set("OpenAI-Beta", "realtime=v1")
	}
	requestHeader.Set("Authorization", "Bearer "+meta.APIKey)

	dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second, Proxy: http.ProxyFromEnvironment}
	upstreamConn, _, derr := dialer.Dial(u.String(), requestHeader)
	if derr != nil {
		_ = clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "upstream connect failed"))
		return &rmodel.ErrorWithStatusCode{
			Error:      rmodel.Error{Message: "upstream realtime connect failed: " + derr.Error(), Type: "one_api_error", Code: "upstream_connect_failed"},
			StatusCode: http.StatusBadGateway,
		}, nil
	}
	defer func() { _ = upstreamConn.Close() }()

	// Bi-directional pump
	errc := make(chan error, 2)
	usage := &rmodel.Usage{}
	go func() { errc <- copyWSUpstreamToClient(upstreamConn, clientConn, usage) }()
	go func() { errc <- copyWS(clientConn, upstreamConn) }()

	// Wait for either direction to error/close
	if e := <-errc; e != nil {
		logger.Logger.Debug("realtime ws closed", zap.String("error", e.Error()))
	}

	// Compute total tokens if we have parts
	if usage != nil && usage.TotalTokens == 0 {
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}

	return nil, usage
}

func copyWS(src, dst *websocket.Conn) error {
	for {
		mt, msg, err := src.ReadMessage()
		if err != nil {
			return errors.WithStack(err)
		}
		// Mirror frame type
		if werr := dst.WriteMessage(mt, msg); werr != nil {
			return errors.WithStack(werr)
		}
	}
}

// copyWSUpstreamToClient forwards frames and tries best-effort to parse usage from upstream JSON text messages.
func copyWSUpstreamToClient(src, dst *websocket.Conn, usage *rmodel.Usage) error {
	for {
		mt, msg, err := src.ReadMessage()
		if err != nil {
			return errors.WithStack(err)
		}
		if mt == websocket.TextMessage {
			maybeParseRealtimeUsage(msg, usage)
		}
		if werr := dst.WriteMessage(mt, msg); werr != nil {
			return errors.WithStack(werr)
		}
	}
}

// maybeParseRealtimeUsage attempts to extract token usage from response.done-like events.
func maybeParseRealtimeUsage(msg []byte, u *rmodel.Usage) {
	// Avoid heavy processing if no accumulator
	if u == nil || len(msg) == 0 {
		return
	}
	// Very permissive JSON parsing into generic map
	var m map[string]any
	if err := json.Unmarshal(msg, &m); err != nil {
		return
	}
	// Expect events with type and nested response.usage
	resp, _ := m["response"].(map[string]any)
	if resp == nil {
		return
	}
	usageObj, _ := resp["usage"].(map[string]any)
	if usageObj == nil {
		return
	}
	// input_tokens / output_tokens are preferred
	if v, ok := usageObj["input_tokens"].(float64); ok {
		u.PromptTokens += int(v)
	}
	if v, ok := usageObj["output_tokens"].(float64); ok {
		u.CompletionTokens += int(v)
	}
	if v, ok := usageObj["total_tokens"].(float64); ok {
		// If total provided, prefer exact
		u.TotalTokens += int(v)
	}
}
