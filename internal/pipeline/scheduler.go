package pipeline

import (
	"context"
	"log/slog"
	"time"

	"github.com/felixgeelhaar/chronos/internal/detect"
	"github.com/felixgeelhaar/chronos/internal/ports"
)

// Scheduler runs detection in-process at a configurable cadence,
// without re-fetching observations from adapters. It is the bridge
// between the streaming Ingest path (POST /v1/ingest) and consumers
// of /v1/signals/stream — without a scheduler, signals only get
// produced by the compute CLI in a separate process and SSE clients
// would never see anything.
//
// Each tick the scheduler enumerates scopes that have observations,
// loads each scope's history, runs every detector over it, and
// persists the resulting signals through the configured (typically
// notifier-wrapped) SignalRepository. Detection is idempotent at the
// signal-id level via uuid.NewV4 inside detectors, so re-runs over
// the same data simply produce additional signals — operators choose
// retention via DeleteOlderThan and the cadence via
// CHRONOS_DETECTION_INTERVAL.
type Scheduler struct {
	states   ports.EntityStateRepository
	signals  ports.SignalRepository
	engine   *detect.Engine
	interval time.Duration
	logger   *slog.Logger
}

// NewScheduler builds a scheduler. interval == 0 produces a scheduler
// whose Run is a no-op (returns immediately) — convenient for the
// "scheduler off" config path so callers don't have to special-case
// the nil scheduler.
func NewScheduler(states ports.EntityStateRepository, signals ports.SignalRepository, engine *detect.Engine, interval time.Duration, logger *slog.Logger) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Scheduler{
		states:   states,
		signals:  signals,
		engine:   engine,
		interval: interval,
		logger:   logger,
	}
}

// Run blocks until ctx is cancelled, ticking detection every
// interval. interval <= 0 is treated as disabled and Run returns
// immediately. Each tick is bounded by ctx — if a scope load takes
// longer than the interval, the next tick still fires (no
// throttling) but the previous tick's writes are best-effort.
func (s *Scheduler) Run(ctx context.Context) error {
	if s.interval <= 0 {
		return nil
	}
	t := time.NewTicker(s.interval)
	defer t.Stop()

	s.logger.Info("detection scheduler started", "interval", s.interval)
	for {
		select {
		case <-ctx.Done():
			s.logger.Info("detection scheduler stopped")
			return nil
		case <-t.C:
			s.tick(ctx)
		}
	}
}

// tick performs one full detection pass across all scopes. Errors on
// any single scope are logged and do not abort the pass.
func (s *Scheduler) tick(ctx context.Context) {
	scopes, err := s.states.ListScopes(ctx)
	if err != nil {
		s.logger.Error("scheduler: list scopes failed", "err", err)
		return
	}
	for _, scopeID := range scopes {
		states, err := s.states.ListByScope(ctx, scopeID)
		if err != nil {
			s.logger.Error("scheduler: load scope failed", "scope_id", scopeID, "err", err)
			continue
		}
		if len(states) == 0 {
			continue
		}
		signals := s.engine.Detect(ctx, states)
		for _, sig := range signals {
			if err := s.signals.Save(ctx, sig); err != nil {
				s.logger.Error("scheduler: signal save failed", "scope_id", scopeID, "signal_id", sig.ID, "err", err)
			}
		}
	}
}
