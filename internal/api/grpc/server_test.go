package grpc

import (
	"context"
	"testing"
	"time"

	chronosv1 "github.com/felixgeelhaar/chronos/api/proto/chronos/v1"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/store/memory"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupServer(t *testing.T) (*Server, *memory.Conn) {
	t.Helper()
	mem := memory.New()
	return NewServer(mem.EntityStates, mem.Signals, nil, nil), mem
}

func TestIngest_PersistsState(t *testing.T) {
	srv, mem := setupServer(t)
	ctx := context.Background()

	scope := uuid.New().String()
	entity := uuid.New().String()
	resp, err := srv.Ingest(ctx, &chronosv1.IngestRequest{
		EntityId: entity,
		ScopeId:  scope,
		Features: []float64{1, 2, 3, 5},
		Adapter:  "test",
	})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if resp.Status != "accepted" {
		t.Fatalf("status = %q", resp.Status)
	}
	got, _ := mem.EntityStates.ListByScope(ctx, uuid.MustParse(scope))
	if len(got) != 1 || got[0].EntityID.String() != entity {
		t.Fatalf("state not persisted: %+v", got)
	}
}

func TestIngest_RejectsMissingEntityID(t *testing.T) {
	srv, _ := setupServer(t)
	ctx := context.Background()

	_, err := srv.Ingest(ctx, &chronosv1.IngestRequest{
		ScopeId:  uuid.New().String(),
		Features: []float64{1, 2, 3},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestIngest_RejectsMissingFeatures(t *testing.T) {
	srv, _ := setupServer(t)
	ctx := context.Background()

	_, err := srv.Ingest(ctx, &chronosv1.IngestRequest{
		EntityId: uuid.New().String(),
		ScopeId:  uuid.New().String(),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestListSignals_FiltersAndReturnsSignals(t *testing.T) {
	srv, mem := setupServer(t)
	ctx := context.Background()

	scope := uuid.New()
	now := time.Now()
	for _, conf := range []float64{0.6, 0.8, 0.95} {
		sig := domain.Signal{
			ID:         uuid.New(),
			ScopeID:    scope,
			Series:     uuid.New(),
			Pattern:    domain.PatternTypeRecurrence,
			DetectedAt: now,
			Window:     domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
			Strength:   conf,
			Confidence: conf,
			Metrics:    map[string]float64{"avg_similarity": conf},
		}
		if err := mem.Signals.Save(ctx, sig); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	resp, err := srv.ListSignals(ctx, &chronosv1.ListSignalsRequest{
		ScopeId:       scope.String(),
		MinConfidence: 0.7,
	})
	if err != nil {
		t.Fatalf("ListSignals: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("count = %d, want 2", resp.Count)
	}
	for _, s := range resp.Signals {
		if s.Pattern != chronosv1.PatternType_PATTERN_TYPE_RECURRENCE {
			t.Errorf("pattern = %v", s.Pattern)
		}
		if s.Confidence < 0.7 {
			t.Errorf("confidence %f below filter threshold", s.Confidence)
		}
	}
}

func TestListSignals_BadScope(t *testing.T) {
	srv, _ := setupServer(t)
	ctx := context.Background()

	_, err := srv.ListSignals(ctx, &chronosv1.ListSignalsRequest{
		ScopeId: "not-a-uuid",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestListSignals_MissingScope(t *testing.T) {
	srv, _ := setupServer(t)
	ctx := context.Background()

	_, err := srv.ListSignals(ctx, &chronosv1.ListSignalsRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", st.Code())
	}
}

// TestListSignals_ScopeIDsAllowlist pins the multi-scope tenant
// boundary: scope_ids returns only signals from explicitly listed
// scopes, never from scopes outside it. Same contract as the HTTP
// scope_in parameter — gRPC parity.
func TestListSignals_ScopeIDsAllowlist(t *testing.T) {
	srv, mem := setupServer(t)
	ctx := context.Background()

	scopeA := uuid.New()
	scopeB := uuid.New()
	scopeC := uuid.New() // outside allowlist
	now := time.Now()

	for _, sc := range []uuid.UUID{scopeA, scopeB, scopeC} {
		if err := mem.Signals.Save(ctx, domain.Signal{
			ID: uuid.New(), ScopeID: sc, Series: uuid.New(),
			Pattern:    domain.PatternTypeRecurrence,
			DetectedAt: now,
			Window:     domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
			Strength:   0.9, Confidence: 0.9,
		}); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	resp, err := srv.ListSignals(ctx, &chronosv1.ListSignalsRequest{
		ScopeIds: []string{scopeA.String(), scopeB.String()},
	})
	if err != nil {
		t.Fatalf("ListSignals: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("count = %d, want 2 (allowlist contains 2 scopes)", resp.Count)
	}
	for _, s := range resp.Signals {
		if s.ScopeId == scopeC.String() {
			t.Errorf("response leaked signal from scope outside allowlist")
		}
	}
}

// TestListSignals_InvalidScopeIDsEntry fails closed on malformed
// uuids rather than silently dropping them — a silently-dropped entry
// would shrink the intended allowlist without telling the caller.
func TestListSignals_InvalidScopeIDsEntry(t *testing.T) {
	srv, _ := setupServer(t)
	ctx := context.Background()

	_, err := srv.ListSignals(ctx, &chronosv1.ListSignalsRequest{
		ScopeIds: []string{uuid.New().String(), "not-a-uuid"},
	})
	if err == nil {
		t.Fatal("expected error for malformed scope_ids entry")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", st.Code())
	}
}

// TestGetSignal_ExplanationSurfaced pins gRPC parity with the HTTP DTO:
// detector-side context (FeatureEvolution, ComparablePeers,
// BaselineWindowDays, ThresholdUsed, DetectorVersion) survives the
// round-trip so downstream narrators get the same WHY-payload over
// either transport.
func TestGetSignal_ExplanationSurfaced(t *testing.T) {
	srv, mem := setupServer(t)
	ctx := context.Background()

	scope := uuid.New()
	now := time.Now()
	sig := domain.Signal{
		ID:         uuid.New(),
		ScopeID:    scope,
		Series:     uuid.New(),
		Pattern:    domain.PatternTypeTrend,
		DetectedAt: now,
		Window:     domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
		Strength:   0.8, Confidence: 0.8,
		Explanation: domain.Explanation{
			FeatureEvolution: []domain.FeatureSample{
				{At: now.Add(-2 * time.Hour), Value: 18.0},
				{At: now.Add(-time.Hour), Value: 22.0},
				{At: now, Value: 26.0},
			},
			ComparablePeers:    12,
			BaselineWindowDays: 90,
			ThresholdUsed:      2.5,
			DetectorVersion:    "trend-v2",
		},
	}
	if err := mem.Signals.Save(ctx, sig); err != nil {
		t.Fatalf("Save: %v", err)
	}

	resp, err := srv.GetSignal(ctx, &chronosv1.GetSignalRequest{Id: sig.ID.String()})
	if err != nil {
		t.Fatalf("GetSignal: %v", err)
	}
	if resp.Explanation == nil {
		t.Fatal("Explanation missing — detector context must surface over gRPC")
	}
	if resp.Explanation.DetectorVersion != "trend-v2" {
		t.Errorf("detector_version = %q, want trend-v2", resp.Explanation.DetectorVersion)
	}
	if resp.Explanation.ComparablePeers != 12 {
		t.Errorf("comparable_peers = %d, want 12", resp.Explanation.ComparablePeers)
	}
	if resp.Explanation.BaselineWindowDays != 90 {
		t.Errorf("baseline_window_days = %d, want 90", resp.Explanation.BaselineWindowDays)
	}
	if resp.Explanation.ThresholdUsed != 2.5 {
		t.Errorf("threshold_used = %v, want 2.5", resp.Explanation.ThresholdUsed)
	}
	if len(resp.Explanation.FeatureEvolution) != 3 {
		t.Errorf("feature_evolution len = %d, want 3", len(resp.Explanation.FeatureEvolution))
	}
	if resp.Explanation.FeatureEvolution[2].Value != 26.0 {
		t.Errorf("last sample value = %v, want 26.0", resp.Explanation.FeatureEvolution[2].Value)
	}
}

// TestGetSignal_NoExplanationOmitted: a zero-Explanation signal yields
// no Explanation field on the wire, so callers distinguish absence from
// the zero value.
func TestGetSignal_NoExplanationOmitted(t *testing.T) {
	srv, mem := setupServer(t)
	ctx := context.Background()

	scope := uuid.New()
	now := time.Now()
	sig := domain.Signal{
		ID:         uuid.New(),
		ScopeID:    scope,
		Series:     uuid.New(),
		Pattern:    domain.PatternTypeRecurrence,
		DetectedAt: now,
		Window:     domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
		Strength:   0.7, Confidence: 0.7,
	}
	if err := mem.Signals.Save(ctx, sig); err != nil {
		t.Fatalf("Save: %v", err)
	}
	resp, err := srv.GetSignal(ctx, &chronosv1.GetSignalRequest{Id: sig.ID.String()})
	if err != nil {
		t.Fatalf("GetSignal: %v", err)
	}
	if resp.Explanation != nil {
		t.Errorf("expected nil Explanation, got %+v", resp.Explanation)
	}
}

func TestGetSignal_Found(t *testing.T) {
	srv, mem := setupServer(t)
	ctx := context.Background()

	scope := uuid.New()
	sig := domain.Signal{
		ID:         uuid.New(),
		ScopeID:    scope,
		Series:     uuid.New(),
		Pattern:    domain.PatternTypeTrend,
		DetectedAt: time.Now(),
		Window:     domain.TimeWindow{Start: time.Now().Add(-time.Hour), End: time.Now()},
		Strength:   0.9,
		Confidence: 0.85,
		Metrics:    map[string]float64{"slope": 0.12},
	}
	if err := mem.Signals.Save(ctx, sig); err != nil {
		t.Fatalf("Save: %v", err)
	}

	resp, err := srv.GetSignal(ctx, &chronosv1.GetSignalRequest{Id: sig.ID.String()})
	if err != nil {
		t.Fatalf("GetSignal: %v", err)
	}
	if resp.Id != sig.ID.String() {
		t.Errorf("id = %s, want %s", resp.Id, sig.ID.String())
	}
	if resp.Pattern != chronosv1.PatternType_PATTERN_TYPE_TREND {
		t.Errorf("pattern = %v", resp.Pattern)
	}
}

func TestGetSignal_NotFound(t *testing.T) {
	srv, _ := setupServer(t)
	ctx := context.Background()

	_, err := srv.GetSignal(ctx, &chronosv1.GetSignalRequest{Id: uuid.New().String()})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.NotFound {
		t.Fatalf("code = %v, want NotFound", st.Code())
	}
}

func TestGetSignal_InvalidID(t *testing.T) {
	srv, _ := setupServer(t)
	ctx := context.Background()

	_, err := srv.GetSignal(ctx, &chronosv1.GetSignalRequest{Id: "bad-id"})
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", st.Code())
	}
}

func TestIngestBatch_PersistsAll(t *testing.T) {
	srv, mem := setupServer(t)
	ctx := context.Background()
	scope := uuid.New().String()
	resp, err := srv.IngestBatch(ctx, &chronosv1.IngestBatchRequest{
		Observations: []*chronosv1.IngestRequest{
			{EntityId: uuid.New().String(), ScopeId: scope, Features: []float64{1, 2}},
			{EntityId: uuid.New().String(), ScopeId: scope, Features: []float64{3, 4}},
		},
	})
	if err != nil {
		t.Fatalf("IngestBatch: %v", err)
	}
	if resp.Accepted != 2 {
		t.Errorf("accepted = %d", resp.Accepted)
	}
	got, _ := mem.EntityStates.ListByScope(ctx, uuid.MustParse(scope))
	if len(got) != 2 {
		t.Fatalf("persisted %d, want 2", len(got))
	}
}

func TestValidateConfig_ReportsDetectors(t *testing.T) {
	srv, _ := setupServer(t)
	resp, err := srv.ValidateConfig(context.Background(), &chronosv1.ValidateConfigRequest{
		Env: map[string]string{"CHRONOS_TREND_MIN_POINTS": "1"},
	})
	if err != nil {
		t.Fatalf("ValidateConfig: %v", err)
	}
	if len(resp.Detectors) == 0 {
		t.Fatal("expected detector reports")
	}
}

func TestStreamSignals_UnimplementedWithoutStreamer(t *testing.T) {
	srv, _ := setupServer(t)
	err := srv.StreamSignals(&chronosv1.StreamSignalsRequest{ScopeId: uuid.New().String()}, nil)
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unimplemented {
		t.Fatalf("code = %v, want Unimplemented", err)
	}
}

func TestExportFederation_DisabledByDefault(t *testing.T) {
	t.Setenv("CHRONOS_FEDERATION_ENABLED", "")
	srv, _ := setupServer(t)
	_, err := srv.ExportFederation(context.Background(), &chronosv1.ExportFederationRequest{})
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unimplemented {
		t.Fatalf("code = %v, want Unimplemented", err)
	}
}

func TestListSignals_SinceCursor(t *testing.T) {
	srv, mem := setupServer(t)
	ctx := context.Background()
	scope := uuid.New()
	t0 := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	old := domain.Signal{
		ID: uuid.New(), ScopeID: scope, Series: uuid.New(),
		Pattern: domain.PatternTypeStall, DetectedAt: t0,
		Window: domain.TimeWindow{Start: t0, End: t0}, Strength: 0.5, Confidence: 0.5,
	}
	newer := domain.Signal{
		ID: uuid.New(), ScopeID: scope, Series: uuid.New(),
		Pattern: domain.PatternTypeStall, DetectedAt: t0.Add(time.Hour),
		Window: domain.TimeWindow{Start: t0, End: t0}, Strength: 0.9, Confidence: 0.9,
	}
	_ = mem.Signals.Save(ctx, old)
	_ = mem.Signals.Save(ctx, newer)

	first, err := srv.ListSignals(ctx, &chronosv1.ListSignalsRequest{ScopeId: scope.String()})
	if err != nil {
		t.Fatalf("ListSignals: %v", err)
	}
	if first.NextCursor == "" {
		t.Fatal("expected next_cursor")
	}
	page, err := srv.ListSignals(ctx, &chronosv1.ListSignalsRequest{
		ScopeId:     scope.String(),
		SinceCursor: first.NextCursor,
	})
	if err != nil {
		t.Fatalf("ListSignals cursor: %v", err)
	}
	if page.Count != 0 {
		t.Errorf("page after newest cursor count = %d, want 0", page.Count)
	}
}
