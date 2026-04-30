package client

import (
	"context"
	"net"
	"testing"
	"time"

	chronosv1 "github.com/felixgeelhaar/chronos/api/proto/chronos/v1"
	grpctransport "github.com/felixgeelhaar/chronos/internal/api/grpc"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/store/memory"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpcmetadata "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// setupGRPCServer spins up an in-memory gRPC server and returns its address
// and the underlying memory store so tests can assert side effects.
func setupGRPCServer(t *testing.T) (string, *memory.Conn) {
	t.Helper()
	mem := memory.New()
	srv := grpctransport.NewServer(mem.EntityStates, mem.Signals, nil, nil)
	grpcSrv := grpc.NewServer()
	chronosv1.RegisterChronosServiceServer(grpcSrv, srv)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			t.Logf("grpc serve: %v", err)
		}
	}()
	t.Cleanup(func() { grpcSrv.GracefulStop() })
	return lis.Addr().String(), mem
}

func TestGRPCClient_Ingest(t *testing.T) {
	addr, mem := setupGRPCServer(t)
	c, err := NewGRPC(addr)
	if err != nil {
		t.Fatalf("NewGRPC: %v", err)
	}
	defer c.Close()

	scope := uuid.New()
	entity := uuid.New()
	id, err := c.Ingest(context.Background(), IngestRequest{
		EntityID:  entity,
		ScopeID:   scope,
		Timestamp: time.Now(),
		Features:  []float64{1, 2, 3, 5},
		Adapter:   "test",
	})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if id == uuid.Nil {
		t.Fatal("expected non-nil id")
	}
	got, _ := mem.EntityStates.ListByScope(context.Background(), scope)
	if len(got) != 1 || got[0].EntityID != entity {
		t.Fatalf("state not persisted: %+v", got)
	}
}

func TestGRPCClient_ListSignals(t *testing.T) {
	addr, mem := setupGRPCServer(t)
	c, err := NewGRPC(addr)
	if err != nil {
		t.Fatalf("NewGRPC: %v", err)
	}
	defer c.Close()

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
		if err := mem.Signals.Save(context.Background(), sig); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	signals, err := c.Signals().Scope(scope).MinConfidence(0.7).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(signals) != 2 {
		t.Fatalf("len = %d, want 2", len(signals))
	}
	for _, s := range signals {
		if s.Pattern != string(domain.PatternTypeRecurrence) {
			t.Errorf("pattern = %s", s.Pattern)
		}
		if s.Confidence < 0.7 {
			t.Errorf("confidence %f below threshold", s.Confidence)
		}
	}
}

func TestGRPCClient_GetSignal(t *testing.T) {
	addr, mem := setupGRPCServer(t)
	c, err := NewGRPC(addr)
	if err != nil {
		t.Fatalf("NewGRPC: %v", err)
	}
	defer c.Close()

	sig := domain.Signal{
		ID:         uuid.New(),
		ScopeID:    uuid.New(),
		Series:     uuid.New(),
		Pattern:    domain.PatternTypeTrend,
		DetectedAt: time.Now(),
		Window:     domain.TimeWindow{Start: time.Now().Add(-time.Hour), End: time.Now()},
		Strength:   0.9,
		Confidence: 0.85,
		Metrics:    map[string]float64{"slope": 0.12},
	}
	if err := mem.Signals.Save(context.Background(), sig); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := c.Signals().Get(context.Background(), sig.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != sig.ID {
		t.Errorf("id = %v, want %v", got.ID, sig.ID)
	}
	if got.Pattern != string(domain.PatternTypeTrend) {
		t.Errorf("pattern = %s", got.Pattern)
	}
}

func TestGRPCClient_WithToken(t *testing.T) {
	// Server with auth interceptor.
	mem := memory.New()
	srv := grpctransport.NewServer(mem.EntityStates, mem.Signals, nil, nil)
	grpcSrv := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := grpcmetadata.FromIncomingContext(ctx)
		if !ok || len(md.Get("authorization")) == 0 || md.Get("authorization")[0] != "Bearer secret" {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return handler(ctx, req)
	}))
	chronosv1.RegisterChronosServiceServer(grpcSrv, srv)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go grpcSrv.Serve(lis)
	t.Cleanup(func() { grpcSrv.GracefulStop() })

	// Client without token should fail.
	cNoAuth, err := NewGRPC(lis.Addr().String())
	if err != nil {
		t.Fatalf("NewGRPC: %v", err)
	}
	defer cNoAuth.Close()
	_, err = cNoAuth.Ingest(context.Background(), IngestRequest{
		EntityID: uuid.New(),
		ScopeID:  uuid.New(),
		Features: []float64{1, 2},
	})
	if err == nil {
		t.Fatal("expected unauthenticated error without token")
	}

	// Client with token should succeed.
	cAuth, err := NewGRPC(lis.Addr().String(), WithGRPCToken("secret"))
	if err != nil {
		t.Fatalf("NewGRPC: %v", err)
	}
	defer cAuth.Close()
	_, err = cAuth.Ingest(context.Background(), IngestRequest{
		EntityID: uuid.New(),
		ScopeID:  uuid.New(),
		Features: []float64{1, 2},
	})
	if err != nil {
		t.Fatalf("Ingest with token: %v", err)
	}
}

// Ensure the grpc package is imported so the compiler checks the generated
// client code against the server implementation.
var _ = insecure.NewCredentials
