// Package main is the chronos command-line entrypoint. Subcommand bodies
// live in sibling files (compute.go, serve.go, version.go); this file
// dispatches and renders structured errors.
package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/store"

	// Adapters self-register via init(). The bundled chronos binary
	// ships with no adapters baked in — out-of-tree adapters import
	// chronos as a library and register themselves; downstream
	// binaries blank-import the adapters they need. See
	// docs/adapters.md for the contract.

	// Persistence providers self-register with the store factory in
	// init(); blank-imports here ensure those side effects fire before
	// any subcommand calls store.Open. Drop any of these to produce a
	// build that does not link the corresponding driver — useful for
	// shrinking the static binary in single-backend deployments.
	_ "github.com/felixgeelhaar/chronos/internal/store/libsql"
	_ "github.com/felixgeelhaar/chronos/internal/store/memory"
	_ "github.com/felixgeelhaar/chronos/internal/store/mysql"
	_ "github.com/felixgeelhaar/chronos/internal/store/postgres"
	_ "github.com/felixgeelhaar/chronos/internal/store/sqlite"
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
	case "mcp":
		err = runMCP(os.Args[2:])
	case "migrate":
		err = runMigrate(os.Args[2:])
	case "health":
		err = runHealth(os.Args[2:])
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
  mcp        Start the MCP stdio server (list_signals / ingest / describe_detector)
  migrate    Inspect or apply schema migrations (status / up)
  health     Probe a running server's /health endpoint (container HEALTHCHECK)
  version    Print version
  help       Show this help

Run "chronos compute --help" or "chronos serve --help" for command flags.

Configuration is via CHRONOS_* environment variables; see README.md.`)
}

// resolveDSN returns the persistence DSN to hand to store.Open.
// Precedence: CHRONOS_DB_DSN (the new primary, mirrors the Mnemos
// ADR 0001 contract) > legacy CHRONOS_DB_TYPE + CHRONOS_DB_CONN
// translated via store.LegacyDSN. Returns a usage error when both
// paths are empty.
func resolveDSN(cfg *config.Config) (string, error) {
	if cfg.DBDSN != "" {
		return cfg.DBDSN, nil
	}
	dsn, err := store.LegacyDSN(cfg.DBType, cfg.DBConnStr)
	if err != nil {
		return "", err
	}
	if dsn == "" {
		return "", fmt.Errorf("set CHRONOS_DB_DSN (or the legacy CHRONOS_DB_TYPE + CHRONOS_DB_CONN pair)")
	}
	return dsn, nil
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
