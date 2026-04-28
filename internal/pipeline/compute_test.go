package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/felixgeelhaar/chronos/internal/store/memory"
	"github.com/google/uuid"
)

// fakeSource is a chronos.Source for tests; returning err causes Fetch to
// fail with that error, otherwise the configured states are returned.
type fakeSource struct {
	name   string
	states []chronos.EntityState
	err    error
}

func (s *fakeSource) Name() string { return s.name }
func (s *fakeSource) Fetch(_ context.Context, _ map[string]string) ([]chronos.EntityState, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.states, nil
}

func recurrenceCfg() *config.Config {
	return &config.Config{
		MaxSignalsPerRun:    100,
		SimilarityThreshold: 0.8,
		MinSampleSize:       2,
	}
}

func TestCompute_HappyPathPersistsStatesAndSignals(t *testing.T) {
	mem := memory.New()
	scope := uuid.New()
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	now := time.Now()
	states := []chronos.EntityState{
		{ID: uuid.New(), EntityID: a, ScopeID: scope, Timestamp: now, Features: []float64{1, 2, 3, 5}},
		{ID: uuid.New(), EntityID: b, ScopeID: scope, Timestamp: now.Add(-time.Hour), Features: []float64{1.1, 2.1, 3.1, 7}},
		{ID: uuid.New(), EntityID: c, ScopeID: scope, Timestamp: now.Add(-2 * time.Hour), Features: []float64{1.05, 2.05, 3.05, 6.5}},
	}

	res, err := Compute(context.Background(), ComputeInput{
		Source:       &fakeSource{name: "test", states: states},
		EntityStates: mem.EntityStates,
		Signals:      mem.Signals,
		Engine:       NewEngine(recurrenceCfg()),
	})
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	if res.StatesFetched != 3 {
		t.Errorf("StatesFetched = %d, want 3", res.StatesFetched)
	}
	if res.SignalsCreated == 0 {
		t.Errorf("SignalsCreated = 0; expected at least one Recurrence signal")
	}

	// States must be persisted under the source's name.
	stored, err := mem.EntityStates.ListByScope(context.Background(), scope)
	if err != nil {
		t.Fatalf("ListByScope: %v", err)
	}
	if len(stored) != 3 {
		t.Errorf("persisted states = %d, want 3", len(stored))
	}
	count, _ := mem.EntityStates.Count(context.Background(), "test")
	if count != 3 {
		t.Errorf("Count(test) = %d, want 3", count)
	}

	// Signals must be persisted and discoverable.
	n, _ := mem.Signals.Count(context.Background(), ports.SignalFilter{ScopeID: scope})
	if n != int64(res.SignalsCreated) {
		t.Errorf("persisted signals = %d, want %d", n, res.SignalsCreated)
	}
}

func TestCompute_NilSourceReturnsError(t *testing.T) {
	mem := memory.New()
	_, err := Compute(context.Background(), ComputeInput{
		EntityStates: mem.EntityStates,
		Signals:      mem.Signals,
		Engine:       NewEngine(recurrenceCfg()),
	})
	if err == nil {
		t.Fatal("expected error for nil source")
	}
}

func TestCompute_NilEngineReturnsError(t *testing.T) {
	mem := memory.New()
	_, err := Compute(context.Background(), ComputeInput{
		Source:       &fakeSource{name: "test"},
		EntityStates: mem.EntityStates,
		Signals:      mem.Signals,
	})
	if err == nil {
		t.Fatal("expected error for nil engine")
	}
}

func TestCompute_FetchErrorIsWrapped(t *testing.T) {
	mem := memory.New()
	sentinel := errors.New("upstream is down")
	_, err := Compute(context.Background(), ComputeInput{
		Source:       &fakeSource{name: "test", err: sentinel},
		EntityStates: mem.EntityStates,
		Signals:      mem.Signals,
		Engine:       NewEngine(recurrenceCfg()),
	})
	if err == nil {
		t.Fatal("expected error from fetch failure")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("error chain does not wrap sentinel: %v", err)
	}

	// Nothing should have been persisted on a fetch failure.
	count, _ := mem.EntityStates.Count(context.Background(), "test")
	if count != 0 {
		t.Errorf("states persisted despite fetch error: %d", count)
	}
}

func TestCompute_EmptyStatesReturnsZeros(t *testing.T) {
	mem := memory.New()
	res, err := Compute(context.Background(), ComputeInput{
		Source:       &fakeSource{name: "test", states: nil},
		EntityStates: mem.EntityStates,
		Signals:      mem.Signals,
		Engine:       NewEngine(recurrenceCfg()),
	})
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	if res.StatesFetched != 0 || res.SignalsCreated != 0 {
		t.Errorf("expected zero counts, got %+v", res)
	}
}
