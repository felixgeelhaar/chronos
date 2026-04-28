// Package postgres provides a PostgreSQL-backed implementation of the
// persistence ports. Postgres is the recommended backend for multi-
// process deployments; SQLite covers single-binary, embedded, and test
// use cases.
//
// Queries are hand-written rather than generated; Postgres is a
// secondary backend, and the duplication cost of dual-engine sqlc is
// not yet worth it. If we add complex query patterns we will revisit.
package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	// pq is the PostgreSQL driver registered with database/sql; the
	// blank import is the canonical way to enable "postgres" connection
	// strings without exposing pq's package-level API.
	_ "github.com/lib/pq"
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

// Open connects to a PostgreSQL database and applies migrations.
func Open(connStr string) (*Conn, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("postgres: open: %w", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(connTimeout)

	if err := ensureSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: migrate: %w", err)
	}

	c := &Conn{DB: db}
	c.EntityStates = &EntityStateRepository{conn: c}
	c.Signals = &SignalRepository{conn: c}
	return c, nil
}

// Close releases the underlying database handle.
func (c *Conn) Close() error { return c.DB.Close() }

func ensureSchema(db *sql.DB) error {
	if _, err := db.Exec(migrationInitial); err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
}
