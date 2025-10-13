package logger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	errors "github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"
)

// StartLogRetentionCleaner launches a background worker that deletes log files older than the
// configured retention period. The cleanup runs immediately and then once every 24 hours until
// the provided context is cancelled. The ctx parameter controls the lifecycle, retentionDays sets
// the age threshold in days, and logDir provides the directory containing log files.
func StartLogRetentionCleaner(ctx context.Context, retentionDays int, logDir string) {
	if retentionDays <= 0 {
		Logger.Debug("log retention disabled", zap.Int("log_retention_days", retentionDays))
		return
	}

	if strings.TrimSpace(logDir) == "" {
		Logger.Warn("log retention enabled but log directory is empty", zap.Int("log_retention_days", retentionDays))
		return
	}

	cleanup := func() {
		if err := deleteExpiredLogFiles(retentionDays, logDir); err != nil {
			Logger.Warn("log retention cleanup failed", zap.Error(err))
		}
	}

	cleanup()

	ticker := time.NewTicker(24 * time.Hour)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				Logger.Info("log retention cleaner stopped", zap.Error(ctx.Err()))
				return
			case <-ticker.C:
				cleanup()
			}
		}
	}()

	Logger.Info("log retention cleaner started", zap.Int("log_retention_days", retentionDays), zap.String("log_dir", logDir))
}

// deleteExpiredLogFiles removes log files older than the retention window from the configured log directory.
// The retentionDays parameter defines the age threshold in days, logDir is the directory to scan,
// and the returned error reports failures when listing entries.
func deleteExpiredLogFiles(retentionDays int, logDir string) error {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.Wrap(err, "read log directory")
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".log") {
			continue
		}

		info, infoErr := entry.Info()
		if infoErr != nil {
			Logger.Warn("skip log file without metadata", zap.String("log_path", filepath.Join(logDir, name)), zap.Error(infoErr))
			continue
		}

		modTime := info.ModTime().UTC()
		if !modTime.Before(cutoff) {
			continue
		}

		fullPath := filepath.Join(logDir, name)
		if removeErr := os.Remove(fullPath); removeErr != nil {
			Logger.Warn("failed to delete expired log file", zap.String("log_path", fullPath), zap.Error(removeErr))
			continue
		}

		Logger.Info("deleted expired log file", zap.String("log_path", fullPath), zap.Time("modified_at", modTime))
	}

	return nil
}
