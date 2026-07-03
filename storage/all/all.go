// Package all registers every Chronos SQL storage backend at once
// (Postgres, SQLite, MySQL/MariaDB, libSQL) for consumers that don't want to
// pick individual drivers. Blank-import it:
//
//	import _ "github.com/felixgeelhaar/chronos/storage/all"
//
// This pulls in every backend's driver dependency. Prefer a single
// per-backend import (e.g. storage/postgres) when you only need one, to keep
// your dependency graph and binary lean.
package all

import (
	// Each blank import registers one durable backend via its shim (see the
	// package doc). Importing them all is the whole purpose of this package.
	_ "github.com/felixgeelhaar/chronos/storage/libsql"
	_ "github.com/felixgeelhaar/chronos/storage/mysql"
	_ "github.com/felixgeelhaar/chronos/storage/postgres"
	_ "github.com/felixgeelhaar/chronos/storage/sqlite"
)
