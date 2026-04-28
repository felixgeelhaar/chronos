package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/google/uuid"
)

func openTestConn(t *testing.T) *Conn {
	t.Helper()
	c, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestSQLite_EntityStateRoundTrip(t *testing.T) {
	c := openTestConn(t)
	ctx := context.Background()
	scope := uuid.New()
	a := uuid.New()
	b := uuid.New()
	now := time.Now()

	if err := c.EntityStates.Ingest(ctx, "test", chronos.EntityState{ID: uuid.New(), EntityID: a, ScopeID: scope, Timestamp: now, Features: []float64{1, 2, 3, 5}}); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if err := c.EntityStates.Save(ctx, "test", []chronos.EntityState{
		{ID: uuid.New(), EntityID: b, ScopeID: scope, Timestamp: now.Add(-time.Hour), Features: []float64{1.1, 2.1, 3.1, 7}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := c.EntityStates.ListByScope(ctx, scope)
	if err != nil {
		t.Fatalf("ListByScope: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListByScope length = %d, want 2", len(got))
	}
	if !got[0].Timestamp.After(got[1].Timestamp) {
		t.Errorf("ListByScope must be most-recent-first")
	}

	count, _ := c.EntityStates.Count(ctx, "test")
	if count != 2 {
		t.Errorf("Count = %d, want 2", count)
	}
}

func mkSignal(scope, series uuid.UUID, pattern domain.PatternType, conf float64, t time.Time) domain.Signal {
	return domain.Signal{
		ID:         uuid.New(),
		ScopeID:    scope,
		Series:     series,
		Pattern:    pattern,
		DetectedAt: t,
		Window:     domain.TimeWindow{Start: t.Add(-time.Hour), End: t},
		Strength:   conf,
		Confidence: conf,
		Metrics:    map[string]float64{"avg_similarity": conf},
	}
}

func TestSQLite_SignalRoundTripWithEvidence(t *testing.T) {
	c := openTestConn(t)
	ctx := context.Background()
	scope := uuid.New()
	series := uuid.New()
	now := time.Now()

	sig := mkSignal(scope, series, domain.PatternTypeRecurrence, 0.85, now)
	sig.Evidence = []domain.Evidence{
		{Series: uuid.New(), Time: now.Add(-time.Hour), Kind: "similar_state", Score: 0.92, Metrics: map[string]float64{"outcome_diff": 1.5}},
	}
	if err := c.Signals.Save(ctx, sig); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := c.Signals.Get(ctx, sig.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Pattern != domain.PatternTypeRecurrence {
		t.Errorf("Pattern lost: %s", got.Pattern)
	}
	if len(got.Evidence) != 1 {
		t.Fatalf("evidence count = %d", len(got.Evidence))
	}
	if got.Evidence[0].Kind != "similar_state" {
		t.Errorf("evidence kind = %q", got.Evidence[0].Kind)
	}
	if got.Evidence[0].Metrics["outcome_diff"] != 1.5 {
		t.Errorf("evidence metric lost: %v", got.Evidence[0].Metrics)
	}
}

func TestSQLite_SignalListFilters(t *testing.T) {
	c := openTestConn(t)
	ctx := context.Background()
	scope := uuid.New()
	other := uuid.New()
	now := time.Now()

	in := []domain.Signal{
		mkSignal(scope, uuid.New(), domain.PatternTypeRecurrence, 0.9, now),
		mkSignal(scope, uuid.New(), domain.PatternTypeRecurrence, 0.7, now.Add(-time.Hour)),
		mkSignal(scope, uuid.New(), domain.PatternTypeTrend, 0.6, now.Add(-2*time.Hour)),
		mkSignal(other, uuid.New(), domain.PatternTypeRecurrence, 0.95, now),
	}
	for _, s := range in {
		if err := c.Signals.Save(ctx, s); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	got, _ := c.Signals.List(ctx, ports.SignalFilter{ScopeID: scope})
	if len(got) != 3 {
		t.Errorf("scope filter returned %d, want 3", len(got))
	}
	if !got[0].DetectedAt.After(got[1].DetectedAt) {
		t.Errorf("ordering must be detected-at desc")
	}

	pat := domain.PatternTypeRecurrence
	got, _ = c.Signals.List(ctx, ports.SignalFilter{ScopeID: scope, Pattern: &pat})
	if len(got) != 2 {
		t.Errorf("pattern filter returned %d, want 2", len(got))
	}

	conf := 0.8
	got, _ = c.Signals.List(ctx, ports.SignalFilter{ScopeID: scope, MinConfidence: &conf})
	if len(got) != 1 {
		t.Errorf("min confidence filter returned %d, want 1", len(got))
	}

	got, _ = c.Signals.List(ctx, ports.SignalFilter{ScopeID: scope, Limit: 1})
	if len(got) != 1 {
		t.Errorf("limit returned %d, want 1", len(got))
	}

	n, _ := c.Signals.Count(ctx, ports.SignalFilter{ScopeID: scope})
	if n != 3 {
		t.Errorf("Count = %d, want 3", n)
	}
}

func TestSQLite_SignalGetMissing(t *testing.T) {
	c := openTestConn(t)
	if _, err := c.Signals.Get(context.Background(), uuid.New()); !errors.Is(err, domain.ErrSignalNotFound) {
		t.Errorf("Get missing = %v, want ErrSignalNotFound", err)
	}
}
