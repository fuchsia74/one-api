package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	gutils "github.com/Laisky/go-utils/v5"
	glog "github.com/Laisky/go-utils/v5/log"
	"github.com/Laisky/zap"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common/config"
)

var (
	Logger       glog.Logger
	setupLogOnce sync.Once
	initLogOnce  sync.Once
)

// init initializes the logger automatically when the package is imported
func init() {
	initLogger()
}

// initLogger initializes the go-utils logger
func initLogger() {
	initLogOnce.Do(func() {
		var err error
		level := glog.LevelInfo
		if config.DebugEnabled {
			level = glog.LevelDebug
		}

		Logger, err = glog.NewConsoleWithName("one-api", level)
		if err != nil {
			panic(fmt.Sprintf("failed to create logger: %+v", err))
		}
	})
}

func SetupLogger() {
	setupLogOnce.Do(func() {
		if LogDir != "" {
			var logPath string
			if config.OnlyOneLogFile {
				logPath = filepath.Join(LogDir, "oneapi.log")
			} else {
				logPath = filepath.Join(LogDir, fmt.Sprintf("oneapi-%s.log", time.Now().Format("20060102")))
			}
			fd, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal("failed to open log file")
			}
			gin.DefaultWriter = io.MultiWriter(os.Stdout, fd)
			gin.DefaultErrorWriter = io.MultiWriter(os.Stderr, fd)
		}
	})
}

// SetupEnhancedLogger sets up the logger with alertPusher integration
func SetupEnhancedLogger(ctx context.Context) {
	opts := []zap.Option{}

	// Setup alert pusher if configured
	if config.LogPushAPI != "" {
		ratelimiter, err := gutils.NewRateLimiter(ctx, gutils.RateLimiterArgs{
			Max:     1,
			NPerSec: 1,
		})
		if err != nil {
			Logger.Panic("create ratelimiter", zap.Error(err))
		}

		alertPusher, err := glog.NewAlert(
			ctx,
			config.LogPushAPI,
			glog.WithAlertType(config.LogPushType),
			glog.WithAlertToken(config.LogPushToken),
			glog.WithAlertHookLevel(zap.ErrorLevel),
			glog.WithRateLimiter(ratelimiter),
		)
		if err != nil {
			Logger.Panic("create AlertPusher", zap.Error(err))
		}

		opts = append(opts, zap.HooksWithFields(alertPusher.GetZapHook()))
		Logger.Info("alert pusher configured",
			zap.String("alert_api", config.LogPushAPI),
			zap.String("alert_type", config.LogPushType),
		)
	}

	// Get hostname for logger context
	hostname, err := os.Hostname()
	if err != nil {
		Logger.Panic("get hostname", zap.Error(err))
	}

	// Apply options and add hostname context
	logger := Logger.WithOptions(opts...).With(
		zap.String("host", hostname),
	)
	Logger = logger

	// Set log level based on debug mode
	if config.DebugEnabled {
		_ = Logger.ChangeLevel("debug")
		Logger.Info("running in debug mode with enhanced logging")
	} else {
		_ = Logger.ChangeLevel("info")
		Logger.Info("running in production mode with enhanced logging")
	}
}
