package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// This test ensures the 413 path's tryLargerMaxTokens filtering logic does not panic and
// returns an error when no candidates exist after filtering.
func TestCacheGetRandomSatisfiedChannelExcluding_413Filtering_NoCandidates(t *testing.T) {
	// Memory cache disabled path falls back to DB function which we do not exercise here.
	// We call the function directly with an empty in-memory map to validate safe error.
	group := "default"
	model := "gpt-3.5-turbo"
	_, err := CacheGetRandomSatisfiedChannelExcluding(group, model, false, map[int]bool{1: true}, true)
	require.Error(t, err)
}
