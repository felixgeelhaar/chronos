package detect

import (
	"context"
	"testing"

	"github.com/felixgeelhaar/chronos/internal/config"
	"github.com/google/uuid"
)

// TestAnonymizeID_Deterministic pins that the same input always
// hashes to the same opaque uuid — consumers need that to group
// anonymous-signal pairs across runs.
func TestAnonymizeID_Deterministic(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	if anonymizeID(id) != anonymizeID(id) {
		t.Errorf("anonymizeID not deterministic for %s", id)
	}
}

// TestAnonymizeID_DistinctInputsDistinctOutputs guards against the
// hash function silently collapsing two real scopes onto the same
// opaque uuid. A collision wouldn't be a security hole, but it would
// destroy the population-science usefulness.
func TestAnonymizeID_DistinctInputsDistinctOutputs(t *testing.T) {
	t.Parallel()
	a, b := uuid.New(), uuid.New()
	if anonymizeID(a) == anonymizeID(b) {
		t.Errorf("collision: anonymizeID(%s) == anonymizeID(%s)", a, b)
	}
}

// TestAnonymizeID_NotIdentity proves the hash is not a pass-through —
// the whole point is that the output != the input.
func TestAnonymizeID_NotIdentity(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	if anonymizeID(id) == id {
		t.Errorf("anonymizeID is identity for %s", id)
	}
}

// TestCrossScopeCorrelation_AnonymizeRedactsIdentity proves the
// detector replaces scope/series with opaque hashes when the config
// flag is on, while preserving the statistical payload (r, n, etc.).
func TestCrossScopeCorrelation_AnonymizeRedactsIdentity(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		CrossScopeMin:       0.8,
		CrossScopeMinPoints: 5,
		AnonymizeCrossScope: true,
	}
	scopeA, scopeB := uuid.New(), uuid.New()
	seriesA, seriesB := uuid.New(), uuid.New()
	a := cscSeries(scopeA, seriesA, []float64{1, 2, 3, 4, 5, 6, 7, 8})
	b := cscSeries(scopeB, seriesB, []float64{2, 4, 6, 8, 10, 12, 14, 16})
	states := append(a, b...)
	d := NewCrossScopeCorrelation(cfg)

	signals := d.CrossDetect(context.Background(), states)
	if len(signals) != 1 {
		t.Fatalf("len = %d, want 1", len(signals))
	}
	s := signals[0]
	// ScopeID + Series + evidence series must NOT match the raw inputs.
	if s.ScopeID == scopeA || s.ScopeID == scopeB {
		t.Errorf("ScopeID %s leaks a real scope", s.ScopeID)
	}
	if s.Series == seriesA || s.Series == seriesB {
		t.Errorf("Series %s leaks a real series", s.Series)
	}
	if len(s.Evidence) != 1 {
		t.Fatalf("expected 1 evidence row, got %d", len(s.Evidence))
	}
	if s.Evidence[0].Series == seriesA || s.Evidence[0].Series == seriesB {
		t.Errorf("evidence.Series %s leaks a real series", s.Evidence[0].Series)
	}
	// Hashed ids must equal anonymizeID() of the lex-smaller side. The
	// detector emits one signal per pair owned by the lex-smaller
	// (scope, series), so a hardcoded expectation would over-constrain;
	// we assert the value matches one of the candidates.
	if s.ScopeID != anonymizeID(scopeA) && s.ScopeID != anonymizeID(scopeB) {
		t.Errorf("ScopeID %s is neither anonymizeID(scopeA) nor anonymizeID(scopeB)", s.ScopeID)
	}
	// Statistical payload survives anonymization.
	if s.Metrics["n"] == 0 {
		t.Errorf("n metric lost: %+v", s.Metrics)
	}
	if _, ok := s.Metrics["r"]; !ok {
		t.Errorf("r metric lost: %+v", s.Metrics)
	}
	if _, ok := s.Metrics["direction"]; !ok {
		t.Errorf("direction metric lost: %+v", s.Metrics)
	}
	if s.Strength == 0 || s.Confidence == 0 {
		t.Errorf("strength/confidence lost: %+v", s)
	}
	if v := s.Metrics["anonymized"]; v != 1 {
		t.Errorf("anonymized metric not set, got %v", s.Metrics)
	}
}

// TestCrossScopeCorrelation_CleartextModeKeepsRealIDs is the inverse
// guardrail: the default config must still emit real scope/series so
// operators who consciously accept the cross-tenant risk don't lose
// the ability to correlate across runs.
func TestCrossScopeCorrelation_CleartextModeKeepsRealIDs(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		CrossScopeMin:       0.8,
		CrossScopeMinPoints: 5,
		AnonymizeCrossScope: false,
	}
	scopeA, scopeB := uuid.New(), uuid.New()
	seriesA, seriesB := uuid.New(), uuid.New()
	a := cscSeries(scopeA, seriesA, []float64{1, 2, 3, 4, 5, 6, 7, 8})
	b := cscSeries(scopeB, seriesB, []float64{2, 4, 6, 8, 10, 12, 14, 16})
	states := append(a, b...)
	d := NewCrossScopeCorrelation(cfg)

	signals := d.CrossDetect(context.Background(), states)
	if len(signals) != 1 {
		t.Fatalf("len = %d", len(signals))
	}
	s := signals[0]
	// Owning side is the lex-smaller scope.
	if s.ScopeID != scopeA && s.ScopeID != scopeB {
		t.Errorf("ScopeID = %s, expected a real scope", s.ScopeID)
	}
	if s.Series != seriesA && s.Series != seriesB {
		t.Errorf("Series = %s, expected a real series", s.Series)
	}
	if v := s.Metrics["anonymized"]; v != 0 {
		t.Errorf("anonymized metric should be absent or 0 in cleartext mode, got %v", v)
	}
}
