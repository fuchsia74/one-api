package internal

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

// BatchJob represents a batch processing job
type BatchJob struct {
	TableInfo TableInfo
	Offset    int64
	Limit     int
	JobID     int
}

// BatchResult represents the result of a batch processing job
type BatchResult struct {
	JobID       int
	RecordCount int64
	Error       error
}

// TableMigrationOrder defines the order in which tables should be migrated
// to respect foreign key constraints
var TableMigrationOrder = []TableInfo{
	{"users", &model.User{}},
	{"options", &model.Option{}},
	{"tokens", &model.Token{}},
	{"channels", &model.Channel{}},
	{"redemptions", &model.Redemption{}},
	{"abilities", &model.Ability{}},
	{"logs", &model.Log{}},
	{"user_request_costs", &model.UserRequestCost{}},
}

// TableInfo holds information about a table and its corresponding model
type TableInfo struct {
	Name  string
	Model interface{}
}

// migrateData performs the actual data migration
func (m *Migrator) migrateData(ctx context.Context, stats *MigrationStats) error {
	logger.Logger.Info("Starting data migration...")

	for _, tableInfo := range TableMigrationOrder {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := m.migrateTable(ctx, tableInfo, stats); err != nil {
			stats.Errors = append(stats.Errors, fmt.Errorf("failed to migrate table %s: %w", tableInfo.Name, err))
			logger.Logger.Error("Failed to migrate table",
				zap.String("table", tableInfo.Name),
				zap.Error(err))
			continue
		}

		stats.TablesDone++
		logger.Logger.Info(fmt.Sprintf("Successfully migrated table %s (%d/%d)", tableInfo.Name, stats.TablesDone, stats.TablesTotal))
	}

	logger.Logger.Info("Data migration completed")
	return nil
}

// migrateTable migrates data for a specific table using concurrent workers
func (m *Migrator) migrateTable(ctx context.Context, tableInfo TableInfo, stats *MigrationStats) error {
	// Check if table exists in source
	exists, err := m.sourceConn.TableExists(tableInfo.Name)
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %w", err)
	}
	if !exists {
		logger.Logger.Warn(fmt.Sprintf("Table %s does not exist in source database, skipping", tableInfo.Name))
		return nil
	}

	// Get total count for progress tracking
	totalCount, err := m.sourceConn.GetRowCount(tableInfo.Name)
	if err != nil {
		return fmt.Errorf("failed to get row count: %w", err)
	}

	if totalCount == 0 {
		logger.Logger.Info(fmt.Sprintf("Table %s is empty, skipping", tableInfo.Name))
		return nil
	}

	logger.Logger.Info(fmt.Sprintf("Migrating table %s (%d records) with %d workers, batch size %d",
		tableInfo.Name, totalCount, m.Workers, m.BatchSize))

	// Use concurrent processing for better performance
	if m.Workers > 1 {
		return m.migrateTableConcurrent(ctx, tableInfo, totalCount, stats)
	} else {
		return m.migrateTableSequential(ctx, tableInfo, totalCount, stats)
	}
}

// migrateTableSequential migrates data sequentially (single-threaded)
func (m *Migrator) migrateTableSequential(ctx context.Context, tableInfo TableInfo, totalCount int64, stats *MigrationStats) error {
	var offset int64 = 0
	var migratedCount int64 = 0
	var lastProgressReport int64 = 0

	for offset < totalCount {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		batchCount, err := m.migrateBatch(tableInfo, offset, m.BatchSize)
		if err != nil {
			return fmt.Errorf("failed to migrate batch at offset %d: %w", offset, err)
		}

		migratedCount += batchCount
		offset += int64(m.BatchSize)
		atomic.AddInt64(&stats.RecordsDone, batchCount)

		// Show progress every 10% or every 10,000 records, whichever is less frequent
		progressThreshold := totalCount / 10
		if progressThreshold < 10000 {
			progressThreshold = 10000
		}

		if m.Verbose && (migratedCount-lastProgressReport >= progressThreshold || migratedCount == totalCount) {
			progress := float64(migratedCount) / float64(totalCount) * 100
			logger.Logger.Info(fmt.Sprintf("Table %s: %d/%d records (%.1f%%)", tableInfo.Name, migratedCount, totalCount, progress))
			lastProgressReport = migratedCount
		}

		// Break if we've processed all records
		if batchCount < int64(m.BatchSize) {
			break
		}
	}

	logger.Logger.Info(fmt.Sprintf("Table %s migration completed: %d records", tableInfo.Name, migratedCount))
	return nil
}

