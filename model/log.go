package model

import (
	"context"
	"fmt"
	"time"

	"github.com/Laisky/zap"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
)

type Log struct {
	Id                int    `json:"id"`
	UserId            int    `json:"user_id" gorm:"index"`
	CreatedAt         int64  `json:"created_at" gorm:"bigint;index:idx_created_at_type"`
	Type              int    `json:"type" gorm:"index:idx_created_at_type"`
	Content           string `json:"content"`
	Username          string `json:"username" gorm:"index:index_username_model_name,priority:2;default:''"`
	TokenName         string `json:"token_name" gorm:"index;default:''"`
	ModelName         string `json:"model_name" gorm:"index;index:index_username_model_name,priority:1;default:''"`
	Quota             int    `json:"quota" gorm:"default:0;index"`             // Added index for sorting
	PromptTokens      int    `json:"prompt_tokens" gorm:"default:0;index"`     // Added index for sorting
	CompletionTokens  int    `json:"completion_tokens" gorm:"default:0;index"` // Added index for sorting
	ChannelId         int    `json:"channel" gorm:"index"`
	RequestId         string `json:"request_id" gorm:"default:''"`
	TraceId           string `json:"trace_id" gorm:"type:varchar(64);index;default:''"` // TraceID from gin-middlewares
	UpdatedAt         int64  `json:"updated_at" gorm:"bigint;autoUpdateTime:milli"`
	ElapsedTime       int64  `json:"elapsed_time" gorm:"default:0;index"` // Added index for sorting (unit is ms)
	IsStream          bool   `json:"is_stream" gorm:"default:false"`
	SystemPromptReset bool   `json:"system_prompt_reset" gorm:"default:false"`
}

const (
	LogTypeUnknown = iota
	LogTypeTopup
	LogTypeConsume
	LogTypeManage
	LogTypeSystem
	LogTypeTest
)

func GetLogOrderClause(sortBy string, sortOrder string) string {
	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Map frontend field names to database column names and validate
	switch sortBy {
	case "created_time":
		return "created_at " + sortOrder
	case "prompt_tokens":
		return "prompt_tokens " + sortOrder
	case "completion_tokens":
		return "completion_tokens " + sortOrder
	case "quota":
		return "quota " + sortOrder
	case "elapsed_time":
		return "elapsed_time " + sortOrder
	default:
		return "id desc" // Default sorting
	}
}

func recordLogHelper(ctx context.Context, log *Log) {
	requestId := helper.GetRequestID(ctx)
	log.RequestId = requestId

	// Also set TraceID from gin-middlewares if available
	traceId := helper.GetTraceIDFromContext(ctx)
	log.TraceId = traceId

	// Debug logging to see what's happening
	logger.Logger.Debug("setting trace_id in log record",
		zap.String("trace_id", traceId),
		zap.String("request_id", requestId))

	err := LOG_DB.Create(log).Error
	if err != nil {
		// For billing logs (consume type), this is critical as it means we sent upstream request but failed to log it
		if log.Type == LogTypeConsume {
			logger.Logger.Error("failed to record billing log - audit trail incomplete",
				zap.Error(err),
				zap.Int("userId", log.UserId),
				zap.Int("channelId", log.ChannelId),
				zap.String("model", log.ModelName),
				zap.Int("quota", log.Quota),
				zap.String("requestId", log.RequestId),
				zap.String("note", "billing completed successfully but log recording failed"))
		} else {
			logger.Logger.Error("failed to record log", zap.Error(err))
		}
		return
	}
	logger.Logger.Info("record log",
		zap.Int("user_id", log.UserId),
		zap.String("username", log.Username),
		zap.Int64("created_at", log.CreatedAt),
		zap.Int("type", log.Type),
		zap.String("content", log.Content),
		zap.String("request_id", log.RequestId),
		zap.String("trace_id", log.TraceId))
}

