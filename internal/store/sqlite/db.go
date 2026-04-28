// Package sqlite provides a SQLite-backed implementation of the
// persistence ports. The driver is modernc.org/sqlite (pure Go, no CGO)
// so chronos builds and cross-compiles without a C toolchain.
//
// The package owns one [Conn] type that bundles the per-aggregate
// repositories (entity states, signals) on a shared *sql.DB. SQL goes
// through sqlc-generated code in the sqlcgen subpackage; this file is
// only responsible for opening the database, configuring PRAGMAs, and
// applying migrations.
package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/felixgeelhaar/chronos/internal/store/sqlite/sqlcgen"

	// modernc.org/sqlite is a pure-Go SQLite driver; the blank import
	// registers it with database/sql under the name "sqlite". Using a
	// pure-Go driver lets chronos cross-compile without a C toolchain.
	_ "modernc.org/sqlite"
)

//go:embed migrations/001_initial.sql
var migrationInitial string

// Conn bundles the SQLite-backed repositories. Construct with [Open].
type Conn struct {
	DB *sql.DB
	q  *sqlcgen.Queries

	EntityStates *EntityStateRepository
	Signals      *SignalRepository
}

// Open connects to a SQLite database at path (or ":memory:") and
// applies migrations. Foreign keys, WAL journal, and a 5-second busy
// timeout are enabled via PRAGMAs encoded in the DSN.
func Open(path string) (*Conn, error) {
	dsn := buildDSN(path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}

	// SQLite serialises writes; one connection avoids "database is
	// locked" surprises in tests and small workloads. Production
	// deployments that need higher throughput should use Postgres.
	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(0)

	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: ping: %w", err)
	}
	if err := ensureSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: migrate: %w", err)
	}

	q := sqlcgen.New(db)
	c := &Conn{DB: db, q: q}
	c.EntityStates = &EntityStateRepository{conn: c}
	c.Signals = &SignalRepository{conn: c}
	return c, nil
}

// Close releases the underlying database handle.
func (c *Conn) Close() error { return c.DB.Close() }

// buildDSN encodes the engine PRAGMAs as modernc.org/sqlite query
// parameters.
func buildDSN(path string) string {
	q := url.Values{}
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "journal_mode(wal)")
	q.Add("_pragma", "busy_timeout(5000)")
	if path == ":memory:" {
		return "file::memory:?" + q.Encode()
	}
	return "file:" + path + "?" + q.Encode()
}

// ensureSchema applies the embedded migration. The migration uses
// CREATE TABLE / CREATE INDEX in a fresh schema, so we run statements
// one-at-a-time to satisfy database/sql's prepared-statement model.
func ensureSchema(db *sql.DB) error {
	for _, stmt := range splitStatements(migrationInitial) {
		s := stripLeadingCommentsAndBlanks(stmt)
		if s == "" {
			continue
		}
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("apply %q: %w", firstLine(s), err)
		}
	}
	return nil
}

func splitStatements(sqlText string) []string {
	var out []string
	var current strings.Builder
	for _, line := range strings.Split(sqlText, "\n") {
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			out = append(out, current.String())
			current.Reset()
		}
	}
	if rest := strings.TrimSpace(current.String()); rest != "" {
		out = append(out, rest)
	}
	return out
}

// stripLeadingCommentsAndBlanks removes leading `--`-style comment
// lines and blank lines from a candidate statement so the result either
// is a runnable SQL statement or an empty string we skip.
func stripLeadingCommentsAndBlanks(stmt string) string {
	lines := strings.Split(stmt, "\n")
	i := 0
	for ; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if t == "" || strings.HasPrefix(t, "--") {
			continue
		}
		break
	}
	if i >= len(lines) {
		return ""
	}
	return strings.TrimSpace(strings.Join(lines[i:], "\n"))
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx > 0 {
		return s[:idx]
	}
	return s
}

func formatTime(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}
