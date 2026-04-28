package notify

import (
	"context"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

func TestSSE_BroadcastsToAllSubscribers(t *testing.T) {
	s := NewSSE(8)
	_, chA := s.Subscribe(uuid.Nil, "")
	_, chB := s.Subscribe(uuid.Nil, "")

	sig := sampleSignal()
	s.Notify(context.Background(), sig)

	for i, ch := range []<-chan domain.Signal{chA, chB} {
		select {
		case got := <-ch:
			if got.ID != sig.ID {
				t.Errorf("subscriber %d got wrong signal", i)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d did not receive signal", i)
		}
	}
}

func TestSSE_FilterByScope(t *testing.T) {
	s := NewSSE(4)
	scopeA := uuid.New()
	scopeB := uuid.New()

	_, chA := s.Subscribe(scopeA, "")
	_, chB := s.Subscribe(scopeB, "")

	sigA := sampleSignal()
	sigA.ScopeID = scopeA
	s.Notify(context.Background(), sigA)

	select {
	case got := <-chA:
		if got.ScopeID != scopeA {
			t.Errorf("scope-A subscriber got wrong scope %v", got.ScopeID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("scope-A subscriber missed its signal")
	}

	select {
	case unwanted := <-chB:
		t.Fatalf("scope-B subscriber should not have received scope-A signal: %v", unwanted)
	case <-time.After(50 * time.Millisecond):
		// Expected — no signal.
	}
}

func TestSSE_FilterByPattern(t *testing.T) {
	s := NewSSE(4)
	_, chR := s.Subscribe(uuid.Nil, string(domain.PatternTypeRecurrence))
	_, chT := s.Subscribe(uuid.Nil, string(domain.PatternTypeTrend))

	sig := sampleSignal() // Pattern = recurrence
	s.Notify(context.Background(), sig)

	select {
	case <-chR:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("recurrence subscriber missed signal")
	}
	select {
	case unwanted := <-chT:
		t.Fatalf("trend subscriber should not have received: %v", unwanted)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestSSE_SlowConsumerIsDropped(t *testing.T) {
	s := NewSSE(1) // tiny buffer so we can overflow easily
	_, ch := s.Subscribe(uuid.Nil, "")

	// Fill the buffer.
	s.Notify(context.Background(), sampleSignal())
	// Without anyone reading, the next Notify should drop silently
	// rather than block.
	done := make(chan struct{})
	go func() {
		s.Notify(context.Background(), sampleSignal())
		close(done)
	}()
	select {
	case <-done:
		// Notify returned without blocking — dropped on a slow consumer.
	case <-time.After(time.Second):
		t.Fatal("Notify blocked on a slow consumer instead of dropping")
	}
	// Drain the buffered one so the test goroutine can exit.
	<-ch
}

func TestSSE_UnsubscribeClosesChannel(t *testing.T) {
	s := NewSSE(4)
	id, ch := s.Subscribe(uuid.Nil, "")
	if s.SubscriberCount() != 1 {
		t.Fatalf("subscriber count should be 1, got %d", s.SubscriberCount())
	}
	s.Unsubscribe(id)
	if s.SubscriberCount() != 0 {
		t.Fatalf("subscriber count should be 0 after unsubscribe, got %d", s.SubscriberCount())
	}
	// Channel should be closed.
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("channel had a value after unsubscribe; expected closed")
		}
	case <-time.After(time.Second):
		t.Fatal("channel was not closed by unsubscribe")
	}
}

func TestSSE_UnsubscribeUnknownIDIsNoop(t *testing.T) {
	s := NewSSE(4)
	s.Unsubscribe(uuid.New()) // must not panic, must not corrupt state
	if s.SubscriberCount() != 0 {
		t.Fatalf("count should be 0, got %d", s.SubscriberCount())
	}
}

func TestSSE_BufferSizeFlooredAtOne(t *testing.T) {
	s := NewSSE(0)
	if s.bufferSize != 1 {
		t.Fatalf("expected buffer floored at 1, got %d", s.bufferSize)
	}
}
