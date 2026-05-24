package api

import (
	"os"
	"testing"
)

// lookupEnvSafe returns the current value of key and whether it was
// set, for tests that want to snapshot before mutating and restore
// afterwards.
func lookupEnvSafe(t *testing.T, key string) (string, bool) {
	t.Helper()
	v, ok := os.LookupEnv(key)
	return v, ok
}

func restoreEnv(t *testing.T, key, value string, wasSet bool) {
	t.Helper()
	if wasSet {
		_ = os.Setenv(key, value)
		return
	}
	_ = os.Unsetenv(key)
}
