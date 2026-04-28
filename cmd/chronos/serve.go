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
	"github.com/felixgeelhaar/chronos/internal/observability"
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
	srv := api.NewServer(conn.EntityStates, conn.Signals, metrics, logger)
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

	// Graceful shutdown on SIGINT / SIGTERM. Stops accepting new
	// requests; lets in-flight ones drain for up to 10 seconds.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(ctx)
	}()

	logger.Info("listening", "addr", addr, "store", cfg.DBType)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return NewSystemError(err, "serve: %v", err)
	}
	return nil
}
