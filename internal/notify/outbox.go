package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/google/uuid"
)

// AckingNotifier is the contract for downstream notifiers that can
// report whether a delivery succeeded. The standard webhook is
// best-effort; AckingWebhook (next sibling) reports per-call success
// so the outbox can decide whether to retry.
type AckingNotifier interface {
	NotifyAck(ctx context.Context, sig domain.Signal) error
}

// Outbox is a Notifier wrapper that retries failed deliveries on a
// background sweeper, providing at-least-once semantics within a
// process lifetime. Pending deliveries do NOT survive process
// restart — operators wanting durable at-least-once delivery wire
// their own persistent Notifier.
//
// Retry policy: exponential backoff starting at MinBackoff,
// doubling up to MaxBackoff, with up to MaxAttempts before the
// delivery is dropped and metrics increment a "give-up" counter.
type Outbox struct {
	inner       AckingNotifier
	maxAttempts int
	minBackoff  time.Duration
	maxBackoff  time.Duration
	persistPath string

	mu      sync.Mutex
	pending []pendingDelivery

	wakeup chan struct{}
	stop   chan struct{}
	wg     sync.WaitGroup
}

type pendingDelivery struct {
	signal     domain.Signal
	attempts   int
	nextRetry  time.Time
	lastError  error
}

// OutboxConfig tunes the retrier. All fields default to sensible
// values when zero:
//
//	MaxAttempts default 5
//	MinBackoff  default 1s
//	MaxBackoff  default 30s
//
// PersistencePath, when set, makes the outbox durable across
// process restarts: pending deliveries are snapshotted to the path
// after each enqueue/flush, and re-loaded at Start. The file
// format is JSON and the parent directory is created with 0o700
// permissions.
type OutboxConfig struct {
	MaxAttempts     int
	MinBackoff      time.Duration
	MaxBackoff      time.Duration
	PersistencePath string
}

// NewOutbox constructs an Outbox. Callers must invoke Start to
// activate the sweeper goroutine; Stop drains in-flight retries.
func NewOutbox(inner AckingNotifier, cfg OutboxConfig) (*Outbox, error) {
	if inner == nil {
		return nil, errors.New("outbox: inner notifier required")
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.MinBackoff <= 0 {
		cfg.MinBackoff = time.Second
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 30 * time.Second
	}
	o := &Outbox{
		inner:       inner,
		maxAttempts: cfg.MaxAttempts,
		minBackoff:  cfg.MinBackoff,
		maxBackoff:  cfg.MaxBackoff,
		persistPath: cfg.PersistencePath,
		wakeup:      make(chan struct{}, 1),
		stop:        make(chan struct{}),
	}
	if o.persistPath != "" {
		if err := o.loadFromDisk(); err != nil {
			return nil, fmt.Errorf("outbox: load %s: %w", o.persistPath, err)
		}
	}
	return o, nil
}

// Start runs the sweeper until ctx is done or [Stop] is called.
func (o *Outbox) Start(ctx context.Context) {
	o.wg.Add(1)
	go o.sweep(ctx)
}

// Stop signals the sweeper to drain. Returns when the goroutine has
// exited.
func (o *Outbox) Stop() {
	close(o.stop)
	o.wg.Wait()
}

// Notify is the [ports.Notifier] entry point: fire-and-forget. The
// underlying delivery is enqueued in the outbox; immediate
// delivery is attempted once, and failed deliveries are retried by
// the sweeper.
func (o *Outbox) Notify(ctx context.Context, sig domain.Signal) {
	if err := o.inner.NotifyAck(ctx, sig); err != nil {
		o.enqueue(sig, err)
	}
}

func (o *Outbox) enqueue(sig domain.Signal, lastErr error) {
	o.mu.Lock()
	o.pending = append(o.pending, pendingDelivery{
		signal:    sig,
		attempts:  1,
		nextRetry: time.Now().Add(o.minBackoff),
		lastError: lastErr,
	})
	o.mu.Unlock()
	o.persistLocked()
	select {
	case o.wakeup <- struct{}{}:
	default:
	}
}

// persistLocked snapshots the pending queue to disk when persistence
// is configured. Errors are logged via stderr but don't fail the
// caller — durability is best-effort layered above the in-memory
// outbox.
func (o *Outbox) persistLocked() {
	if o.persistPath == "" {
		return
	}
	o.mu.Lock()
	pending := make([]persistedDelivery, 0, len(o.pending))
	for _, p := range o.pending {
		var lastErr string
		if p.lastError != nil {
			lastErr = p.lastError.Error()
		}
		pending = append(pending, persistedDelivery{
			Signal:    p.signal,
			Attempts:  p.attempts,
			NextRetry: p.nextRetry,
			LastError: lastErr,
		})
	}
	o.mu.Unlock()
	body, err := json.Marshal(pending)
	if err != nil {
		fmt.Fprintf(os.Stderr, "outbox: marshal: %v\n", err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(o.persistPath), 0o700); err != nil {
		fmt.Fprintf(os.Stderr, "outbox: mkdir: %v\n", err)
		return
	}
	tmp := o.persistPath + ".tmp"
	if err := os.WriteFile(tmp, body, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "outbox: write: %v\n", err)
		return
	}
	if err := os.Rename(tmp, o.persistPath); err != nil {
		fmt.Fprintf(os.Stderr, "outbox: rename: %v\n", err)
	}
}

// loadFromDisk restores pending deliveries from the configured path.
// Missing files are not errors — first-boot has nothing to restore.
func (o *Outbox) loadFromDisk() error {
	data, err := os.ReadFile(filepath.Clean(o.persistPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	var rows []persistedDelivery
	if err := json.Unmarshal(data, &rows); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, r := range rows {
		var lastErr error
		if r.LastError != "" {
			lastErr = errors.New(r.LastError)
		}
		o.pending = append(o.pending, pendingDelivery{
			signal:    r.Signal,
			attempts:  r.Attempts,
			nextRetry: r.NextRetry,
			lastError: lastErr,
		})
	}
	return nil
}

// persistedDelivery is the on-disk representation. Errors are stored
// as their message string; we don't try to round-trip Go error
// types.
type persistedDelivery struct {
	Signal    domain.Signal `json:"signal"`
	Attempts  int           `json:"attempts"`
	NextRetry time.Time     `json:"next_retry"`
	LastError string        `json:"last_error,omitempty"`
}

// PendingCount returns the number of deliveries waiting to retry.
// Useful for tests and metrics surface.
func (o *Outbox) PendingCount() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return len(o.pending)
}

func (o *Outbox) sweep(ctx context.Context) {
	defer o.wg.Done()
	t := time.NewTicker(o.minBackoff)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-o.stop:
			return
		case <-t.C:
			o.flush(ctx)
		case <-o.wakeup:
			o.flush(ctx)
		}
	}
}