// migrateTableConcurrent migrates data using concurrent workers
func (m *Migrator) migrateTableConcurrent(ctx context.Context, tableInfo TableInfo, totalCount int64, stats *MigrationStats) error {
	// Create job and result channels
	jobs := make(chan BatchJob, m.Workers*2)
	results := make(chan BatchResult, m.Workers*2)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < m.Workers; i++ {
		wg.Add(1)
		go m.batchWorker(ctx, jobs, results, &wg)
	}

	// Start result collector
	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	var migratedCount int64
	var lastProgressReport int64
	go func() {
		defer collectorWg.Done()
		for result := range results {
			if result.Error != nil {
				logger.Logger.Error("Batch job failed",
					zap.Int("job_id", result.JobID),
					zap.Error(result.Error))
				continue
			}

			atomic.AddInt64(&migratedCount, result.RecordCount)
			atomic.AddInt64(&stats.RecordsDone, result.RecordCount)

			// Show progress
			currentCount := atomic.LoadInt64(&migratedCount)
			progressThreshold := totalCount / 10
			if progressThreshold < 10000 {
				progressThreshold = 10000
			}

			if m.Verbose && (currentCount-lastProgressReport >= progressThreshold || currentCount >= totalCount) {
				progress := float64(currentCount) / float64(totalCount) * 100
				logger.Logger.Info(fmt.Sprintf("Table %s: %d/%d records (%.1f%%)", tableInfo.Name, currentCount, totalCount, progress))
				lastProgressReport = currentCount
			}
		}
	}()

	// Generate jobs
	jobID := 0
	for offset := int64(0); offset < totalCount; offset += int64(m.BatchSize) {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			close(results)
			collectorWg.Wait()
			return ctx.Err()
		case jobs <- BatchJob{
			TableInfo: tableInfo,
			Offset:    offset,
			Limit:     m.BatchSize,
			JobID:     jobID,
		}:
			jobID++
		}
	}

	// Close jobs channel and wait for workers to finish
	close(jobs)
	wg.Wait()
	close(results)
	collectorWg.Wait()

	finalCount := atomic.LoadInt64(&migratedCount)
	logger.Logger.Info(fmt.Sprintf("Table %s migration completed: %d records", tableInfo.Name, finalCount))
	return nil
}

// batchWorker processes batch jobs concurrently
func (m *Migrator) batchWorker(ctx context.Context, jobs <-chan BatchJob, results chan<- BatchResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
		}

		count, err := m.migrateBatch(job.TableInfo, job.Offset, job.Limit)
		results <- BatchResult{
			JobID:       job.JobID,
			RecordCount: count,
			Error:       err,
		}
	}
}

// migrateBatch migrates a batch of records for a specific table
func (m *Migrator) migrateBatch(tableInfo TableInfo, offset int64, limit int) (int64, error) {
	// Create a slice to hold the batch data
	modelType := reflect.TypeOf(tableInfo.Model).Elem()
	sliceType := reflect.SliceOf(modelType)
	batch := reflect.New(sliceType).Interface()

	// Fetch batch from source database
	query := m.sourceConn.DB.Limit(limit).Offset(int(offset))
	if err := query.Find(batch).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch batch from source: %w", err)
	}

	// Get the actual slice value
	batchValue := reflect.ValueOf(batch).Elem()
	batchLen := batchValue.Len()

	if batchLen == 0 {
		return 0, nil
	}

	// Skip insertion in dry run mode
	if m.DryRun {
		return int64(batchLen), nil
	}

	// Insert batch into target database with conflict resolution
	if err := m.insertBatchWithConflictResolution(batch, tableInfo); err != nil {
		return 0, fmt.Errorf("failed to insert batch into target: %w", err)
	}

	return int64(batchLen), nil
}

// insertBatchWithConflictResolution inserts a batch with conflict resolution
func (m *Migrator) insertBatchWithConflictResolution(batch interface{}, tableInfo TableInfo) error {
	// First try a simple insert for better performance
	err := m.targetConn.DB.Create(batch).Error
	if err == nil {
		return nil // Success - no conflicts
	}

	// If we get a conflict error, use upsert approach
	if m.isConflictError(err) {
		if m.Verbose {
			logger.Logger.Info(fmt.Sprintf("Conflict detected in table %s, switching to upsert mode", tableInfo.Name))
		}
		return m.upsertBatch(batch, tableInfo)
	}

	// For non-conflict errors, return the original error
	return err
}

