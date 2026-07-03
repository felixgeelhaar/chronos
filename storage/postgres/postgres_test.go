package postgres_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/felixgeelhaar/chronos/embed"

	// The unit under test: blank-importing this shim must register the
	// Postgres provider so embed.New can build a durable engine. Without it,
	// embed.New("postgres://…") fails with an unknown-scheme error because the
	// provider lives under internal/ and is unreachable from outside the module.
	_ "github.com/felixgeelhaar/chronos/storage/postgres"
)

// TestPostgresShim_RegistersDurableProvider proves the public shim makes the
// internal Postgres backend reachable to external consumers via embed.New.
// Gated on TEST_POSTGRES_DSN so the default `go test ./...` stays hermetic.
func TestPostgresShim_RegistersDurableProvider(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN not set; skipping durable-chronos shim test")
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	eng, err := embed.New(embed.WithStorage(dsn + sep + "namespace=chronos_shim_test"))
	if err != nil {
		t.Fatalf("embed.New(postgres) failed — shim did not register the provider: %v", err)
	}
	t.Cleanup(func() { _ = eng.Close() })
	// A trivial round-trip proves the engine is live, not just constructed.
	if _, err := eng.Query(context.Background(), embed.QueryOpts{}); err != nil {
		t.Fatalf("engine.Query on durable postgres: %v", err)
	}
}
