// Package postgres registers Chronos's Postgres storage backend so an
// external consumer can build a durable engine with
// embed.New(embed.WithStorage("postgres://…")).
//
// The provider implementation lives under internal/store and cannot be
// imported from outside the Chronos module (Go's internal-package rule).
// Blank-import this package to register it, exactly like a database/sql
// driver:
//
//	import (
//		"github.com/felixgeelhaar/chronos/embed"
//		_ "github.com/felixgeelhaar/chronos/storage/postgres"
//	)
//
//	eng, err := embed.New(embed.WithStorage("postgres://HOST/DB?namespace=chronos"))
//
// Registering also pulls in the pgx driver, so only consumers that want
// Postgres pay for it — memory-only consumers stay dependency-light.
package postgres

import (
	// Blank import runs the provider's init(), registering it with the store
	// factory (see the package doc). It is the whole purpose of this shim.
	_ "github.com/felixgeelhaar/chronos/internal/store/postgres"
)
