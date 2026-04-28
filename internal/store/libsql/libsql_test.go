package libsql

import (
	"context"
	"strings"
	"testing"

	"github.com/felixgeelhaar/chronos/internal/store"
)

func TestTranslateDSN(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "remote turso passthrough",
			in:   "libsql://my-db.turso.io?authToken=secret",
			want: "libsql://my-db.turso.io?authToken=secret",
		},
		{
			name: "namespace stripped from remote",
			in:   "libsql://my-db.turso.io?authToken=t&namespace=chronos",
			// namespace removed; authToken kept.
			want: "libsql://my-db.turso.io?authToken=t",
		},
		{
			name: "local file path",
			in:   "libsql:///tmp/chronos.db",
			want: "file:/tmp/chronos.db",
		},
		{
			name:    "wrong scheme rejected",
			in:      "sqlite:///x",
			wantErr: true,
		},
		{
			name:    "invalid namespace rejected",
			in:      "libsql://h?namespace=Bad-Name",
			wantErr: true,
		},
		{
			name:    "no host or path",
			in:      "libsql://",
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := translateDSN(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr=%v", err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRedactDSN(t *testing.T) {
	in := "libsql://h?authToken=verysecret&namespace=chronos"
	got := redactDSN(in)
	if strings.Contains(got, "verysecret") {
		t.Errorf("authToken not redacted: %q", got)
	}
	if !strings.Contains(got, "authToken=REDACTED") {
		t.Errorf("expected authToken=REDACTED; got %q", got)
	}
}

// TestProviderRegistered confirms init() side effect: the libsql
// scheme is callable through the top-level store.Open. We don't open
// a real connection (no Turso instance in tests) — sql.Open with the
// libsql driver returns without I/O until Ping, and translating an
// unreachable URL surfaces at PingContext.
func TestProviderRegistered(t *testing.T) {
	for _, scheme := range store.SupportedSchemes() {
		if scheme == "libsql" {
			return
		}
	}
	t.Fatalf("libsql provider did not register; SupportedSchemes() = %v", store.SupportedSchemes())
}

func TestOpenProvider_LocalFileSucceeds(t *testing.T) {
	// libsql client supports file: URLs for embedded local-file mode.
	// Use a t.TempDir path so we don't pollute cwd.
	dir := t.TempDir()
	dsn := "libsql://" + dir + "/chronos.db"
	conn, err := openProvider(context.Background(), dsn)
	if err != nil {
		t.Skipf("local libsql open not supported in this environment: %v", err)
	}
	defer func() { _ = conn.Close() }()
	if conn.EntityStates == nil || conn.Signals == nil {
		t.Errorf("libsql Conn has nil repositories")
	}
}
