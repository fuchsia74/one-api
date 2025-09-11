package helper

import (
	"fmt"
	"time"
)

// GetTimestamp get current timestamp in seconds
func GetTimestamp() int64 {
	return time.Now().Unix()
}

func GetTimeString() string {
	now := time.Now()
	return fmt.Sprintf("%s%d", now.Format("20060102150405"), now.UnixNano()%1e9)
}

// CalcElapsedTime return the elapsed time in milliseconds (ms)
func CalcElapsedTime(start time.Time) int64 {
	elapsed := time.Since(start)
	ms := elapsed.Milliseconds()
	if ms == 0 && elapsed > 0 {
		// Ensure non-zero latency for sub-millisecond operations so UI does not show 0
		return 1
	}
	return ms
}
