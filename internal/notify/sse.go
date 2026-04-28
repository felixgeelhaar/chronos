package notify

import (
	"context"
	"sync"

	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/google/uuid"
)

// SSE is an in-memory broadcaster for Server-Sent Events. It
// implements ports.Notifier (push side) and exposes Subscribe /
// Unsubscribe so an HTTP handler can register browser/runtime clients
// for /v1/signals/stream. Each subscriber owns a buffered channel; a
// signal is delivered with a non-blocking send, so a slow client is
// silently dropped rather than back-pressuring the entire bus. SSE
// is therefore a best-effort live feed; consumers needing replay
// must combine the stream with a /v1/signals query keyed on the
// last-seen DetectedAt.
type SSE struct {
	bufferSize int

	mu          sync.RWMutex
	subscribers map[uuid.UUID]*subscriber
}

type subscriber struct {
	ch      chan domain.Signal
	scope   uuid.UUID // uuid.Nil means "any scope"
	pattern string    // "" means "any pattern"
}

// NewSSE returns a broadcaster with the given per-client buffer
// capacity. A buffer of 8-32 is typically right: enough to absorb a
// burst, small enough that genuinely stuck clients are dropped fast.
// A bufferSize of 0 forces every send to block until the receiver
// reads — so we floor at 1 silently to keep the broadcaster healthy.
func NewSSE(bufferSize int) *SSE {
	if bufferSize < 1 {
		bufferSize = 1
	}
	return &SSE{
		bufferSize:  bufferSize,
		subscribers: make(map[uuid.UUID]*subscriber),
	}
}

// Subscribe registers a new client filtered by scope (uuid.Nil = any
// scope) and pattern (empty = any pattern). It returns:
//   - id used to unregister
//   - read-only channel of domain.Signal matching the filter
//
// The returned channel is closed by Unsubscribe; receivers should
// detect the close and exit their read loop. The signature is
// deliberately bare types so internal/api can declare a matching
// interface without importing this package — that breaks what would
// otherwise be a cycle (api -> notify -> api via the wire shape).
func (s *SSE) Subscribe(scope uuid.UUID, pattern string) (uuid.UUID, <-chan domain.Signal) {
	id := uuid.New()
	sub := &subscriber{
		ch:      make(chan domain.Signal, s.bufferSize),
		scope:   scope,
		pattern: pattern,
	}
	s.mu.Lock()
	s.subscribers[id] = sub
	s.mu.Unlock()
	return id, sub.ch
}

// Unsubscribe removes the client and closes its channel. Safe to call
// with an unknown id (no-op).
func (s *SSE) Unsubscribe(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.subscribers[id]
	if !ok {
		return
	}
	delete(s.subscribers, id)
	close(sub.ch)
}

// Notify fans the signal out to every matching subscriber via a
// non-blocking send. Slow consumers are dropped silently. Honours
// ctx cancellation to skip iteration when the broadcaster is being
// shut down.
func (s *SSE) Notify(ctx context.Context, sig domain.Signal) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sub := range s.subscribers {
		if !matches(sub, sig) {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case sub.ch <- sig:
		default:
			// Drop: client is too slow. The /v1/signals/stream
			// handler will recover by reconnecting.
		}
	}
}

// SubscriberCount returns the number of currently-attached clients.
// Useful for /metrics and tests.
func (s *SSE) SubscriberCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.subscribers)
}

func matches(sub *subscriber, sig domain.Signal) bool {
	if sub.scope != uuid.Nil && sub.scope != sig.ScopeID {
		return false
	}
	if sub.pattern != "" && sub.pattern != string(sig.Pattern) {
		return false
	}
	return true
}
