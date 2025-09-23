package model

import (
	"math/rand"
	"sync"
	"time"

	"github.com/Laisky/errors/v2"
	"github.com/Laisky/zap"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
)

type UserRequestCost struct {
	Id          int   `json:"id"`
	CreatedTime int64 `json:"created_time" gorm:"bigint"`
	UserID      int   `json:"user_id"`
	// Enforce uniqueness to avoid duplicate rows for the same request
	RequestID string  `json:"request_id" gorm:"uniqueIndex"`
	Quota     int64   `json:"quota"`
	CostUSD   float64 `json:"cost_usd" gorm:"-"`
	CreatedAt int64   `json:"created_at" gorm:"bigint;autoCreateTime:milli"`
	UpdatedAt int64   `json:"updated_at" gorm:"bigint;autoUpdateTime:milli"`
}

// NewUserRequestCost create a new UserRequestCost
func NewUserRequestCost(userID int, quotaID string, quota int64) *UserRequestCost {
	return &UserRequestCost{
		CreatedTime: helper.GetTimestamp(),
		UserID:      userID,
		RequestID:   quotaID,
		Quota:       quota,
	}
}

func (docu *UserRequestCost) Insert() error {
	go removeOldRequestCost()

	err := DB.Create(docu).Error
	return errors.Wrap(err, "failed to insert UserRequestCost")
}

// UpdateUserRequestCostQuotaByRequestID updates the quota for an existing request-cost record by request_id.
// If the record does not exist, it will create a new one with the provided userID and quota.
func UpdateUserRequestCostQuotaByRequestID(userID int, requestID string, quota int64) error {
	if requestID == "" {
		return errors.New("request id is empty")
	}

	go removeOldRequestCost()

	// Update-first approach to avoid unique conflict races without using clause.OnConflict
	// 1) Try update by request_id
	tx := DB.Model(&UserRequestCost{}).
		Where("request_id = ?", requestID).
		Update("quota", quota)
	if tx.Error != nil {
		return errors.Wrap(tx.Error, "failed to update UserRequestCost quota")
	}
	affected := tx.RowsAffected
	if affected > 0 {
		return nil
	}

	docu := &UserRequestCost{
		CreatedTime: helper.GetTimestamp(),
		UserID:      userID,
		RequestID:   requestID,
		Quota:       quota,
	}
	if err := DB.Create(docu).Error; err == nil {
		return nil
	} else {
		// If create failed (possibly due to unique race), retry update once
		if err2 := DB.Model(&UserRequestCost{}).
			Where("request_id = ?", requestID).
			Update("quota", quota).Error; err2 != nil {
			return errors.Wrap(err2, "failed to update UserRequestCost quota after create race")
		}
		return nil
	}
}

// GetCostByRequestId get cost by request id
func GetCostByRequestId(reqid string) (*UserRequestCost, error) {
	if reqid == "" {
		return nil, errors.New("request id is empty")
	}

	docu := &UserRequestCost{RequestID: reqid}
	var err error = nil
	if err = DB.First(docu, "request_id = ?", reqid).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get cost by request id")
	}

	docu.CostUSD = float64(docu.Quota) / 500000
	return docu, nil
}

var muRemoveOldRequestCost sync.Mutex

// removeOldRequestCost remove old request cost data,
// this function will be executed every 1/1000 times.
func removeOldRequestCost() {
	if rand.Float32() > 0.001 {
		return
	}

	if ok := muRemoveOldRequestCost.TryLock(); !ok {
		return
	}
	defer muRemoveOldRequestCost.Unlock()

	err := DB.
		Where("created_time < ?", helper.GetTimestamp()-3600*24*7).
		Delete(&UserRequestCost{}).Error
	if err != nil {
		logger.Logger.Error("failed to remove old request cost", zap.Error(err))
	}
}

// MigrateUserRequestCostEnsureUniqueRequestID ensures a unique index on request_id and deduplicates prior data.
// It is safe to run multiple times and should be invoked before AutoMigrate in InitDB.
func MigrateUserRequestCostEnsureUniqueRequestID() error {
	// If table does not exist yet, skip quietly; AutoMigrate will create it with the unique index from tags
	if !DB.Migrator().HasTable(&UserRequestCost{}) {
		return nil
	}
	// 1) Create table if missing via AutoMigrate-esque call on a temp DB clone? We assume table exists or will be created by later AutoMigrate.
	// 2) Deduplicate: keep the newest row (max(created_at)) per request_id, delete others.
	// Use vendor-specific SQL where necessary.

	// Dedup only if table exists
	// We try a vendor-agnostic approach using DELETE with subquery; if it fails, we log and continue with index creation.
	// MySQL/Postgres support DELETE using a subquery with row_number over partition; SQLite 3.25+ supports windows.
	// To keep it simple and safe, perform dedup in Go if SQL dialect errors.

	type pair struct {
		RequestID    string
		MaxUpdatedAt int64
	}

	// Fetch the latest record per request_id
	var latest []pair
	// Using GORM to perform group-by selection
	if err := DB.Table("user_request_costs").
		Select("request_id, MAX(updated_at) as max_updated_at").
		Group("request_id").
		Scan(&latest).Error; err != nil {
		return errors.Wrap(err, "scan latest user_request_costs per request_id failed")
	}

	if len(latest) > 0 {
		// Build a map for quick lookup
		keep := make(map[string]int64, len(latest))
		for _, p := range latest {
			keep[p.RequestID] = p.MaxUpdatedAt
		}

		// Delete duplicates where updated_at < max(updated_at) for that request_id
		// Do it in batches to avoid large transactions
		batchSize := 1000
		for reqID, maxU := range keep {
			// small sleep to reduce lock contention in large upgrades
			_ = maxU
			if err := DB.Where("request_id = ? AND updated_at < ?", reqID, keep[reqID]).
				Delete(&UserRequestCost{}).Error; err != nil {
				// Log and continue; this is best-effort
				logger.Logger.Warn("dedup delete failed", zap.Error(err))
			}
			batchSize--
			if batchSize == 0 {
				time.Sleep(10 * time.Millisecond)
				batchSize = 1000
			}
		}
	}

	// 3) Create unique index if missing. Use generic GORM API.
	// AutoMigrate later will also set the unique index from struct tag, but we ensure here too.
	// For safety across dialects, attempt raw SQL for each common DB.
	if common.UsingPostgreSQL {
		if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_request_costs_request_id ON user_request_costs (request_id)").Error; err != nil {
			return errors.Wrap(err, "create unique index on user_request_costs.request_id failed (postgres)")
		}
	} else if common.UsingMySQL {
		if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_request_costs_request_id ON user_request_costs (request_id)").Error; err != nil {
			// Some MySQL versions do not support IF NOT EXISTS for indexes; fallback: check existence
			// Try create without IF NOT EXISTS and ignore duplicate error
			_ = DB.Exec("CREATE UNIQUE INDEX idx_user_request_costs_request_id ON user_request_costs (request_id)").Error
		}
	} else if common.UsingSQLite {
		if err := DB.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_request_costs_request_id ON user_request_costs (request_id)").Error; err != nil {
			return errors.Wrap(err, "create unique index on user_request_costs.request_id failed (sqlite)")
		}
	}
	return nil
}
