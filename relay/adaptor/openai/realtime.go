package openai

import (
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/Laisky/errors/v2"
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/Laisky/zap"

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
        CheckOrigin: func(r *http.Request) bool { return true },
        HandshakeTimeout: 10 * time.Second,
    }

    clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return &rmodel.ErrorWithStatusCode{
            Error: rmodel.Error{Message: "websocket upgrade failed: " + err.Error(), Type: "one_api_error", Code: "ws_upgrade_failed"},
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
    // Preserve original query (model, etc.)
    rawQuery := c.Request.URL.RawQuery
    u, _ := url.Parse(base)
    u.Scheme = strings.Replace(u.Scheme, "http", "ws", 1) // http->ws, https->wss
    if u.Scheme == "" || u.Scheme == "http" {
        u.Scheme = "wss"
    } else if u.Scheme == "https" {
        u.Scheme = "wss"
    }
    u.Path = "/v1/realtime"
    u.RawQuery = rawQuery

    // Prepare headers and subprotocols
    requestHeader := http.Header{}
    if sp := c.GetHeader("Sec-WebSocket-Protocol"); sp != "" {
        requestHeader.Set("Sec-WebSocket-Protocol", sp)
    }
    if beta := c.GetHeader("OpenAI-Beta"); beta != "" {
        requestHeader.Set("OpenAI-Beta", beta)
    }
    requestHeader.Set("Authorization", "Bearer "+meta.APIKey)

    dialer := websocket.Dialer{HandshakeTimeout: 10 * time.Second, Proxy: http.ProxyFromEnvironment}
    upstreamConn, _, derr := dialer.Dial(u.String(), requestHeader)
    if derr != nil {
        _ = clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "upstream connect failed"))
        return &rmodel.ErrorWithStatusCode{
            Error: rmodel.Error{Message: "upstream realtime connect failed: " + derr.Error(), Type: "one_api_error", Code: "upstream_connect_failed"},
            StatusCode: http.StatusBadGateway,
        }, nil
    }
    defer func() { _ = upstreamConn.Close() }()

    // Bi-directional pump
    errc := make(chan error, 2)
    go func() { errc <- copyWS(upstreamConn, clientConn) }() // upstream -> client
    go func() { errc <- copyWS(clientConn, upstreamConn) }() // client -> upstream

    // Wait for either direction to error/close
    if e := <-errc; e != nil {
        logger.Logger.Debug("realtime ws closed", zap.String("error", e.Error()))
    }
    return nil, &rmodel.Usage{}
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
