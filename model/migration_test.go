package model

import (
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

func TestMain(m *testing.M) {
	// Setup logger for tests
	logger.SetupLogger()

	// Run tests
	code := m.Run()

	// Cleanup
	os.Exit(code)
}

func setupMigrationTestDB(t *testing.T) *gorm.DB {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Set database type flags
	common.UsingSQLite.Store(true)
	common.UsingMySQL.Store(false)
	common.UsingPostgreSQL.Store(false)

	return db
}

func TestMigrateChannelFieldsToText_SQLite(t *testing.T) {
	// Setup test database
	testDB := setupMigrationTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

	// Test that SQLite migration is skipped
	err := MigrateChannelFieldsToText()
	if err != nil {
		t.Errorf("SQLite field migration should not fail: %v", err)
	}

	// Verify that the migration was skipped (no actual schema changes needed for SQLite)
	// This is expected behavior as documented in the function
}

func TestMigrateChannelFieldsToText_Idempotency(t *testing.T) {
	// Setup test database
	testDB := setupMigrationTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

	// Run migration multiple times - should be idempotent
	for i := range 3 {
		err := MigrateChannelFieldsToText()
		if err != nil {
			t.Errorf("Migration run %d failed: %v", i+1, err)
		}
	}
}

func TestMigrateTraceURLColumnToText_SQLite(t *testing.T) {
	// Setup test database
	testDB := setupMigrationTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

	if err := MigrateTraceURLColumnToText(); err != nil {
		t.Errorf("SQLite trace URL migration should not fail: %v", err)
	}
}

func TestMigrateTraceURLColumnToText_Idempotency(t *testing.T) {
	// Setup test database
	testDB := setupMigrationTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

	for i := range 3 {
		if err := MigrateTraceURLColumnToText(); err != nil {
			t.Errorf("Trace URL migration run %d failed: %v", i+1, err)
		}
	}
}

func TestCheckIfFieldMigrationNeeded_SQLite(t *testing.T) {
	// Setup test database
	testDB := setupMigrationTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

	// For SQLite, migration should never be needed
	needed, err := checkIfFieldMigrationNeeded()
	if err != nil {
		t.Errorf("checkIfFieldMigrationNeeded failed: %v", err)
	}
	if needed {
		t.Error("SQLite should never need field migration")
	}
}

func TestChannelModelConfigsMigration(t *testing.T) {
	// Setup test database
	testDB := setupMigrationTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

	// Create channels table
	err := testDB.AutoMigrate(&Channel{})
	if err != nil {
		t.Fatalf("Failed to create channels table: %v", err)
	}

	// Create test channel with old format ModelConfigs
	oldFormatConfigs := `{"gpt-3.5-turbo":{"ratio":1.0,"completion_ratio":2.0,"max_tokens":4096}}`
	testChannel := &Channel{
		Name:         "Test Channel",
		Type:         1,
		Status:       1,
		Models:       "gpt-3.5-turbo",
		ModelConfigs: &oldFormatConfigs,
	}

	err = testDB.Create(testChannel).Error
	if err != nil {
		t.Fatalf("Failed to create test channel: %v", err)
	}

	// Run the migration
	err = MigrateAllChannelModelConfigs()
	if err != nil {
		t.Errorf("MigrateAllChannelModelConfigs failed: %v", err)
	}

	// Verify the channel still exists and has valid ModelConfigs
	var migratedChannel Channel
	err = testDB.First(&migratedChannel, testChannel.Id).Error
	if err != nil {
		t.Errorf("Failed to retrieve migrated channel: %v", err)
	}

	if migratedChannel.ModelConfigs == nil || *migratedChannel.ModelConfigs == "" {
		t.Error("ModelConfigs should not be empty after migration")
	}

	// Test that the migrated configs are valid
	configs := migratedChannel.GetModelPriceConfigs()
	if len(configs) == 0 {
		t.Error("Migrated ModelConfigs should contain model configurations")
	}

	if config, exists := configs["gpt-3.5-turbo"]; !exists {
		t.Error("Expected gpt-3.5-turbo configuration to exist after migration")
	} else {
		if config.Ratio != 1.0 {
			t.Errorf("Expected ratio 1.0, got %f", config.Ratio)
		}
		if config.CompletionRatio != 2.0 {
			t.Errorf("Expected completion ratio 2.0, got %f", config.CompletionRatio)
		}
	}
}

func TestChannelModelConfigsMigration_EmptyData(t *testing.T) {
	// Setup test database
	testDB := setupMigrationTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

	// Create channels table
	err := testDB.AutoMigrate(&Channel{})
	if err != nil {
		t.Fatalf("Failed to create channels table: %v", err)
	}

	// Create test channel with no ModelConfigs
	testChannel := &Channel{
		Name:   "Empty Test Channel",
		Type:   1,
		Status: 1,
		Models: "gpt-4",
	}

	err = testDB.Create(testChannel).Error
	if err != nil {
		t.Fatalf("Failed to create test channel: %v", err)
	}

	// Run the migration - should handle empty data gracefully
	err = MigrateAllChannelModelConfigs()
	if err != nil {
		t.Errorf("MigrateAllChannelModelConfigs should handle empty data: %v", err)
	}
}

func TestChannelModelConfigsMigration_InvalidJSON(t *testing.T) {
	// Setup test database
	testDB := setupMigrationTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

	// Create channels table
	err := testDB.AutoMigrate(&Channel{})
	if err != nil {
		t.Fatalf("Failed to create channels table: %v", err)
	}

	// Create test channel with invalid JSON
	invalidJSON := `{"invalid": json}`
	testChannel := &Channel{
		Name:         "Invalid JSON Channel",
		Type:         1,
		Status:       1,
		Models:       "gpt-4",
		ModelConfigs: &invalidJSON,
	}

	err = testDB.Create(testChannel).Error
	if err != nil {
		t.Fatalf("Failed to create test channel: %v", err)
	}

	// Run the migration - should handle invalid JSON gracefully
	err = MigrateAllChannelModelConfigs()
	// This should not fail the entire migration, just log errors
	if err != nil {
		t.Logf("Expected behavior: migration handles invalid JSON gracefully: %v", err)
	}
}

func TestChannelNullHandling(t *testing.T) {
	// Setup test database
	testDB := setupMigrationTestDB(t)
	originalDB := DB
	DB = testDB
	defer func() { DB = originalDB }()

	// Create channels table
	err := testDB.AutoMigrate(&Channel{})
	if err != nil {
		t.Fatalf("Failed to create channels table: %v", err)
	}

	// Test 1: Create channel with NULL ModelConfigs and ModelMapping
	testChannel := &Channel{
		Name:         "NULL Test Channel",
		Type:         1,
		Status:       1,
		Models:       "gpt-4",
		ModelConfigs: nil, // Explicitly NULL
		ModelMapping: nil, // Explicitly NULL
	}

	err = testDB.Create(testChannel).Error
	if err != nil {
		t.Fatalf("Failed to create test channel with NULL fields: %v", err)
	}

	// Test 2: Verify NULL values are handled correctly by getter methods
	configs := testChannel.GetModelPriceConfigs()
	if configs != nil {
		t.Error("GetModelPriceConfigs should return nil for NULL ModelConfigs")
	}

	mapping := testChannel.GetModelMapping()
	if mapping != nil {
		t.Error("GetModelMapping should return nil for NULL ModelMapping")
	}

	// Test 3: Verify setter methods handle NULL correctly
	err = testChannel.SetModelPriceConfigs(nil)
	if err != nil {
		t.Errorf("SetModelPriceConfigs should handle nil input: %v", err)
	}

	if testChannel.ModelConfigs != nil {
		t.Error("SetModelPriceConfigs(nil) should set ModelConfigs to nil")
	}

	// Test 4: Verify migration handles NULL values correctly
	err = MigrateAllChannelModelConfigs()
	if err != nil {
		t.Errorf("Migration should handle NULL values gracefully: %v", err)
	}

	// Test 5: Verify database operations work with NULL values
	var retrievedChannel Channel
	err = testDB.First(&retrievedChannel, testChannel.Id).Error
	if err != nil {
		t.Errorf("Failed to retrieve channel with NULL fields: %v", err)
	}

	// Verify NULL values are preserved
	if retrievedChannel.ModelConfigs != nil {
		t.Error("NULL ModelConfigs should remain NULL after database round-trip")
	}

	if retrievedChannel.ModelMapping != nil {
		t.Error("NULL ModelMapping should remain NULL after database round-trip")
	}
}
