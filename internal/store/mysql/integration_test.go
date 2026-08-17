package mysql_test

// Integration tests that exercise a live MySQL/MariaDB server. Mirrors
// the pattern Mnemos uses for the same backend: gate on TEST_MYSQL_DSN
// so developer machines without a server just see SKIP, while CI
// (and operators running locally) can supply a docker-compose- or
// testcontainers-managed instance.
//
// Each test gets a unique namespace (database) so parallel runs don't
// collide and a leftover database from a previous failure doesn't
// poison the next run. Cleanup drops the database.

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/felixgeelhaar/chronos"
	"github.com/felixgeelhaar/chronos/internal/domain"
	"github.com/felixgeelhaar/chronos/internal/ports"
	"github.com/felixgeelhaar/chronos/internal/store"
	_ "github.com/felixgeelhaar/chronos/internal/store/mysql"
	"github.com/google/uuid"
)

func requireLiveDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN not set; skipping mysql integration test")
	}
	return dsn
}

// withConn opens a Conn against a fresh per-test namespace and
// schedules its teardown. Pattern matches Mnemos's mysql package.
func withConn(t *testing.T) (*store.Conn, string) {
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
		// Drop the per-test database so repeated runs stay clean.
		if raw, ok := conn.Raw.(interface {
			ExecContext(context.Context, string, ...any) (any, error)
		}); ok {
			_, _ = raw.ExecContext(context.Background(), fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", ns))
		}
		_ = conn.Close()
	})
	return conn, ns
}

func withNamespaceQuery(dsn, ns string) string {
	if strings.Contains(dsn, "?") {
		return dsn + "&namespace=" + ns
	}
	return dsn + "?namespace=" + ns
}

func TestMySQL_OpenBootstrapsSchema(t *testing.T) {
	conn, _ := withConn(t)
	if conn.EntityStates == nil || conn.Signals == nil {
		t.Fatalf("mysql Conn has nil ports: %+v", conn)
	}
}

func TestMySQL_EntityStateRoundTrip(t *testing.T) {
	conn, _ := withConn(t)
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

func TestMySQL_SignalRoundTripWithEvidence(t *testing.T) {
	conn, _ := withConn(t)
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
	if got.Evidence[0].Metrics["outcome_diff"] != 1.5 {
		t.Errorf("evidence metrics lost: %+v", got.Evidence[0].Metrics)
	}
	if !got.Explanation.IsZero() {
		t.Errorf("zero Explanation should round-trip empty, got %+v", got.Explanation)
	}

	listed, err := conn.Signals.List(ctx, ports.SignalFilter{ScopeID: scope})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(listed) != 1 {
		t.Errorf("List length = %d, want 1", len(listed))
	}

	n, err := conn.Signals.Count(ctx, ports.SignalFilter{ScopeID: scope})
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 1 {
		t.Errorf("Count = %d, want 1", n)
	}
}

func TestMySQL_GetMissingReturnsErrSignalNotFound(t *testing.T) {
	conn, _ := withConn(t)
	if _, err := conn.Signals.Get(context.Background(), uuid.New()); err == nil {
		t.Fatal("expected error when fetching missing signal")
	} else if err.Error() != domain.ErrSignalNotFound.Error() {
		t.Errorf("err = %v, want ErrSignalNotFound", err)
	}
}
