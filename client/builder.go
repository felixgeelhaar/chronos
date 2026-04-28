package client

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// SignalQuery is a fluent builder for /v1/signals. Construct via
// [Client.Signals]. Chain filters, then call List or Get.
type SignalQuery struct {
	c             *Client
	scope         uuid.UUID
	series        *uuid.UUID
	pattern       *string
	since, until  *time.Time
	minConfidence *float64
	limit         int
}

// Scope filters by scope ID. Required for List.
func (q *SignalQuery) Scope(id uuid.UUID) *SignalQuery {
	q.scope = id
	return q
}

// Series filters to signals about a single entity.
func (q *SignalQuery) Series(id uuid.UUID) *SignalQuery {
	q.series = &id
	return q
}

// Pattern filters to a single PatternType (use the PatternType*
// constants from this package).
func (q *SignalQuery) Pattern(name string) *SignalQuery {
	q.pattern = &name
	return q
}

// Since restricts results to signals detected at or after t.
func (q *SignalQuery) Since(t time.Time) *SignalQuery {
	q.since = &t
	return q
}

// Until restricts results to signals detected before t.
func (q *SignalQuery) Until(t time.Time) *SignalQuery {
	q.until = &t
	return q
}

// MinConfidence drops signals below threshold.
func (q *SignalQuery) MinConfidence(threshold float64) *SignalQuery {
	q.minConfidence = &threshold
	return q
}

// Limit caps the number of returned signals.
func (q *SignalQuery) Limit(n int) *SignalQuery {
	q.limit = n
	return q
}

// List runs the query.
func (q *SignalQuery) List(ctx context.Context) ([]Signal, error) {
	if q.scope == uuid.Nil {
		return nil, errors.New("chronos client: Scope is required")
	}
	v := url.Values{}
	v.Set("scope_id", q.scope.String())
	if q.series != nil {
		v.Set("series", q.series.String())
	}
	if q.pattern != nil {
		v.Set("pattern", *q.pattern)
	}
	if q.since != nil {
		v.Set("since", q.since.UTC().Format(time.RFC3339))
	}
	if q.until != nil {
		v.Set("until", q.until.UTC().Format(time.RFC3339))
	}
	if q.minConfidence != nil {
		v.Set("min_confidence", strconv.FormatFloat(*q.minConfidence, 'f', -1, 64))
	}
	if q.limit > 0 {
		v.Set("limit", strconv.Itoa(q.limit))
	}

	var body struct {
		Signals []Signal `json:"signals"`
		Count   int      `json:"count"`
	}
	if err := q.c.do(ctx, "GET", "/v1/signals?"+v.Encode(), nil, &body); err != nil {
		return nil, err
	}
	return body.Signals, nil
}

// Stream subscribes to /v1/signals/stream and returns a channel that
// emits each newly-detected signal until ctx is cancelled or the
// server closes the stream. Scope is required (the SSE endpoint
// rejects unscoped streams). Pattern is honoured server-side when set.
//
// The returned channel is closed by the SDK on ctx cancellation,
// connection drop, or fatal protocol error — `for sig := range ch`
// is safe.
//
// Streaming bypasses the client's Timeout so connections can outlive
// the configured per-request deadline; cancel via ctx instead.
//
// For at-most-once gap recovery, pair the stream with a Since-keyed
// List call: the persisted /v1/signals query is the source of truth,
// the stream is a courtesy. De-duplicate by Signal.ID.
func (q *SignalQuery) Stream(ctx context.Context) (<-chan Signal, error) {
	if q.scope == uuid.Nil {
		return nil, errors.New("chronos client: Scope is required for Stream")
	}
	v := url.Values{}
	v.Set("scope_id", q.scope.String())
	if q.pattern != nil {
		v.Set("pattern", *q.pattern)
	}
	return q.c.openStream(ctx, "/v1/signals/stream?"+v.Encode())
}

// Get returns a single signal by ID.
func (q *SignalQuery) Get(ctx context.Context, id uuid.UUID) (Signal, error) {
	var out Signal
	if err := q.c.do(ctx, "GET", fmt.Sprintf("/v1/signals/%s", id.String()), nil, &out); err != nil {
		return Signal{}, err
	}
	return out, nil
}
