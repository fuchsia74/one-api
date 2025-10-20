package logger

import "sync"

// ResetSetupLogOnceForTests resets the setupLogOnce guard so tests can re-run SetupLogger without touching
// unexported state in production code. It must only be used from test files.
func ResetSetupLogOnceForTests() {
	setupLogOnce = sync.Once{}
}
