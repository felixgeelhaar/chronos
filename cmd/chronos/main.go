package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felixgeelhaar/chronos/internal/adapter"
	"github.com/felixgeelhaar/chronos/internal/api"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/insights"
	"github.com/felixgeelhaar/chronos/internal/store"
	"github.com/google/uuid"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compute":
		computeCmd()
	case "serve":
		serveCmd()
	case "version":
		fmt.Println("chronos v0.1.0")
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Chronos - Generic pattern detection engine for time-series data

Usage:
  chronos <command> [flags]

Commands:
  compute    Run pattern detection computation
  serve      Start HTTP API server
  version    Print version
  help       Show this help

Environment Variables:
  CHRONOS_DB_TYPE         Database type (sqlite, postgres, memory)
  CHRONOS_DB_CONN         Database connection string or path
  CHRONOS_SIM_THRESHOLD   Minimum similarity threshold (0.0-1.0)
  CHRONOS_MIN_SAMPLE      Minimum sample size for insights
  CHRONOS_HTTP_PORT       HTTP server port
  CHRONOS_HTTP_HOST       HTTP server host

Examples:
  chronos compute --adapter=ascend --coach-id=xxx
  chronos serve --port=7778
`)
}

func computeCmd() {
	fs := flag.NewFlagSet("compute", flag.ExitOnError)
	adapterName := fs.String("adapter", "", "Adapter name (e.g., ascend)")
	coachID := fs.String("coach-id", "", "Coach ID for scope")
	_ = fs.Parse(os.Args[2:])

	if *adapterName == "" || *coachID == "" {
		fmt.Fprintln(os.Stderr, "Error: --adapter and --coach-id are required")
		os.Exit(1)
	}

	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Create store
	factory := &store.Factory{}
	st, err := factory.New(cfg.DBType, cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Store error: %v", err)
	}
	defer st.Close()

	// Load adapter
	src, ok := adapter.Get(*adapterName)
	if !ok {
		log.Fatalf("Unknown adapter: %s", *adapterName)
	}

	scopeID, err := uuid.Parse(*coachID)
	if err != nil {
		log.Fatalf("Invalid coach-id: %v", err)
	}
	_ = scopeID.String() // unused but validated

	// Fetch data
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ComputationTimeout)
	defer cancel()

	states, err := src.Fetch(ctx, map[string]string{"coach_id": *coachID})
	if err != nil {
		log.Fatalf("Fetch error: %v", err)
	}

	log.Printf("Fetched %d entity states from %s", len(states), *adapterName)

	// Save to store
	if err := st.SaveEntityStates(ctx, *adapterName, states); err != nil {
		log.Fatalf("Save error: %v", err)
	}

	// Generate insights
	gen := insights.NewGenerator(cfg)
	newInsights, err := gen.Generate(states)
	if err != nil {
		log.Fatalf("Generation error: %v", err)
	}

	log.Printf("Generated %d insights", len(newInsights))

	// Save insights
	for _, in := range newInsights {
		if err := st.SaveInsight(ctx, in); err != nil {
			log.Printf("Failed to save insight: %v", err)
		}
	}

	log.Println("Computation complete")
}

func serveCmd() {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 0, "HTTP port (overrides CHRONOS_HTTP_PORT)")
	host := fs.String("host", "", "HTTP host (overrides CHRONOS_HTTP_HOST)")
	_ = fs.Parse(os.Args[2:])

	cfg := config.Default()
	if *port > 0 {
		cfg.HTTPPort = *port
	}
	if *host != "" {
		cfg.HTTPHost = *host
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Create store
	factory := &store.Factory{}
	st, err := factory.New(cfg.DBType, cfg.DBConnStr)
	if err != nil {
		log.Fatalf("Store error: %v", err)
	}
	defer st.Close()

	// Setup HTTP server
	mux := http.NewServeMux()
	server := api.NewServer(st, cfg)
	server.RegisterRoutes(mux)

	addr := fmt.Sprintf("%s:%d", cfg.HTTPHost, cfg.HTTPPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	log.Printf("Chronos server listening on http://%s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
