package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateUserRequestCostQuotaByRequestID_CreateAndUpdate(t *testing.T) {
	setupTestDatabase(t)

	reqID := "test-upsert-req-001"
	userID := 12345

	// Ensure clean slate
	_ = DB.Where("request_id = ?", reqID).Delete(&UserRequestCost{}).Error

	// Create via upsert helper (record does not exist yet)
	err := UpdateUserRequestCostQuotaByRequestID(userID, reqID, 42)
	require.NoError(t, err)

	rec, err := GetCostByRequestId(reqID)
	require.NoError(t, err)
	require.NotNil(t, rec)
	assert.Equal(t, userID, rec.UserID)
	assert.Equal(t, reqID, rec.RequestID)
	assert.Equal(t, int64(42), rec.Quota)

	// Update existing record
	err = UpdateUserRequestCostQuotaByRequestID(userID, reqID, 100)
	require.NoError(t, err)

	rec2, err := GetCostByRequestId(reqID)
	require.NoError(t, err)
	require.NotNil(t, rec2)
	assert.Equal(t, userID, rec2.UserID)
	assert.Equal(t, reqID, rec2.RequestID)
	assert.Equal(t, int64(100), rec2.Quota)
}
