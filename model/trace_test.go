package model

import (
	"context"
	"strings"
	"testing"

	gmw "github.com/Laisky/gin-middlewares/v6"
	"github.com/stretchr/testify/require"

	"github.com/songquanpeng/one-api/common/logger"
)

// TestCreateTraceWithLongURL verifies that trace creation succeeds even when the request URL includes very long query strings.
func TestCreateTraceWithLongURL(t *testing.T) {
	setupTestDatabase(t)

	require.NoError(t, DB.Exec("DELETE FROM traces WHERE trace_id LIKE 'test-trace-long-url%'").Error)

	longURL := "/api/verification?token=" + strings.Repeat("abc123", 1000)
	require.Greater(t, len(longURL), maxTraceURLLength)

	ctx := gmw.SetLogger(context.Background(), logger.Logger)

	traceID := "test-trace-long-url"
	trace, err := CreateTrace(ctx, traceID, longURL, "GET", 0)
	require.NoError(t, err)
	require.NotNil(t, trace)
	require.Equal(t, maxTraceURLLength, len(trace.URL))

	var stored Trace
	err = DB.Where("trace_id = ?", traceID).First(&stored).Error
	require.NoError(t, err)
	require.Equal(t, maxTraceURLLength, len(stored.URL))
	require.Equal(t, trace.URL, stored.URL)
}

func TestCreateTraceURLWithinLimit(t *testing.T) {
	setupTestDatabase(t)
	require.NoError(t, DB.Exec("DELETE FROM traces WHERE trace_id = 'test-trace-within-limit'").Error)

	url := "/api/status"
	require.LessOrEqual(t, len(url), maxTraceURLLength)
	ctx := gmw.SetLogger(context.Background(), logger.Logger)
	trace, err := CreateTrace(ctx, "test-trace-within-limit", url, "GET", 0)
	require.NoError(t, err)
	require.Equal(t, url, trace.URL)

	var stored Trace
	err = DB.Where("trace_id = ?", "test-trace-within-limit").First(&stored).Error
	require.NoError(t, err)
	require.Equal(t, url, stored.URL)
}
