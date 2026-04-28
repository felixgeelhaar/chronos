package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/pipeline"
	"github.com/felixgeelhaar/chronos/internal/store"
	"github.com/google/uuid"
)

func runCompute(args []string) error {
	fs := flag.NewFlagSet("compute", flag.ContinueOnError)
	adapterName := fs.String("adapter", "", "adapter name (e.g. ascend)")
	scopeIDStr := fs.String("scope-id", "", "scope ID (UUID); alias for --coach-id retained for backward compatibility")
	coachIDStr := fs.String("coach-id", "", "scope ID (UUID), supplied to the adapter as coach_id")
	if err := fs.Parse(args); err != nil {
		return NewUserError("compute: %v", err)
	}

	if *adapterName == "" {
		return NewUserError("compute: --adapter is required")
	}
	scope := *scopeIDStr
	if scope == "" {
		scope = *coachIDStr
	}
	if scope == "" {
		return NewUserError("compute: --scope-id (or --coach-id) is required")
	}
	if _, err := uuid.Parse(scope); err != nil {
		return NewUserError("compute: invalid scope ID %q: %v", scope, err)
	}

	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		return NewUserError("compute: invalid configuration: %v", err)
	}

	src, ok := chronos.Get(*adapterName)
	if !ok {
		return NewNotFoundError("adapter %q not registered (available: %v)", *adapterName, chronos.Adapters())
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ComputationTimeout)
	defer cancel()

	conn, err := store.Open(ctx, cfg.DBType, cfg.DBConnStr)
	if err != nil {
		return NewSystemError(err, "compute: open store: %v", err)
	}
	defer func() { _ = conn.Close() }()

	logger := slog.Default().With("cmd", "compute", "adapter", *adapterName)

	res, err := pipeline.Compute(ctx, pipeline.ComputeInput{
		Source:       src,
		AdapterCfg:   map[string]string{"coach_id": scope, "scope_id": scope},
		EntityStates: conn.EntityStates,
		Signals:      conn.Signals,
		Engine:       pipeline.NewEngine(cfg),
		Logger:       logger,
	})
	if err != nil {
		return NewSystemError(err, "compute: %v", err)
	}

	fmt.Printf("Fetched %d entity states; emitted %d signals.\n", res.StatesFetched, res.SignalsCreated)
	return nil
}