func recordLogHelperWithTraceID(ctx context.Context, traceId string, log *Log) {
	requestId := helper.GetRequestID(ctx)
	log.RequestId = requestId
	log.TraceId = traceId

	// Debug logging to see what's happening
	logger.Logger.Debug("setting trace_id in log record (with explicit trace_id)",
		zap.String("trace_id", traceId),
		zap.String("request_id", requestId))

	err := LOG_DB.Create(log).Error
	if err != nil {
		// For billing logs (consume type), this is critical as it means we sent upstream request but failed to log it
		if log.Type == LogTypeConsume {
			logger.Logger.Error("failed to record billing log - audit trail incomplete",
				zap.Error(err),
				zap.Int("userId", log.UserId),
				zap.Int("channelId", log.ChannelId),
				zap.String("model", log.ModelName),
				zap.Int("quota", log.Quota),
				zap.String("requestId", log.RequestId),
				zap.String("note", "billing completed successfully but log recording failed"))
		} else {
			logger.Logger.Error("failed to record log", zap.Error(err))
		}
		return
	}
	logger.Logger.Info("record log",
		zap.Int("user_id", log.UserId),
		zap.String("username", log.Username),
		zap.Int64("created_at", log.CreatedAt),
		zap.Int("type", log.Type),
		zap.String("content", log.Content),
		zap.String("request_id", log.RequestId),
		zap.String("trace_id", log.TraceId))
}

func RecordLog(ctx context.Context, userId int, logType int, content string) {
	if logType == LogTypeConsume && !config.LogConsumeEnabled {
		return
	}
	log := &Log{
		UserId:    userId,
		Username:  GetUsernameById(userId),
		CreatedAt: helper.GetTimestamp(),
		Type:      logType,
		Content:   content,
	}
	recordLogHelper(ctx, log)
}

func RecordTopupLog(ctx context.Context, userId int, content string, quota int) {
	log := &Log{
		UserId:    userId,
		Username:  GetUsernameById(userId),
		CreatedAt: helper.GetTimestamp(),
		Type:      LogTypeTopup,
		Content:   content,
		Quota:     quota,
	}
	recordLogHelper(ctx, log)
}

func RecordConsumeLog(ctx context.Context, log *Log) {
	if !config.LogConsumeEnabled {
		return
	}
	log.Username = GetUsernameById(log.UserId)
	log.CreatedAt = helper.GetTimestamp()
	log.Type = LogTypeConsume
	recordLogHelper(ctx, log)
}

func RecordConsumeLogWithTraceID(ctx context.Context, traceId string, log *Log) {
	if !config.LogConsumeEnabled {
		return
	}
	log.Username = GetUsernameById(log.UserId)
	log.CreatedAt = helper.GetTimestamp()
	log.Type = LogTypeConsume
	recordLogHelperWithTraceID(ctx, traceId, log)
}

func RecordTestLog(ctx context.Context, log *Log) {
	log.CreatedAt = helper.GetTimestamp()
	log.Type = LogTypeTest
	recordLogHelper(ctx, log)
}

func GetAllLogs(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, startIdx int, num int, channel int, sortBy string, sortOrder string) (logs []*Log, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = LOG_DB
	} else {
		tx = LOG_DB.Where("type = ?", logType)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}

	// Apply sorting with timeout for sorting queries
	orderClause := GetLogOrderClause(sortBy, sortOrder)
	if sortBy != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = tx.WithContext(ctx).Order(orderClause).Limit(num).Offset(startIdx).Find(&logs).Error
	} else {
		err = tx.Order(orderClause).Limit(num).Offset(startIdx).Find(&logs).Error
	}
	return logs, err
}

func GetAllLogsCount(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int) (count int64, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = LOG_DB
	} else {
		tx = LOG_DB.Where("type = ?", logType)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}

	err = tx.Model(&Log{}).Count(&count).Error
	return count, err
}

func GetUserLogs(userId int, logType int, startTimestamp int64, endTimestamp int64, modelName string, tokenName string, startIdx int, num int, sortBy string, sortOrder string) (logs []*Log, err error) {
	var tx *gorm.DB
	if logType == LogTypeUnknown {
		tx = LOG_DB.Where("user_id = ?", userId)
	} else {
		tx = LOG_DB.Where("user_id = ? and type = ?", userId, logType)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}

	// Apply sorting with timeout for sorting queries
	orderClause := GetLogOrderClause(sortBy, sortOrder)
	if sortBy != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = tx.WithContext(ctx).Order(orderClause).Limit(num).Offset(startIdx).Find(&logs).Error
	} else {
		err = tx.Order(orderClause).Limit(num).Offset(startIdx).Find(&logs).Error
	}
	return logs, err
}

