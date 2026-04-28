// Package main is the chronos command-line entrypoint. Subcommand bodies
// live in sibling files (compute.go, serve.go, version.go); this file
// dispatches and renders structured errors.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	// Adapters must be imported here for their init() registrations to take
	// effect. New adapters are added to this list (or to a build-tagged
	// shim file) so they appear in chronos.Adapters().
	_ "github.com/felixgeelhaar/chronos/adapters/ascend"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(int(ExitUsage))
	}

	var err error
	switch os.Args[1] {
	case "compute":
		err = runCompute(os.Args[2:])
	case "serve":
		err = runServe(os.Args[2:])
	case "version", "-v", "--version":
		runVersion()
		return
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		err = NewUserError("unknown command %q", os.Args[1])
	}

	if err != nil {
		exitWithError(err)
	}
}

func printUsage() {
	fmt.Println(`Chronos — generic pattern detection engine for time-series data.

Usage:
  chronos <command> [flags]

Commands:
  compute    Run pattern detection for a scope
  serve      Start HTTP API server
  version    Print version
  help       Show this help

Run "chronos compute --help" or "chronos serve --help" for command flags.

Configuration is via CHRONOS_* environment variables; see README.md.`)
}

// exitWithError renders the error in a consistent shape and exits with the
// code carried by ChronosError (or ExitError for any other error).
func exitWithError(err error) {
	var ce *ChronosError
	if errors.As(err, &ce) {
		fmt.Fprintf(os.Stderr, "error: %s\n", ce.Message)
		if ce.Hint != "" {
			fmt.Fprintf(os.Stderr, "hint: %s\n", ce.Hint)
		}
		if ce.Cause != nil && os.Getenv("CHRONOS_VERBOSE") != "" {
			fmt.Fprintf(os.Stderr, "cause: %v\n", ce.Cause)
		}
		os.Exit(int(ce.Code))
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(int(ExitError))
}
