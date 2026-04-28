// Package mysql implements a [store] provider backed by MySQL or
// MariaDB. The wire protocol is shared and the SQL dialect we rely on
// is the common subset, so one provider serves both. Wire-protocol
// compatibles known to work with this driver: MariaDB, PlanetScale,
// TiDB, Vitess.
//
// Per Mnemos ADR 0001 §3, MySQL has no per-tenant schemas — the
// namespace translates to a *database*. The provider runs
// CREATE DATABASE IF NOT EXISTS <namespace> against the server, then
// reconnects with that database selected. Schema bootstrap
// (schema.sql) is applied on every Open and is idempotent.
//
// Both the canonical Chronos URL form (`mysql://...`) and the MariaDB
// alias (`mariadb://...`) are accepted. The Go driver consumes a
// libmysql-style DSN (`user:pw@tcp(host:port)/db?...`), not a URL, so
// this package translates between the two so the rest of the stack
// can stay URL-uniform.
package mysql

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/felixgeelhaar/chronos/internal/store"

	// go-sql-driver/mysql registers the "mysql" sql.DB driver in init().
	_ "github.com/go-sql-driver/mysql"
)

//go:embed schema.sql
var schemaSQL string

// connTimeout caps the lifetime of pooled connections.
const connTimeout = 5 * time.Minute

// Conn bundles the MySQL-backed repositories.
type Conn struct {
	DB *sql.DB

	EntityStates *EntityStateRepository
	Signals      *SignalRepository
}

func init() {
	store.Register("mysql", openProvider)
	store.Register("mariadb", openProvider)
}

// dsnParts holds the relevant pieces of a parsed Chronos MySQL DSN.
type dsnParts struct {
	// AdminDSN is the same as DriverDSN but without a database name.
	// Used for the CREATE DATABASE step before we reconnect with the
	// namespace selected.
	AdminDSN string
	// DriverDSN is the libmysql-style DSN with the namespace selected
	// as the default database. Includes parseTime=true&loc=UTC so
	// DATETIME columns scan into time.Time as UTC.
	DriverDSN string
	// Namespace is the validated database name.
	Namespace string
}

// openProvider is the store.OpenFunc that backs mysql:// / mariadb://
// DSNs.
func openProvider(ctx context.Context, dsn string) (*store.Conn, error) {
	parsed, err := parseDSN(dsn)
	if err != nil {
		return nil, err
	}

	// Step 1: connect without a database to issue CREATE DATABASE.
	admin, err := sql.Open("mysql", parsed.AdminDSN)
	if err != nil {
		return nil, fmt.Errorf("mysql: open admin: %w", err)
	}
	if err := admin.PingContext(ctx); err != nil {
		_ = admin.Close()
		return nil, fmt.Errorf("mysql: ping admin: %w", err)
	}
	// MySQL identifiers in DDL are not parameter-substitutable; the
	// namespace has been validated against namespaceRE so safe to
	// interpolate.
	if _, err := admin.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", parsed.Namespace)); err != nil {
		_ = admin.Close()
		return nil, fmt.Errorf("mysql: create database %q: %w", parsed.Namespace, err)
	}
	_ = admin.Close()

	// Step 2: reconnect with the namespace selected.
	db, err := sql.Open("mysql", parsed.DriverDSN)
	if err != nil {
		return nil, fmt.Errorf("mysql: open: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: ping: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(connTimeout)

	if err := ensureSchema(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql: migrate: %w", err)
	}

	c := &Conn{DB: db}
	c.EntityStates = &EntityStateRepository{conn: c}
	c.Signals = &SignalRepository{conn: c}
	return &store.Conn{
		EntityStates: c.EntityStates,
		Signals:      c.Signals,
		Raw:          db,
		Closer:       c.Close,
	}, nil
}

// Close releases the underlying database handle.
func (c *Conn) Close() error { return c.DB.Close() }

// parseDSN translates a `mysql://` or `mariadb://` URL into the
// libmysql DSN the Go driver requires, plus an admin DSN for the
// CREATE DATABASE step.
func parseDSN(dsn string) (dsnParts, error) {
	if !strings.HasPrefix(dsn, "mysql://") && !strings.HasPrefix(dsn, "mariadb://") {
		return dsnParts{}, fmt.Errorf("mysql: not a mysql/mariadb dsn: %q", dsn)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return dsnParts{}, fmt.Errorf("mysql: parse dsn: %w", err)
	}

	ns, err := store.ParseNamespace(u)
	if err != nil {
		return dsnParts{}, err
	}

	// User info.
	user := u.User.Username()
	pass, hasPass := u.User.Password()

	// Host / port. Default to 3306 if unspecified.
	host := u.Host
	if host == "" {
		host = "localhost:3306"
	} else if !strings.Contains(host, ":") {
		host += ":3306"
	}

	// Build the libmysql query string. Drop ?namespace= and add the
	// driver knobs Chronos always wants.
	q := u.Query()
	q.Del("namespace")
	q.Set("parseTime", "true")
	q.Set("loc", "UTC")
	driverQuery := q.Encode()

	// Build user:pass@tcp(host)/db?... form.
	auth := user
	if hasPass {
		auth = user + ":" + pass
	}
	if user == "" {
		auth = "root"
	}

	driverDSN := fmt.Sprintf("%s@tcp(%s)/%s?%s", auth, host, ns, driverQuery)
	adminDSN := fmt.Sprintf("%s@tcp(%s)/?%s", auth, host, driverQuery)

	return dsnParts{
		AdminDSN:  adminDSN,
		DriverDSN: driverDSN,
		Namespace: ns,
	}, nil
}

// ensureSchema applies schema.sql. The MySQL driver runs only one
// statement per Exec call by default, so split on `;`. Empty fragments
// (trailing whitespace) are skipped.
func ensureSchema(ctx context.Context, db *sql.DB) error {
	for _, stmt := range splitStatements(schemaSQL) {
		s := strings.TrimSpace(stmt)
		if s == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, s); err != nil {
			return fmt.Errorf("apply schema: %w", err)
		}
	}
	return nil
}

func splitStatements(sqlText string) []string {
	var out []string
	var current strings.Builder
	for _, line := range strings.Split(sqlText, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") || trimmed == "" {
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
		if strings.HasSuffix(trimmed, ";") {
			out = append(out, current.String())
			current.Reset()
		}
	}
	if rest := strings.TrimSpace(current.String()); rest != "" {
		out = append(out, rest)
	}
	return out
}
