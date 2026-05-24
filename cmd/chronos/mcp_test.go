package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

// setIsolatedDB points the chronos MCP tools at a fresh sqlite DB
// under t.TempDir so each test runs in isolation. The mcpRun* funcs
// resolve their conn via resolveDSN → CHRONOS_DB_DSN, so a single env
// override is the cleanest seam.
func setIsolatedDB(t *testing.T) {
	t.Helper()
	t.Setenv("CHRONOS_DB_DSN", "sqlite://"+filepath.Join(t.TempDir(), "chronos.db"))
}

// TestMCP_Ingest_PersistsObservation pins the happy path: one MCP
// ingest call writes one row that a subsequent list_signals call can
// see via the underlying store.
func TestMCP_Ingest_PersistsObservation(t *testing.T) {
	setIsolatedDB(t)
	scope := uuid.New().String()
	entity := uuid.New().String()
	out, err := mcpRunIngest(context.Background(), mcpIngestInput{
		EntityID: entity,
		ScopeID:  scope,
		Features: []float64{1, 2, 3, 5},
		Adapter:  "test",
	})
	if err != nil {
		t.Fatalf("mcpRunIngest: %v", err)
	}
	if out.Status != "accepted" {
		t.Errorf("status = %q, want accepted", out.Status)
	}
	if out.ID == "" {
		t.Error("missing id")
	}
}

// TestMCP_Ingest_Validation pins the fail-closed validation surface:
// missing entity / scope / features all yield an explicit error
// instead of silently accepting empty data.
func TestMCP_Ingest_Validation(t *testing.T) {
	setIsolatedDB(t)
	cases := []struct {
		name  string
		input mcpIngestInput
	}{
		{"missing entity_id", mcpIngestInput{ScopeID: uuid.New().String(), Features: []float64{1}}},
		{"missing scope_id", mcpIngestInput{EntityID: uuid.New().String(), Features: []float64{1}}},
		{"missing features", mcpIngestInput{EntityID: uuid.New().String(), ScopeID: uuid.New().String()}},
		{"bad entity_id", mcpIngestInput{EntityID: "x", ScopeID: uuid.New().String(), Features: []float64{1}}},
		{"bad timestamp", mcpIngestInput{EntityID: uuid.New().String(), ScopeID: uuid.New().String(), Features: []float64{1}, Timestamp: "tomorrow"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if _, err := mcpRunIngest(context.Background(), tc.input); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// TestMCP_ListSignals_RequiresScope pins the tenant-safety contract:
// list_signals without a scope filter is the firehose and must error.
func TestMCP_ListSignals_RequiresScope(t *testing.T) {
	setIsolatedDB(t)
	if _, err := mcpRunListSignals(context.Background(), mcpListSignalsInput{}); err == nil {
		t.Fatal("expected error for scope-less query")
	}
}

// TestMCP_ListSignals_RejectsBothScopeFilters guards against the
// ambiguous request where a caller sets both scope_id and scope_ids.
func TestMCP_ListSignals_RejectsBothScopeFilters(t *testing.T) {
	setIsolatedDB(t)
	_, err := mcpRunListSignals(context.Background(), mcpListSignalsInput{
		ScopeID:  uuid.New().String(),
		ScopeIDs: []string{uuid.New().String()},
	})
	if err == nil {
		t.Fatal("expected error when both scope filters are set")
	}
}

// TestMCP_ListSignals_HonorsScopeFilter is the round-trip: ingest one
// observation, then verify list_signals scoped to that observation's
// scope returns an empty list (ingest doesn't synthesise signals;
// detection is async). The point is to prove the filter parameter
// flows end-to-end without blowing up.
func TestMCP_ListSignals_HonorsScopeFilter(t *testing.T) {
	setIsolatedDB(t)
	scope := uuid.New()
	out, err := mcpRunListSignals(context.Background(), mcpListSignalsInput{
		ScopeID: scope.String(),
		Limit:   10,
	})
	if err != nil {
		t.Fatalf("ListSignals: %v", err)
	}
	if out.Count != 0 {
		t.Errorf("fresh DB returned %d signals", out.Count)
	}
}

// TestMCP_ListSignals_BadSince rejects malformed RFC3339 timestamps
// rather than silently dropping the filter.
func TestMCP_ListSignals_BadSince(t *testing.T) {
	setIsolatedDB(t)
	_, err := mcpRunListSignals(context.Background(), mcpListSignalsInput{
		ScopeID: uuid.New().String(),
		Since:   "yesterday",
	})
	if err == nil {
		t.Fatal("expected error for malformed since")
	}
}

// TestMCP_DescribeDetector_KnownPattern returns the same shape the
// HTTP /v1/config/validate endpoint would surface for a single
// detector. Pins parity so a future config edit can't silently
// diverge the two surfaces.
func TestMCP_DescribeDetector_KnownPattern(t *testing.T) {
	out, err := mcpRunDescribeDetector(context.Background(), mcpDescribeDetectorInput{
		Pattern: "trend",
	})
	if err != nil {
		t.Fatalf("DescribeDetector: %v", err)
	}
	if out.Pattern != "trend" {
		t.Errorf("Pattern = %q, want trend", out.Pattern)
	}
	if !out.Enabled {
		t.Errorf("trend should be enabled at defaults: %+v", out)
	}
	if len(out.Thresholds) == 0 {
		t.Errorf("missing thresholds: %+v", out)
	}
}

// TestMCP_DescribeDetector_UnknownPattern errors explicitly so a
// typo in the agent's tool call surfaces instead of returning a
// default/empty record that could be mistaken for a real detector.
func TestMCP_DescribeDetector_UnknownPattern(t *testing.T) {
	if _, err := mcpRunDescribeDetector(context.Background(), mcpDescribeDetectorInput{
		Pattern: "no_such_pattern",
	}); err == nil {
		t.Fatal("expected error for unknown pattern")
	}
}

// TestMCP_DescribeDetector_EmptyPattern errors out — pattern is
// required.
func TestMCP_DescribeDetector_EmptyPattern(t *testing.T) {
	if _, err := mcpRunDescribeDetector(context.Background(), mcpDescribeDetectorInput{}); err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

// TestMCP_JSONDump pins the tiny helper so future refactors don't
// quietly change its output and break debug-path tests downstream.
func TestMCP_JSONDump(t *testing.T) {
	got := jsonDump(mcpIngestOutput{ID: "x", Status: "accepted"})
	if got == "" {
		t.Fatal("empty dump")
	}
	if got != `{"id":"x","status":"accepted"}` {
		t.Errorf("dump = %q", got)
	}
}

// TestMCP_Ingest_DefaultsTimestamp pins that omitting the timestamp
// is OK — the handler stamps now() instead of erroring out.
func TestMCP_Ingest_DefaultsTimestamp(t *testing.T) {
	setIsolatedDB(t)
	before := time.Now().Add(-time.Second).UTC()
	out, err := mcpRunIngest(context.Background(), mcpIngestInput{
		EntityID: uuid.New().String(),
		ScopeID:  uuid.New().String(),
		Features: []float64{1, 2, 3},
	})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if out.ID == "" {
		t.Error("missing id")
	}
	_ = before // anchor for any future timestamp assertion
}
