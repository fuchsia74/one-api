package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
)

// MigrateAbilitySuspendUntilColumn normalizes legacy suspend_until values so the column can be stored
// as a proper datetime/timestamp type. This migration is idempotent and safe to run on every startup.
// It relies on information_schema metadata queries; deployments must grant read access or the migration
// will fail fast so operators can address the missing privileges explicitly.
func MigrateAbilitySuspendUntilColumn() error {
	logger.Logger.Info("Starting ability suspend_until migration")
	if !DB.Migrator().HasTable(&Ability{}) {
		logger.Logger.Debug("Abilities table not found, skipping suspend_until migration")
		return nil
	}

	var err error
	switch {
	case common.UsingMySQL.Load():
		err = migrateAbilitySuspendUntilMySQL()
		err = errors.Wrap(err, "migrateAbilitySuspendUntilMySQL")
	case common.UsingPostgreSQL.Load():
		err = migrateAbilitySuspendUntilPostgres()
		err = errors.Wrap(err, "migrateAbilitySuspendUntilPostgres")
	default:
		// SQLite stores everything as TEXT; normalizing the payload is enough.
		err = normalizeAbilitySuspendUntilValues()
		err = errors.Wrap(err, "normalizeAbilitySuspendUntilValues")
	}
	if err != nil {
		return err
	}

	logger.Logger.Debug("Completed ability suspend_until migration")
	return nil
}

// migrateAbilitySuspendUntilMySQL converts legacy suspend_until metadata on MySQL installations
// into real DATETIME columns after normalizing the existing values.
func migrateAbilitySuspendUntilMySQL() error {
	logger.Logger.Debug("Running MySQL ability suspend_until migration")
	dataType, err := mysqlColumnDataType("abilities", "suspend_until")
	if err != nil {
		return errors.Wrap(err, "query abilities.suspend_until column type")
	}
	if dataType == "" || dataType == "datetime" || dataType == "timestamp" {
		logger.Logger.Debug("MySQL suspend_until column already normalized",
			zap.String("data_type", dataType))
		return nil
	}

	logger.Logger.Debug("Normalizing legacy MySQL suspend_until values before column alter")
	if err := normalizeAbilitySuspendUntilValues(); err != nil {
		return err
	}

	if err := DB.Exec("ALTER TABLE abilities MODIFY suspend_until DATETIME NULL").Error; err != nil {
		return errors.Wrap(err, "alter abilities.suspend_until to DATETIME")
	}

	logger.Logger.Debug("MySQL suspend_until column migrated to DATETIME")
	return nil
}

// migrateAbilitySuspendUntilPostgres converts legacy suspend_until metadata on PostgreSQL installations
// into TIMESTAMP columns after normalizing the existing values.
func migrateAbilitySuspendUntilPostgres() error {
	logger.Logger.Debug("Running PostgreSQL ability suspend_until migration")
	dataType, err := postgresColumnDataType("abilities", "suspend_until")
	if err != nil {
		return errors.Wrap(err, "query abilities.suspend_until column type (postgres)")
	}
	if dataType == "" || strings.Contains(dataType, "timestamp") {
		logger.Logger.Debug("PostgreSQL suspend_until column already normalized",
			zap.String("data_type", dataType))
		return nil
	}

	logger.Logger.Debug("Normalizing legacy PostgreSQL suspend_until values before column alter")
	if err := normalizeAbilitySuspendUntilValues(); err != nil {
		return err
	}

	alter := "ALTER TABLE abilities ALTER COLUMN suspend_until TYPE TIMESTAMP USING NULLIF(suspend_until, '')::timestamp"
	if err := DB.Exec(alter).Error; err != nil {
		return errors.Wrap(err, "alter abilities.suspend_until to TIMESTAMP (postgres)")
	}
	logger.Logger.Debug("PostgreSQL suspend_until column migrated to TIMESTAMP")
	return nil
}

