package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felixgeelhaar/chronos/internal/api"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/notify"
	"github.com/felixgeelhaar/chronos/internal/observability"
	"github.com/felixgeelhaar/chronos/internal/pipeline"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/felixgeelhaar/chronos/internal/store"
)

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	port := fs.Int("port", 0, "HTTP port (overrides CHRONOS_HTTP_PORT)")
	host := fs.String("host", "", "HTTP host (overrides CHRONOS_HTTP_HOST)")
	if err := fs.Parse(args); err != nil {
		return NewUserError("serve: %v", err)
	}

	cfg := config.Default()
	if *port > 0 {
		cfg.HTTPPort = *port
	}
	if *host != "" {
		cfg.HTTPHost = *host
	}
	if err := cfg.Validate(); err != nil {
		return NewUserError("serve: invalid configuration: %v", err)
	}

	conn, err := store.Open(context.Background(), cfg.DBType, cfg.DBConnStr)
	if err != nil {
		return NewSystemError(err, "serve: open store: %v", err)
	}
	defer func() { _ = conn.Close() }()

	logger := slog.Default().With("cmd", "serve")
	metrics := observability.New()

	// Build the notifier set: webhooks (cross-process) + SSE
	// (in-process / browser). Both implement ports.Notifier and fan
	// out off the same Save call. SSE lives in this process; signals
	// reach it only when the in-process detection scheduler runs, so
	// we pair them: scheduler enabled <=> SSE useful.
	var notifiers notify.Multi
	if len(cfg.WebhookURLs) > 0 {
		notifiers = append(notifiers, notify.NewWebhook(notify.WebhookConfig{
			URLs:    cfg.WebhookURLs,
			Secret:  cfg.WebhookSecret,
			Timeout: cfg.WebhookTimeout,
			Retries: cfg.WebhookRetries,
		}, metrics, logger))
	}
	var sse *notify.SSE
	if cfg.DetectionInterval > 0 {
		sse = notify.NewSSE(16) // small per-client buffer; slow clients are dropped
		notifiers = append(notifiers, sse)
	}
	var notifier ports.Notifier
	if len(notifiers) > 0 {
		notifier = notifiers
	}

	signals := wrapWithNotifier(conn.Signals, notifier)

	srv := api.NewServer(conn.EntityStates, signals, metrics, logger)
	if sse != nil {
		srv = srv.WithSSE(sse)
	}
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	// Recover is the outermost layer so panics never escape the HTTP
	// server; Logging records every request once it has been served;
	// BearerAuth gates the API when CHRONOS_API_TOKEN is set (no-op
	// when empty so the default development experience is unchanged).
	handler := api.Chain(mux,
		api.Recover(logger),
		api.Logging(logger, metrics),
		api.BearerAuth(cfg.APIToken),
	)
	if cfg.APIToken != "" {
		logger.Info("api auth enabled", "scheme", "bearer")
	}

	addr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// rootCtx propagates shutdown to background goroutines (scheduler,
	// SSE drain). The HTTP server has its own Shutdown; cancelling
	// rootCtx unwinds everything else.
	rootCtx, cancelRoot := context.WithCancel(context.Background())
	defer cancelRoot()

	// In-process detection scheduler. Disabled when interval == 0 so
	// `serve` keeps its historical behaviour: streaming ingest only,
	// no detection. Enabled, it ticks Engine.Detect over each scope's
	// observations and writes signals via the notifier-wrapped repo —
	// which is what makes SSE see anything.
	if cfg.DetectionInterval > 0 {
		sched := pipeline.NewScheduler(conn.EntityStates, signals, pipeline.NewEngine(cfg), cfg.DetectionInterval, logger)
		go func() {
			if err := sched.Run(rootCtx); err != nil {
				logger.Error("scheduler exited with error", "err", err)
			}
		}()
	}

	// Graceful shutdown on SIGINT / SIGTERM. Stops accepting new
	// requests; lets in-flight ones drain for up to 10 seconds.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutdown signal received")
		cancelRoot() // stop the scheduler before draining HTTP
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(ctx)
	}()

	logger.Info("listening", "addr", addr, "store", cfg.DBType,
		"detection_interval", cfg.DetectionInterval, "webhooks", len(cfg.WebhookURLs))
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return NewSystemError(err, "serve: %v", err)
	}
	return nil
}
