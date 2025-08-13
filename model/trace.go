package model

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common/logger"
)

// Trace represents a request tracing record with key timestamps
type Trace struct {
	Id         int    `json:"id" gorm:"primaryKey;autoIncrement"`
	TraceId    string `json:"trace_id" gorm:"type:varchar(64);uniqueIndex;not null"` // TraceID from gin-middlewares
	URL        string `json:"url" gorm:"type:varchar(512);not null"`                 // Request URL
	Method     string `json:"method" gorm:"type:varchar(16);not null"`               // HTTP method
	BodySize   int64  `json:"body_size" gorm:"bigint;default:0"`                     // Request body size in bytes
	Status     int    `json:"status" gorm:"default:0"`                               // HTTP status code
	Timestamps string `json:"timestamps" gorm:"type:text"`                           // JSON object with timestamps
	CreatedAt  int64  `json:"created_at" gorm:"bigint;autoCreateTime:milli;index"`
	UpdatedAt  int64  `json:"updated_at" gorm:"bigint;autoUpdateTime:milli"`
}

// TraceTimestamps represents the structure of timestamps stored in the Trace.Timestamps field
type TraceTimestamps struct {
	RequestReceived       *int64 `json:"request_received,omitempty"`        // When request was received
	RequestForwarded      *int64 `json:"request_forwarded,omitempty"`       // When request was forwarded to upstream
	FirstUpstreamResponse *int64 `json:"first_upstream_response,omitempty"` // When first response received from upstream
	FirstClientResponse   *int64 `json:"first_client_response,omitempty"`   // When first response sent to client
	UpstreamCompleted     *int64 `json:"upstream_completed,omitempty"`      // When upstream response completed (for streaming)
	RequestCompleted      *int64 `json:"request_completed,omitempty"`       // When entire request completed
}

// Timestamp constants for consistent key naming
const (
	TimestampRequestReceived       = "request_received"
	TimestampRequestForwarded      = "request_forwarded"
	TimestampFirstUpstreamResponse = "first_upstream_response"
	TimestampFirstClientResponse   = "first_client_response"
	TimestampUpstreamCompleted     = "upstream_completed"
	TimestampRequestCompleted      = "request_completed"
)

// CreateTrace creates a new trace record with initial data
func CreateTrace(ctx context.Context, traceId, url, method string, bodySize int64) (*Trace, error) {
	now := time.Now().UnixMilli()

	timestamps := &TraceTimestamps{
		RequestReceived: &now,
	}

	timestampsJSON, err := json.Marshal(timestamps)
	if err != nil {
		logger.Logger.Error("failed to marshal trace timestamps",
			zap.Error(err),
			zap.String("trace_id", traceId))
		return nil, errors.Wrapf(err, "failed to marshal trace timestamps for trace_id: %s", traceId)
	}

	trace := &Trace{
		TraceId:    traceId,
		URL:        url,
		Method:     method,
		BodySize:   bodySize,
		Timestamps: string(timestampsJSON),
	}

	if err := DB.Create(trace).Error; err != nil {
		logger.Logger.Error("failed to create trace record",
			zap.Error(err),
			zap.String("trace_id", traceId))
		return nil, errors.Wrapf(err, "failed to create trace record for trace_id: %s", traceId)
	}

	logger.Logger.Debug("created trace record",
		zap.String("trace_id", traceId),
		zap.String("url", url),
		zap.String("method", method))

	return trace, nil
}

// UpdateTraceTimestamp updates a specific timestamp in the trace record
func UpdateTraceTimestamp(ctx context.Context, traceId, timestampKey string) error {
	var trace Trace
	if err := DB.Where("trace_id = ?", traceId).First(&trace).Error; err != nil {
		logger.Logger.Warn("trace record not found for timestamp update",
			zap.String("trace_id", traceId),
			zap.String("timestamp_key", timestampKey),
			zap.Error(err))
		return errors.Wrapf(err, "trace record not found for timestamp update, trace_id: %s, key: %s", traceId, timestampKey)
	}

	var timestamps TraceTimestamps
	if err := json.Unmarshal([]byte(trace.Timestamps), &timestamps); err != nil {
		logger.Logger.Error("failed to unmarshal trace timestamps",
			zap.Error(err),
			zap.String("trace_id", traceId))
		return errors.Wrapf(err, "failed to unmarshal trace timestamps for trace_id: %s", traceId)
	}

	now := time.Now().UnixMilli()

	// Update the specific timestamp
	switch timestampKey {
	case TimestampRequestForwarded:
		timestamps.RequestForwarded = &now
	case TimestampFirstUpstreamResponse:
		timestamps.FirstUpstreamResponse = &now
	case TimestampFirstClientResponse:
		timestamps.FirstClientResponse = &now
	case TimestampUpstreamCompleted:
		timestamps.UpstreamCompleted = &now
	case TimestampRequestCompleted:
		timestamps.RequestCompleted = &now
	default:
		logger.Logger.Warn("unknown timestamp key",
			zap.String("trace_id", traceId),
			zap.String("timestamp_key", timestampKey))
		return nil
	}

	timestampsJSON, err := json.Marshal(timestamps)
	if err != nil {
		logger.Logger.Error("failed to marshal updated trace timestamps",
			zap.Error(err),
			zap.String("trace_id", traceId))
		return errors.Wrapf(err, "failed to marshal updated trace timestamps for trace_id: %s", traceId)
	}

	if err := DB.Model(&trace).Update("timestamps", string(timestampsJSON)).Error; err != nil {
		logger.Logger.Error("failed to update trace timestamp",
			zap.Error(err),
			zap.String("trace_id", traceId),
			zap.String("timestamp_key", timestampKey))
		return errors.Wrapf(err, "failed to update trace timestamp for trace_id: %s, key: %s", traceId, timestampKey)
	}

	logger.Logger.Debug("updated trace timestamp",
		zap.String("trace_id", traceId),
		zap.String("timestamp_key", timestampKey))

	return nil
}

// UpdateTraceStatus updates the HTTP status code for a trace
func UpdateTraceStatus(ctx context.Context, traceId string, status int) error {
	if err := DB.Model(&Trace{}).Where("trace_id = ?", traceId).Update("status", status).Error; err != nil {
		logger.Logger.Error("failed to update trace status",
			zap.Error(err),
			zap.String("trace_id", traceId),
			zap.Int("status", status))
		return errors.Wrapf(err, "failed to update trace status for trace_id: %s", traceId)
	}

	logger.Logger.Debug("updated trace status",
		zap.String("trace_id", traceId),
		zap.Int("status", status))

	return nil
}

// GetTraceByTraceId retrieves a trace record by trace ID
func GetTraceByTraceId(traceId string) (*Trace, error) {
	var trace Trace
	if err := DB.Where("trace_id = ?", traceId).First(&trace).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get trace by trace_id: %s", traceId)
	}
	return &trace, nil
}

// GetTraceTimestamps parses and returns the timestamps from a trace record
func (t *Trace) GetTraceTimestamps() (*TraceTimestamps, error) {
	var timestamps TraceTimestamps
	if err := json.Unmarshal([]byte(t.Timestamps), &timestamps); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal trace timestamps for trace_id: %s", t.TraceId)
	}
	return &timestamps, nil
}