// normalizeAbilitySuspendUntilValues rewrites legacy suspend_until data into UTC timestamp strings before
// type conversion. MySQL and SQLite receive "YYYY-MM-DD HH:MM:SS" strings (interpreted as UTC by the caller),
// while PostgreSQL is updated with RFC3339 strings that carry the timezone explicitly.
func normalizeAbilitySuspendUntilValues() error {
	logger.Logger.Debug("Normalizing legacy ability suspend_until values")
	groupCol := "`group`"
	if common.UsingPostgreSQL.Load() {
		groupCol = `"group"`
	}

	selectExpr := fmt.Sprintf("%s AS group_key, model, channel_id, suspend_until", groupCol)
	type abilitySuspendRow struct {
		Group     string `gorm:"column:group_key"`
		Model     string `gorm:"column:model"`
		ChannelID int    `gorm:"column:channel_id"`
		Raw       []byte `gorm:"column:suspend_until"`
	}

	var rows []abilitySuspendRow
	if err := DB.Table("abilities").
		Select(selectExpr).
		Where("suspend_until IS NOT NULL AND suspend_until <> ''").
		Find(&rows).Error; err != nil {
		return errors.Wrap(err, "load legacy ability suspend_until values")
	}

	if len(rows) == 0 {
		logger.Logger.Debug("No legacy ability suspend_until values required normalization")
		return nil
	}

	updateSQL := fmt.Sprintf("UPDATE abilities SET suspend_until = ? WHERE %s = ? AND model = ? AND channel_id = ?", groupCol)
	var updatedCount int
	var clearedCount int
	for _, row := range rows {
		parsed, ok := parseLegacySuspendUntil(row.Raw)
		if !ok {
			logger.Logger.Warn("unable to parse legacy suspend_until value, resetting to NULL",
				zap.String("group", row.Group),
				zap.String("model", row.Model),
				zap.Int("channel_id", row.ChannelID),
				zap.ByteString("raw", row.Raw))
			if err := DB.Exec(fmt.Sprintf("UPDATE abilities SET suspend_until = NULL WHERE %s = ? AND model = ? AND channel_id = ?", groupCol), row.Group, row.Model, row.ChannelID).Error; err != nil {
				return errors.Wrap(err, "clear invalid suspend_until value")
			}
			clearedCount++
			continue
		}

		formatted := parsed.UTC().Format("2006-01-02 15:04:05")
		if common.UsingPostgreSQL.Load() {
			formatted = parsed.UTC().Format(time.RFC3339)
		}

		if err := DB.Exec(updateSQL, formatted, row.Group, row.Model, row.ChannelID).Error; err != nil {
			return errors.Wrap(err, "update normalized suspend_until value")
		}
		updatedCount++
	}

	logger.Logger.Debug("Normalized ability suspend_until values",
		zap.Int("rows_processed", len(rows)),
		zap.Int("rows_updated", updatedCount),
		zap.Int("rows_cleared", clearedCount))

	return nil
}

// parseLegacySuspendUntil attempts to parse historical suspend_until values emitted by various
// releases. It supports Unix epochs in seconds/milliseconds/microseconds as well as numerous ISO-8601
// layouts and returns the parsed time in UTC when possible.
func parseLegacySuspendUntil(raw []byte) (time.Time, bool) {
	if len(raw) == 0 {
		return time.Time{}, false
	}

	str := strings.TrimSpace(string(raw))
	if str == "" {
		return time.Time{}, false
	}

	if unix, err := strconv.ParseInt(str, 10, 64); err == nil {
		switch {
		case len(str) >= 16:
			return time.UnixMicro(unix), true
		case len(str) >= 13:
			return time.UnixMilli(unix), true
		case len(str) >= 10:
			return time.Unix(unix, 0), true
		}
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, str); err == nil {
			return t, true
		}
	}

	withoutZone := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, layout := range withoutZone {
		if t, err := time.ParseInLocation(layout, str, time.UTC); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

// mysqlColumnDataType retrieves the DATA_TYPE entry for the specified column via information_schema.
// The migration fails fast if the current user is not permitted to read metadata for the target table.
func mysqlColumnDataType(table, column string) (string, error) {
	var dataType string
	query := "SELECT DATA_TYPE FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?"
	if err := DB.Raw(query, table, column).Scan(&dataType).Error; err != nil {
		return "", err
	}
	return strings.ToLower(dataType), nil
}

// postgresColumnDataType retrieves the data_type entry for the specified column via information_schema.
// The migration fails fast if the current user is not permitted to read metadata for the target table.
func postgresColumnDataType(table, column string) (string, error) {
	var dataType string
	query := "SELECT data_type FROM information_schema.columns WHERE table_name = ? AND column_name = ?"
	if err := DB.Raw(query, table, column).Scan(&dataType).Error; err != nil {
		return "", err
	}
	return strings.ToLower(dataType), nil
}