func (o *Outbox) flush(ctx context.Context) {
	defer o.persistLocked()
	now := time.Now()
	o.mu.Lock()
	due := make([]pendingDelivery, 0, len(o.pending))
	keep := make([]pendingDelivery, 0, len(o.pending))
	for _, d := range o.pending {
		if d.nextRetry.After(now) {
			keep = append(keep, d)
			continue
		}
		due = append(due, d)
	}
	o.pending = keep
	o.mu.Unlock()

	for _, d := range due {
		err := o.inner.NotifyAck(ctx, d.signal)
		if err == nil {
			continue
		}
		d.attempts++
		d.lastError = err
		if d.attempts >= o.maxAttempts {
			continue // dropped
		}
		// Exponential backoff capped at MaxBackoff.
		backoff := o.minBackoff * time.Duration(1<<uint(d.attempts-1))
		if backoff > o.maxBackoff {
			backoff = o.maxBackoff
		}
		d.nextRetry = time.Now().Add(backoff)
		o.mu.Lock()
		o.pending = append(o.pending, d)
		o.mu.Unlock()
	}
}

// Compile-time check that Outbox satisfies the Notifier port.
var _ ports.Notifier = (*Outbox)(nil)

// signalSeenAt is unused but reserved so future durable variants can
// add a persisted timestamp without breaking the in-memory shape.
//
//nolint:unused
type signalSeenAt struct {
	at time.Time
}

// helper: a quick way to build a pending row without exposing the
// internal struct. Used from tests.
func newPending(sig domain.Signal, after time.Duration) pendingDelivery {
	return pendingDelivery{signal: sig, attempts: 1, nextRetry: time.Now().Add(after)}
}

// silence "unused" until the durable variant lands.
//
//nolint:unused
func (p pendingDelivery) signalID() uuid.UUID { return p.signal.ID }

// silence "unused".
//
//nolint:unused
var _ = newPending
