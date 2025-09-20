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
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		i := i
		go func() {
			defer wg.Done()
			// Interleaved quotas, last write wins
			_ = UpdateUserRequestCostQuotaByRequestID(userID, reqID, int64(i))
		}()
	}
	wg.Wait()

	rec, err := GetCostByRequestId(reqID)
	require.NoError(t, err)
	require.NotNil(t, rec)

	// Expect a single row and the latest quota value (goroutines-1)
	require.EqualValues(t, int64(goroutines-1), rec.Quota)
}
