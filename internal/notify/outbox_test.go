package notify

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

// flakyNotifier fails the first N calls then succeeds.
type flakyNotifier struct {
	mu        sync.Mutex
	failsLeft int32
	calls     int32
}

func (f *flakyNotifier) NotifyAck(_ context.Context, _ domain.Signal) error {
	atomic.AddInt32(&f.calls, 1)
	if atomic.AddInt32(&f.failsLeft, -1) >= 0 {
		return errors.New("transient")
	}
	return nil
}

func (f *flakyNotifier) Calls() int32 { return atomic.LoadInt32(&f.calls) }

func mkSignal() domain.Signal {
	return domain.Signal{ID: uuid.New()}
}

func TestOutbox_RetriesUntilSuccess(t *testing.T) {
	t.Parallel()
	flaky := &flakyNotifier{failsLeft: 2}
	o, err := NewOutbox(flaky, OutboxConfig{
		MaxAttempts: 5,
		MinBackoff:  10 * time.Millisecond,
		MaxBackoff:  20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	o.Start(ctx)
	defer func() { cancel(); o.Stop() }()

	o.Notify(context.Background(), mkSignal())
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if o.PendingCount() == 0 && flaky.Calls() >= 3 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Errorf("did not converge: pending=%d calls=%d", o.PendingCount(), flaky.Calls())
}

func TestOutbox_GivesUpAfterMaxAttempts(t *testing.T) {
	t.Parallel()
	flaky := &flakyNotifier{failsLeft: 1000}
	o, _ := NewOutbox(flaky, OutboxConfig{
		MaxAttempts: 3,
		MinBackoff:  5 * time.Millisecond,
		MaxBackoff:  5 * time.Millisecond,
	})
	ctx, cancel := context.WithCancel(context.Background())
	o.Start(ctx)
	defer func() { cancel(); o.Stop() }()

	o.Notify(context.Background(), mkSignal())
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if o.PendingCount() == 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if o.PendingCount() != 0 {
		t.Errorf("still pending: %d", o.PendingCount())
	}
	if flaky.Calls() != 3 {
		t.Errorf("calls = %d, want 3 (max attempts)", flaky.Calls())
	}
}

func TestOutbox_SuccessOnFirstSkipsRetry(t *testing.T) {
	t.Parallel()
	flaky := &flakyNotifier{}
	o, _ := NewOutbox(flaky, OutboxConfig{MaxAttempts: 5, MinBackoff: 10 * time.Millisecond, MaxBackoff: 20 * time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	o.Start(ctx)
	defer func() { cancel(); o.Stop() }()

	o.Notify(context.Background(), mkSignal())
	if o.PendingCount() != 0 {
		t.Errorf("pending = %d, want 0", o.PendingCount())
	}
	if flaky.Calls() != 1 {
		t.Errorf("calls = %d, want 1", flaky.Calls())
	}
}

func TestOutbox_RejectsNilInner(t *testing.T) {
	t.Parallel()
	if _, err := NewOutbox(nil, OutboxConfig{}); err == nil {
		t.Fatal("want error")
	}
}

func TestOutbox_PersistenceRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := dir + "/outbox.json"

	// First instance: enqueue one failing delivery.
	flaky := &flakyNotifier{failsLeft: 1000}
	o1, err := NewOutbox(flaky, OutboxConfig{
		MaxAttempts:     5,
		MinBackoff:      time.Hour, // long enough that the sweeper doesn't drain
		MaxBackoff:      time.Hour,
		PersistencePath: path,
	})
	if err != nil {
		t.Fatalf("new1: %v", err)
	}
	sig := mkSignal()
	o1.Notify(context.Background(), sig)
	if o1.PendingCount() != 1 {
		t.Fatalf("pending = %d, want 1", o1.PendingCount())
	}

	// Second instance: must load the pending row from disk.
	flaky2 := &flakyNotifier{failsLeft: 0}
	o2, err := NewOutbox(flaky2, OutboxConfig{
		MaxAttempts:     5,
		MinBackoff:      time.Hour,
		MaxBackoff:      time.Hour,
		PersistencePath: path,
	})
	if err != nil {
		t.Fatalf("new2: %v", err)
	}
	if o2.PendingCount() != 1 {
		t.Fatalf("o2 pending = %d, want 1 (restored from disk)", o2.PendingCount())
	}
}
