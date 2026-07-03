// Package libsql registers Chronos's libSQL storage backend so an external
// consumer can build a durable engine with
// embed.New(embed.WithStorage("libsql://…")).
//
// Blank-import it, like a database/sql driver:
//
//	import _ "github.com/felixgeelhaar/chronos/storage/libsql"
//
// The provider lives under internal/store and is not importable directly
// from outside the module; this in-module shim re-registers it.
package libsql

import (
	// Blank import runs the provider's init(), registering it with the store
	// factory (see the package doc). It is the whole purpose of this shim.
	_ "github.com/felixgeelhaar/chronos/internal/store/libsql"
)
