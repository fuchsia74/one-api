package controller

import (
	"time"

	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
)

// RelayRealtime handles WebSocket Realtime proxying for OpenAI Realtime API.
// It does not perform request conversion or pre-consume quota; billing is best-effort based on upstream usage if emitted.
func RelayRealtime(c *gin.Context) {
	ctx := gmw.Ctx(c)
	_ = ctx // reserved for future use (timeouts, logging)
	start := time.Now()
	relayMeta := meta.GetByContext(c)

	// Record channel requests in flight
	PrometheusMonitor.RecordChannelRequest(relayMeta, start)

	if bizErr, _ := openai.RealtimeHandler(c, relayMeta); bizErr != nil {
		// On handshake/connection error, return JSON error (no WS established)
		c.JSON(bizErr.StatusCode, gin.H{"error": bizErr.Error})
		PrometheusMonitor.RecordRelayRequest(c, relayMeta, start, false, 0, 0, 0)
		return
	}

	// If we reach here, the WS session completed normally (handler handled I/O).
	logger.Logger.Debug("realtime session closed")
	PrometheusMonitor.RecordRelayRequest(c, relayMeta, start, true, 0, 0, 0)
}