func SearchAllLogs(keyword string, startIdx int, num int, sortBy string, sortOrder string) (logs []*Log, total int64, err error) {
	db := LOG_DB.Model(&Log{})
	if keyword != "" {
		db = db.Where("content LIKE ?", "%"+keyword+"%")
	}
	orderClause := GetLogOrderClause(sortBy, sortOrder)
	db = db.Order(orderClause)
	err = db.Count(&total).Limit(num).Offset(startIdx).Find(&logs).Error
	return logs, total, err
}

func SearchUserLogs(userId int, keyword string, startIdx int, num int, sortBy string, sortOrder string) (logs []*Log, total int64, err error) {
	db := LOG_DB.Model(&Log{}).Where("user_id = ?", userId)
	if keyword != "" {
		db = db.Where("content LIKE ?", "%"+keyword+"%")
	}
	orderClause := GetLogOrderClause(sortBy, sortOrder)
	db = db.Order(orderClause)
	err = db.Count(&total).Limit(num).Offset(startIdx).Find(&logs).Error
	return logs, total, err
}

func SumUsedQuota(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string, channel int) (quota int64) {
	ifnull := "ifnull"
	if common.UsingPostgreSQL {
		ifnull = "COALESCE"
	}
	tx := LOG_DB.Table("logs").Select(fmt.Sprintf("%s(sum(quota),0)", ifnull))
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	tx.Where("type = ?", LogTypeConsume).Scan(&quota)
	return quota
}

func SumUsedToken(logType int, startTimestamp int64, endTimestamp int64, modelName string, username string, tokenName string) (token int) {
	ifnull := "ifnull"
	if common.UsingPostgreSQL {
		ifnull = "COALESCE"
	}
	tx := LOG_DB.Table("logs").Select(fmt.Sprintf("%s(sum(prompt_tokens),0) + %s(sum(completion_tokens),0)", ifnull, ifnull))
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if tokenName != "" {
		tx = tx.Where("token_name = ?", tokenName)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	tx.Where("type = ?", LogTypeConsume).Scan(&token)
	return token
}

func DeleteOldLog(targetTimestamp int64) (int64, error) {
	result := LOG_DB.Where("created_at < ?", targetTimestamp).Delete(&Log{})
	return result.RowsAffected, result.Error
}

// GetLogById retrieves a log entry by its ID
func GetLogById(id int) (*Log, error) {
	var log Log
	if err := LOG_DB.Where("id = ?", id).First(&log).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

type LogStatistic struct {
	Day              string `gorm:"column:day"`
	ModelName        string `gorm:"column:model_name"`
	RequestCount     int    `gorm:"column:request_count"`
	Quota            int    `gorm:"column:quota"`
	PromptTokens     int    `gorm:"column:prompt_tokens"`
	CompletionTokens int    `gorm:"column:completion_tokens"`
}

func SearchLogsByDayAndModel(userId, start, end int) (LogStatistics []*LogStatistic, err error) {
	groupSelect := "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as day"

	if common.UsingPostgreSQL {
		groupSelect = "TO_CHAR(date_trunc('day', to_timestamp(created_at)), 'YYYY-MM-DD') as day"
	}

	if common.UsingSQLite {
		groupSelect = "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as day"
	}

	// If userId is 0, query all users (site-wide statistics)
	var query string
	var args []interface{}

	if userId == 0 {
		query = `
			SELECT ` + groupSelect + `,
			model_name, count(1) as request_count,
			sum(quota) as quota,
			sum(prompt_tokens) as prompt_tokens,
			sum(completion_tokens) as completion_tokens
			FROM logs
			WHERE type=2
			AND created_at BETWEEN ? AND ?
			GROUP BY day, model_name
			ORDER BY day, model_name
		`
		args = []interface{}{start, end}
	} else {
		query = `
			SELECT ` + groupSelect + `,
			model_name, count(1) as request_count,
			sum(quota) as quota,
			sum(prompt_tokens) as prompt_tokens,
			sum(completion_tokens) as completion_tokens
			FROM logs
			WHERE type=2
			AND user_id= ?
			AND created_at BETWEEN ? AND ?
			GROUP BY day, model_name
			ORDER BY day, model_name
		`
		args = []interface{}{userId, start, end}
	}

	err = LOG_DB.Raw(query, args...).Scan(&LogStatistics).Error

	return LogStatistics, err
}
