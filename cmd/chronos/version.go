package main

import "fmt"

// These vars are populated at build time via -ldflags "-X main.version=...".
// Keeping them in their own file means `go build` without ldflags still
// produces a binary that prints sensible defaults.
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func runVersion() {
	fmt.Printf("chronos %s (commit %s, built %s)\n", version, commit, buildDate)
}
