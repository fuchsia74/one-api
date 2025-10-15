package model

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test concurrent upserts to the same request_id do not create duplicates and end with the final quota value
func TestUpdateUserRequestCostQuotaByRequestID_Concurrency(t *testing.T) {
	setupTestDatabase(t)

	reqID := "test-upsert-concurrent-001"
	userID := 67890

	// Clean slate
	_ = DB.Where("request_id = ?", reqID).Delete(&UserRequestCost{}).Error

	const goroutines = 10
	var wg sync.WaitGroup

	// Start N-1 concurrent updates first to exercise contention
	for i := range goroutines - 1 {
		wg.Go(func() {
			_ = UpdateUserRequestCostQuotaByRequestID(userID, reqID, int64(i))
		})
	}
	wg.Wait()

	// Perform a final update deterministically so the expected value is known
	require.NoError(t, UpdateUserRequestCostQuotaByRequestID(userID, reqID, int64(goroutines-1)))

	rec, err := GetCostByRequestId(reqID)
	require.NoError(t, err)
	require.NotNil(t, rec)

	// Expect a single row and the latest quota value (goroutines-1)
	require.EqualValues(t, int64(goroutines-1), rec.Quota)

	// Verify no duplicates created for the same request_id
	var cnt int64
	require.NoError(t, DB.Model(&UserRequestCost{}).Where("request_id = ?", reqID).Count(&cnt).Error)
	require.EqualValues(t, 1, cnt, "should have exactly one row for the request_id")
}
