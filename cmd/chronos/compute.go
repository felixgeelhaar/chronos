package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/notify"
	"github.com/felixgeelhaar/chronos/internal/observability"
	"github.com/felixgeelhaar/chronos/internal/pipeline"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/felixgeelhaar/chronos/internal/store"
	"github.com/google/uuid"
)

func runCompute(args []string) error {
	fs := flag.NewFlagSet("compute", flag.ContinueOnError)
	adapterName := fs.String("adapter", "", "adapter name (registered via chronos.Register from a blank-imported package)")
	scopeIDStr := fs.String("scope-id", "", "scope ID (UUID); alias for --coach-id retained for backward compatibility")
	coachIDStr := fs.String("coach-id", "", "deprecated alias for --scope-id; supplied to the adapter as cfg[\"coach_id\"] for back-compat with adapters that key on it")
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

	dsn, err := resolveDSN(cfg)
	if err != nil {
		return NewUserError("compute: %v", err)
	}
	conn, err := store.Open(ctx, dsn)
	if err != nil {
		return NewSystemError(err, "compute: open store: %v", err)
	}
	defer func() { _ = conn.Close() }()

	logger := slog.Default().With("cmd", "compute", "adapter", *adapterName)
	metrics := observability.New()

	// Wrap the signals repository with the configured push transports
	// so newly-detected signals fan out to webhooks (when configured)
	// the moment they are persisted. A nil notifier is a no-op.
	signals := wrapWithNotifier(conn.Signals, buildNotifier(cfg, metrics, logger))

	res, err := pipeline.Compute(ctx, pipeline.ComputeInput{
		Source:       src,
		AdapterCfg:   map[string]string{"coach_id": scope, "scope_id": scope},
		EntityStates: conn.EntityStates,
		Signals:      signals,
		Engine:       pipeline.NewEngine(cfg),
		Logger:       logger,
		Metrics:      metrics,
	})
	if err != nil {
		return NewSystemError(err, "compute: %v", err)
	}

	fmt.Printf("Fetched %d entity states; emitted %d signals.\n", res.StatesFetched, res.SignalsCreated)
	return nil
}

// buildNotifier assembles a Multi notifier from the configured push
// transports. Returns nil when nothing is configured so the call site
// can pass it to wrapWithNotifier unconditionally.
func buildNotifier(cfg *config.Config, metrics *observability.Metrics, logger *slog.Logger) ports.Notifier {
	var ns notify.Multi
	if len(cfg.WebhookURLs) > 0 {
		wh := notify.NewWebhook(notify.WebhookConfig{
			URLs:    cfg.WebhookURLs,
			Secret:  cfg.WebhookSecret,
			Timeout: cfg.WebhookTimeout,
			Retries: cfg.WebhookRetries,
		}, metrics, logger)
		ns = append(ns, wh)
	}
	if len(ns) == 0 {
		return nil
	}
	return ns
}

// wrapWithNotifier returns the bare repository when notifier is nil so
// we don't pay a wrapper cost for the common no-push case.
func wrapWithNotifier(repo ports.SignalRepository, notifier ports.Notifier) ports.SignalRepository {
	if notifier == nil {
		return repo
	}
	return notify.WrapSignals(repo, notifier)
}