// isConflictError checks if the error is a primary key or unique constraint violation
func (m *Migrator) isConflictError(err error) bool {
	errStr := err.Error()
	// Check for common conflict error patterns across different databases
	conflictPatterns := []string{
		// PostgreSQL
		"duplicate key value violates unique constraint",
		"violates unique constraint",
		// SQLite
		"UNIQUE constraint failed",
		"constraint failed: UNIQUE",
		// MySQL
		"Duplicate entry",
		"duplicate key",
		"Duplicate key name",
		// General patterns
		"already exists",
		"constraint violation",
	}

	for _, pattern := range conflictPatterns {
		if contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// upsertBatch performs upsert operation for conflicting records
func (m *Migrator) upsertBatch(batch interface{}, tableInfo TableInfo) error {
	// Get the slice value
	batchValue := reflect.ValueOf(batch).Elem()
	batchLen := batchValue.Len()

	successCount := 0
	errorCount := 0

	// Process each record individually for upsert
	for i := 0; i < batchLen; i++ {
		record := batchValue.Index(i).Addr().Interface()

		// Use GORM's Save method which performs INSERT or UPDATE
		result := m.targetConn.DB.Save(record)
		if result.Error != nil {
			errorCount++
			if m.Verbose {
				logger.Logger.Warn(fmt.Sprintf("Failed to upsert record %d in table %s: %v", i+1, tableInfo.Name, result.Error))
			}
			// Continue with other records instead of failing the entire batch
		} else {
			successCount++
		}
	}

	if m.Verbose && errorCount > 0 {
		logger.Logger.Warn("Table upsert completed with errors",
			zap.String("table", tableInfo.Name),
			zap.Int("successful", successCount),
			zap.Int("failed", errorCount))
	}

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr)))
}

// containsSubstring is a helper function for substring checking
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// validateResults validates the migration results by comparing record counts
func (m *Migrator) validateResults(stats *MigrationStats) error {
	logger.Logger.Info("Validating migration results...")

	var validationErrors []error

	for _, tableInfo := range TableMigrationOrder {
		// Check if table exists in source
		sourceExists, err := m.sourceConn.TableExists(tableInfo.Name)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("failed to check source table %s: %w", tableInfo.Name, err))
			continue
		}

		if !sourceExists {
			continue // Skip tables that don't exist in source
		}

		// Get source count
		sourceCount, err := m.sourceConn.GetRowCount(tableInfo.Name)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("failed to get source count for %s: %w", tableInfo.Name, err))
			continue
		}

		// Get target count
		targetCount, err := m.targetConn.GetRowCount(tableInfo.Name)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("failed to get target count for %s: %w", tableInfo.Name, err))
			continue
		}

		// Compare counts
		if sourceCount != targetCount {
			validationErrors = append(validationErrors, fmt.Errorf("record count mismatch for table %s: source=%d, target=%d", tableInfo.Name, sourceCount, targetCount))
		} else {
			if m.Verbose {
				logger.Logger.Info(fmt.Sprintf("Table %s validation passed: %d records", tableInfo.Name, sourceCount))
			}
		}
	}

	if len(validationErrors) > 0 {
		logger.Logger.Error("Migration validation failed:")
		for _, err := range validationErrors {
			logger.Logger.Error("Migration validation error", zap.Error(err))
		}
		return fmt.Errorf("migration validation failed with %d errors", len(validationErrors))
	}

	logger.Logger.Info("Migration validation completed successfully")
	return nil
}

// ExportData exports data from source database to a structured format
func (m *Migrator) ExportData(ctx context.Context) (map[string]interface{}, error) {
	logger.Logger.Info("Exporting data from source database...")

	exportData := make(map[string]interface{})

	for _, tableInfo := range TableMigrationOrder {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Check if table exists
		exists, err := m.sourceConn.TableExists(tableInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check if table %s exists: %w", tableInfo.Name, err)
		}

		if !exists {
			logger.Logger.Warn(fmt.Sprintf("Table %s does not exist in source database, skipping", tableInfo.Name))
			continue
		}

		// Export table data
		tableData, err := m.exportTable(tableInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to export table %s: %w", tableInfo.Name, err)
		}

		exportData[tableInfo.Name] = tableData
		logger.Logger.Info(fmt.Sprintf("Exported table %s", tableInfo.Name))
	}

	logger.Logger.Info("Data export completed")
	return exportData, nil
}

// exportTable exports all data from a specific table
func (m *Migrator) exportTable(tableInfo TableInfo) (interface{}, error) {
	// Create a slice to hold all table data
	modelType := reflect.TypeOf(tableInfo.Model).Elem()
	sliceType := reflect.SliceOf(modelType)
	tableData := reflect.New(sliceType).Interface()

	// Fetch all data from the table
	if err := m.sourceConn.DB.Find(tableData).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch data from table: %w", err)
	}

	return tableData, nil
}
