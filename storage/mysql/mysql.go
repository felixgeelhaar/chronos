// Package mysql registers Chronos's MySQL/MariaDB storage backend so an
// external consumer can build a durable engine with
// embed.New(embed.WithStorage("mysql://user:pass@host/db")).
//
// Blank-import it, like a database/sql driver:
//
//	import _ "github.com/felixgeelhaar/chronos/storage/mysql"
//
// The provider lives under internal/store and is not importable directly
// from outside the module; this in-module shim re-registers it.
package mysql

import (
	// Blank import runs the provider's init(), registering it with the store
	// factory (see the package doc). It is the whole purpose of this shim.
	_ "github.com/felixgeelhaar/chronos/internal/store/mysql"
)
