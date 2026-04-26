package store

import (
	"github.com/felixgeelhaar/chronos/internal/store/memory"
	"github.com/felixgeelhaar/chronos/internal/store/postgres"
	"github.com/felixgeelhaar/chronos/internal/store/sqlite"
)

func newSQLiteStore(connStr string) (Store, error) {
	return sqlite.NewStore(connStr)
}

func newPostgresStore(connStr string) (Store, error) {
	return postgres.New(connStr)
}

func newMemoryStore() Store {
	return memory.New()
}
