package store_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/felixgeelhaar/chronos/internal/store"

	// Pull in the providers we want to exercise. Memory and sqlite
	// are sufficient to cover registry dispatch + namespace handling
	// without needing a live Postgres or MySQL.
	_ "github.com/felixgeelhaar/chronos/internal/store/memory"
	_ "github.com/felixgeelhaar/chronos/internal/store/sqlite"
)

func TestOpen_RoutesByScheme(t *testing.T) {
	cases := []struct {
		name string
		dsn  string
	}{
		{"memory", "memory://?namespace=chronos"},
		{"sqlite in-memory", "sqlite://:memory:"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := store.Open(context.Background(), tc.dsn)
			if err != nil {
				t.Fatalf("Open(%q): %v", tc.dsn, err)
			}
			defer func() { _ = conn.Close() }()
			if conn.EntityStates == nil || conn.Signals == nil {
				t.Errorf("Conn has nil repositories: %+v", conn)
			}
		})
	}
}

func TestOpen_RejectsMissingScheme(t *testing.T) {
	_, err := store.Open(context.Background(), "no-scheme-here")
	if err == nil {
		t.Fatal("expected error on schemeless dsn")
	}
	if !strings.Contains(err.Error(), "missing scheme") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestOpen_RejectsUnknownScheme(t *testing.T) {
	_, err := store.Open(context.Background(), "neverheard://anywhere")
	if err == nil {
		t.Fatal("expected error on unknown scheme")
	}
	if !strings.Contains(err.Error(), "unknown provider") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestSupportedSchemes_IncludesRegisteredProviders(t *testing.T) {
	got := store.SupportedSchemes()
	for _, want := range []string{"memory", "sqlite"} {
		found := false
		for _, s := range got {
			if s == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SupportedSchemes() missing %q; got %v", want, got)
		}
	}
}

func TestParseNamespace(t *testing.T) {
	cases := []struct {
		name    string
		query   string
		want    string
		wantErr bool
	}{
		{"empty falls back to default", "", store.DefaultNamespace, false},
		{"explicit valid", "namespace=tenant_a", "tenant_a", false},
		{"upper-case rejected", "namespace=Tenant", "", true},
		{"leading digit rejected", "namespace=1tenant", "", true},
		{"hyphen rejected", "namespace=tenant-a", "", true},
		{"too long rejected", "namespace=" + strings.Repeat("a", 64), "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := &url.URL{RawQuery: tc.query}
			got, err := store.ParseNamespace(u)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseNamespace err = %v, wantErr=%v", err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLegacyDSN(t *testing.T) {
	cases := []struct {
		name    string
		dbType  string
		conn    string
		want    string
		wantErr bool
	}{
		{"empty type returns empty", "", "anything", "", false},
		{"memory", "memory", "ignored", "memory://", false},
		{"sqlite path", "sqlite", "chronos.db", "sqlite://chronos.db", false},
		{"sqlite alias", "sqlite3", "chronos.db", "sqlite://chronos.db", false},
		{"sqlite in-memory", "sqlite", ":memory:", "sqlite://:memory:", false},
		{"postgres url passthrough", "postgres", "postgres://u:p@h/db", "postgres://u:p@h/db", false},
		{"postgresql url passthrough", "postgresql", "postgresql://u@h/db", "postgresql://u@h/db", false},
		{"libpq kv form rejected", "postgres", "host=localhost user=foo", "", true},
		{"unknown type rejected", "duckdb", "x", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := store.LegacyDSN(tc.dbType, tc.conn)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr=%v", err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRegister_PanicsOnDuplicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	store.Register("memory", func(context.Context, string) (*store.Conn, error) { return nil, nil })
}

func TestRegister_PanicsOnEmptyScheme(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty scheme")
		}
	}()
	store.Register("", func(context.Context, string) (*store.Conn, error) { return nil, nil })
}

func TestRegister_PanicsOnNilFunc(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil OpenFunc")
		}
	}()
	store.Register("nilfn", nil)
}
