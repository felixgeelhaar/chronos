// Package postgres provides a PostgreSQL-backed implementation of the
// persistence ports. Postgres is the recommended backend for multi-
// process deployments; SQLite covers single-binary, embedded, and test
// use cases.
//
// Wire-protocol compatibles supported by this provider via the same
// driver: CockroachDB, Yugabyte, Neon, Crunchy Bridge, TimescaleDB,
// AlloyDB Omni. They all speak the libpq wire protocol; Chronos
// doesn't use any postgres-specific extension.
//
// Per Mnemos ADR 0001 §3, ?namespace= translates to a Postgres
// schema: CREATE SCHEMA IF NOT EXISTS <ns> + SET search_path TO <ns>
// runs at every Open. The schema-version table lives inside the
// namespace, so two tools (Mnemos and Chronos) sharing one Postgres
// database track their migrations independently.
package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/felixgeelhaar/chronos/internal/store"

	// pgx-stdlib registers the "pgx" sql.DB driver in init(). Replaces
	// lib/pq so the cognitive stack (Mnemos, Chronos, Praxis, Nous)
	// shares one driver surface.
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/001_initial.sql
var migrationInitial string

// connTimeout caps the lifetime of pooled connections.
const connTimeout = 5 * time.Minute

// Conn bundles the Postgres-backed repositories.
type Conn struct {
	DB *sql.DB

	EntityStates *EntityStateRepository
	Signals      *SignalRepository
}

func init() {
	store.Register("postgres", openProvider)
	store.Register("postgresql", openProvider)
}

// openProvider is the store.OpenFunc that backs postgres:// /
// postgresql:// DSNs. It strips the ?namespace= query parameter
// before handing the DSN to pgx (pgx rejects unknown query keys),
// then ensures the schema exists and sets search_path so all
// subsequent queries land in the namespace.
func openProvider(ctx context.Context, dsn string) (*store.Conn, error) {
	parsed, err := parseDSN(dsn)
	if err != nil {
		return nil, err
	}
	c, err := openWithNamespace(ctx, parsed)
	if err != nil {
		return nil, err
	}
	return &store.Conn{
		EntityStates: c.EntityStates,
		Signals:      c.Signals,
		Raw:          c.DB,
		Closer:       c.Close,
	}, nil
}

// dsnParts holds the relevant pieces of a parsed Chronos postgres DSN.
type dsnParts struct {
	// Driver is the libpq-shaped DSN with ?namespace= stripped, ready
	// to hand to pgx.
	Driver string
	// Namespace is the validated schema name (default "chronos").
	Namespace string
}

func parseDSN(dsn string) (dsnParts, error) {
	if !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
		return dsnParts{}, fmt.Errorf("postgres: not a postgres dsn: %q", dsn)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return dsnParts{}, fmt.Errorf("postgres: parse dsn: %w", err)
	}
	ns, err := store.ParseNamespace(u)
	if err != nil {
		return dsnParts{}, err
	}
	q := u.Query()
	q.Del("namespace")
	u.RawQuery = q.Encode()
	return dsnParts{Driver: u.String(), Namespace: ns}, nil
}

// Open is kept exported for tests and back-compat callers; the
// preferred entry point is store.Open(ctx, "postgres://...").
func Open(connStr string) (*Conn, error) {
	parsed, err := parseDSN(connStr)
	if err != nil {
		// Allow back-compat callers to pass a libpq URL without
		// namespace; in that case parseDSN returns an error only on
		// truly malformed input.
		return nil, err
	}
	return openWithNamespace(context.Background(), parsed)
}

func openWithNamespace(ctx context.Context, p dsnParts) (*Conn, error) {
	db, err := sql.Open("pgx", p.Driver)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(connTimeout)

	if err := applyNamespace(ctx, db, p.Namespace); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := ensureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: migrate: %w", err)
	}

	c := &Conn{DB: db}
	c.EntityStates = &EntityStateRepository{conn: c}
	c.Signals = &SignalRepository{conn: c}
	return c, nil
}

// applyNamespace runs CREATE SCHEMA IF NOT EXISTS + SET search_path.
// The schema name has already been validated against namespaceRE so
// fmt-substitution is safe — Postgres identifiers are not parameter-
// substitutable in CREATE SCHEMA / SET, so we cannot use a placeholder.
func applyNamespace(ctx context.Context, db *sql.DB, ns string) error {
	if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", ns)); err != nil {
		return fmt.Errorf("postgres: create schema %q: %w", ns, err)
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", ns)); err != nil {
		return fmt.Errorf("postgres: set search_path %q: %w", ns, err)
	}
	return nil
}

// Close releases the underlying database handle.
func (c *Conn) Close() error { return c.DB.Close() }

func ensureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, migrationInitial); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
}
