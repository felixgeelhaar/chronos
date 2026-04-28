package postgres_test

// Integration tests that exercise a live Postgres server. Mirrors the
// MySQL pattern: gate on TEST_POSTGRES_DSN so developer machines
// without a server just see SKIP, while CI provisions a service
// container.
//
// Each test uses a unique per-run namespace so parallel runs and
// leftover state from previous failures don't poison results. Cleanup
// drops the namespace's tables.

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/felixgeelhaar/chronos/internal/store"
	_ "github.com/felixgeelhaar/chronos/internal/store/postgres"
	"github.com/google/uuid"
)

func requireLiveDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set; skipping postgres integration test")
	}
	return dsn
}

// withConn opens a Conn against a fresh per-test schema namespace and
// schedules its teardown. Cleanup is best-effort: we attempt
// `DROP SCHEMA <ns> CASCADE` and ignore errors so a failing test
// doesn't mask the original assertion.
func withConn(t *testing.T) *store.Conn {
	t.Helper()
	base := requireLiveDSN(t)
	ns := fmt.Sprintf("chronos_test_%d", time.Now().UnixNano())
	full := withNamespaceQuery(base, ns)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := store.Open(ctx, full)
	if err != nil {
		t.Fatalf("store.Open(%s): %v", full, err)
	}
	t.Cleanup(func() {
		if raw, ok := conn.Raw.(*sql.DB); ok {
			// Identifier interpolation is safe — the namespace was
			// validated by store.ParseNamespace at Open time.
			_, _ = raw.ExecContext(context.Background(), fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", ns))
		}
		_ = conn.Close()
	})
	return conn
}

func withNamespaceQuery(dsn, ns string) string {
	if strings.Contains(dsn, "?") {
		return dsn + "&namespace=" + ns
	}
	return dsn + "?namespace=" + ns
}

func TestPostgres_OpenBootstrapsSchema(t *testing.T) {
	conn := withConn(t)
	if conn.EntityStates == nil || conn.Signals == nil {
		t.Fatalf("postgres Conn has nil ports: %+v", conn)
	}
}

func TestPostgres_EntityStateRoundTrip(t *testing.T) {
	conn := withConn(t)
	ctx := context.Background()
	scope := uuid.New()
	a := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	if err := conn.EntityStates.Ingest(ctx, "test", chronos.EntityState{
		ID: uuid.New(), EntityID: a, ScopeID: scope, Timestamp: now,
		Features: []float64{1, 2, 3, 5},
	}); err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	got, err := conn.EntityStates.ListByScope(ctx, scope)
	if err != nil {
		t.Fatalf("ListByScope: %v", err)
	}
	if len(got) != 1 || got[0].EntityID != a {
		t.Fatalf("scope read returned wrong rows: %+v", got)
	}
	scopes, err := conn.EntityStates.ListScopes(ctx)
	if err != nil {
		t.Fatalf("ListScopes: %v", err)
	}
	if len(scopes) != 1 || scopes[0] != scope {
		t.Errorf("ListScopes = %v, want [%s]", scopes, scope)
	}
}

func TestPostgres_SignalRoundTripWithEvidence(t *testing.T) {
	conn := withConn(t)
	ctx := context.Background()
	scope := uuid.New()
	series := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)

	sig := domain.Signal{
		ID:         uuid.New(),
		ScopeID:    scope,
		Series:     series,
		Pattern:    domain.PatternTypeRecurrence,
		DetectedAt: now,
		Window:     domain.TimeWindow{Start: now.Add(-time.Hour), End: now},
		Strength:   0.85,
		Confidence: 0.85,
		Metrics:    map[string]float64{"avg_similarity": 0.85},
		Evidence: []domain.Evidence{{
			Series: uuid.New(), Time: now.Add(-30 * time.Minute),
			Kind: "similar_state", Score: 0.92,
			Metrics: map[string]float64{"outcome_diff": 1.5},
		}},
	}
	if err := conn.Signals.Save(ctx, sig); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := conn.Signals.Get(ctx, sig.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Pattern != domain.PatternTypeRecurrence {
		t.Errorf("Pattern lost: %s", got.Pattern)
	}
	if len(got.Evidence) != 1 || got.Evidence[0].Kind != "similar_state" {
		t.Fatalf("evidence not round-tripped: %+v", got.Evidence)
	}

	listed, err := conn.Signals.List(ctx, ports.SignalFilter{ScopeID: scope})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(listed) != 1 {
		t.Errorf("List length = %d, want 1", len(listed))
	}
}

func TestPostgres_GetMissingReturnsErrSignalNotFound(t *testing.T) {
	conn := withConn(t)
	if _, err := conn.Signals.Get(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected error when fetching missing signal")
	} else if err.Error() != domain.ErrSignalNotFound.Error() {
		t.Errorf("err = %v, want ErrSignalNotFound", err)
	}
}
