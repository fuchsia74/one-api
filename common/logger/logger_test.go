package logger

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common/config"
)

func TestSetupEnhancedLogger(t *testing.T) {
	ctx := context.Background()

	// Test without alert pusher configuration
	t.Run("without_alert_pusher", func(t *testing.T) {
		// Ensure no alert pusher config
		config.LogPushAPI = ""
		config.LogPushType = ""
		config.LogPushToken = ""

		// This should not panic and should work normally
		SetupEnhancedLogger(ctx)

		// Test that logger is working
		Logger.Info("test log message without alert pusher")
	})

	// Test with alert pusher configuration (but invalid URL to avoid actual network calls)
	t.Run("with_alert_pusher_config", func(t *testing.T) {
		// Set alert pusher config
		config.LogPushAPI = "http://invalid-test-url.example.com/api/push"
		config.LogPushType = "test"
		config.LogPushToken = "test-token"

		// This should not panic even with invalid URL during setup
		SetupEnhancedLogger(ctx)

		// Test that logger is working
		Logger.Info("test log message with alert pusher config")
	})
}

func TestSetupEnhancedLoggerWithEnvironmentVariables(t *testing.T) {
	ctx := context.Background()

	// Test with environment variables
	t.Run("with_env_vars", func(t *testing.T) {
		// Set environment variables
		os.Setenv("LOG_PUSH_API", "http://test-api.example.com/push")
		os.Setenv("LOG_PUSH_TYPE", "webhook")
		os.Setenv("LOG_PUSH_TOKEN", "test-env-token")

		// Reload config to pick up env vars
		config.LogPushAPI = os.Getenv("LOG_PUSH_API")
		config.LogPushType = os.Getenv("LOG_PUSH_TYPE")
		config.LogPushToken = os.Getenv("LOG_PUSH_TOKEN")

		// This should not panic
		SetupEnhancedLogger(ctx)

		// Test that logger is working
		Logger.Info("test log message with environment variables")

		// Clean up
		os.Unsetenv("LOG_PUSH_API")
		os.Unsetenv("LOG_PUSH_TYPE")
		os.Unsetenv("LOG_PUSH_TOKEN")
	})
}

func TestLoggerErrorLevelWithAlertPusher(t *testing.T) {
	ctx := context.Background()

	// Test that error level logs would trigger alert pusher (if configured)
	t.Run("error_level_logging", func(t *testing.T) {
		// Set up with mock alert pusher config
		config.LogPushAPI = "http://mock-alert-api.example.com/push"
		config.LogPushType = "mock"
		config.LogPushToken = "mock-token"

		SetupEnhancedLogger(ctx)

		// Test error level logging (this would trigger alert pusher if URL was valid)
		Logger.Error("test error message for alert pusher",
			zap.String("component", "test"),
			zap.String("error_type", "test_error"))

		// Give a small delay to allow any async processing
		time.Sleep(100 * time.Millisecond)
	})
}

func TestLoggerDebugMode(t *testing.T) {
	ctx := context.Background()

	t.Run("debug_mode_enabled", func(t *testing.T) {
		// Enable debug mode
		originalDebugEnabled := config.DebugEnabled
		config.DebugEnabled = true

		SetupEnhancedLogger(ctx)

		// Test debug logging
		Logger.Debug("test debug message")
		Logger.Info("test info message in debug mode")

		// Restore original setting
		config.DebugEnabled = originalDebugEnabled
	})

	t.Run("debug_mode_disabled", func(t *testing.T) {
		// Disable debug mode
		originalDebugEnabled := config.DebugEnabled
		config.DebugEnabled = false

		SetupEnhancedLogger(ctx)

		// Test logging in production mode
		Logger.Info("test info message in production mode")

		// Restore original setting
		config.DebugEnabled = originalDebugEnabled
	})
}
