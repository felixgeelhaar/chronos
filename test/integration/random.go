//go:build integration

package integration

import "crypto/rand"

// readRandom is a tiny wrapper so the test file can stay in build-tag
// mode without pulling crypto/rand into the no-tag build.
func readRandom(b []byte) (int, error) { return rand.Read(b) }
