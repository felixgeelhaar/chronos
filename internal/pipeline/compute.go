// Package pipeline orchestrates Chronos's compute job. It is the only
// place in the codebase that wires "fetch from adapter → persist
// observations → run detectors → persist signals" together. Keeping the
// pipeline in its own package lets multiple entrypoints (CLI, future
// schedulers, future MCP servers) share the same flow.
package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/detect"
	"github.com/felixgeelhaar/chronos/internal/observability"
	"github.com/felixgeelhaar/chronos/internal/ports"
)

// ComputeInput names everything the compute job needs at the boundary.
type ComputeInput struct {
	Source       chronos.Source
	AdapterCfg   map[string]string
	EntityStates ports.EntityStateRepository
	Signals      ports.SignalRepository
	Engine       *detect.Engine
	Logger       *slog.Logger
	// Metrics is optional; when set, the pipeline records observations
	// and emitted signals so the /metrics endpoint can surface them.
	Metrics *observability.Metrics
}

// ComputeResult summarises a successful run for telemetry and CLI output.
type ComputeResult struct {
	StatesFetched  int
	SignalsCreated int
}

// Compute runs the full pipeline once.
func Compute(ctx context.Context, in ComputeInput) (ComputeResult, error) {
	if in.Source == nil {
		return ComputeResult{}, fmt.Errorf("pipeline: source required")
	}
	if in.Engine == nil {
		return ComputeResult{}, fmt.Errorf("pipeline: engine required")
	}
	logger := in.Logger
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("fetch begin", "adapter", in.Source.Name())
	states, err := in.Source.Fetch(ctx, in.AdapterCfg)
	if err != nil {
		return ComputeResult{}, fmt.Errorf("pipeline: fetch: %w", err)
	}
	logger.Info("fetch complete", "adapter", in.Source.Name(), "states", len(states))

	if len(states) == 0 {
		return ComputeResult{}, nil
	}

	if err := in.EntityStates.Save(ctx, in.Source.Name(), states); err != nil {
		return ComputeResult{}, fmt.Errorf("pipeline: save states: %w", err)
	}
	in.Metrics.ObserveObservations(in.Source.Name(), len(states))

	signals := in.Engine.Detect(ctx, states)
	logger.Info("signals detected", "count", len(signals))

	for _, sig := range signals {
		if err := in.Signals.Save(ctx, sig); err != nil {
			logger.Error("signal save failed", "id", sig.ID, "err", err)
			continue
		}
		in.Metrics.ObserveSignal(string(sig.Pattern))
	}

	return ComputeResult{
		StatesFetched:  len(states),
		SignalsCreated: len(signals),
	}, nil
}

// NewEngine is a thin alias so callers can build an Engine from this
// package without depending directly on internal/detect.
func NewEngine(cfg *config.Config) *detect.Engine { return detect.NewEngine(cfg) }
