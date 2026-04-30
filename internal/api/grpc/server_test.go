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
