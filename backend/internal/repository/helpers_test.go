package repository_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Configure GORM for SQLite compatibility
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations with manual schema creation for SQLite compatibility
	// SQLite doesn't support PostgreSQL's uuid type, gen_random_uuid(), or certain constraints
	sqlDB, err := db.DB()
	require.NoError(t, err)

	// Enable foreign key constraints for SQLite
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Create users table
	_, err = sqlDB.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			body_weight REAL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`)
	require.NoError(t, err)

	// Create sessions table
	_, err = sqlDB.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			date DATETIME NOT NULL,
			notes TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	require.NoError(t, err)

	// Create sets table
	_, err = sqlDB.Exec(`
		CREATE TABLE sets (
			id TEXT PRIMARY KEY NOT NULL,
			session_id TEXT NOT NULL,
			exercise_name TEXT NOT NULL,
			set_order INTEGER NOT NULL DEFAULT 1,
			weight REAL NOT NULL,
			reps INTEGER NOT NULL,
			rpe REAL,
			notes TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)
	`)
	require.NoError(t, err)

	// Create one_rep_maxes table
	_, err = sqlDB.Exec(`
		CREATE TABLE one_rep_maxes (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			exercise_name TEXT NOT NULL,
			weight REAL NOT NULL,
			date DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	require.NoError(t, err)

	// Create videos table
	_, err = sqlDB.Exec(`
		CREATE TABLE videos (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			session_id TEXT,
			set_id TEXT,
			url TEXT NOT NULL,
			thumbnail_url TEXT,
			duration INTEGER,
			file_size INTEGER,
			exercise_name TEXT,
			date DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE SET NULL,
			FOREIGN KEY (set_id) REFERENCES sets(id) ON DELETE SET NULL
		)
	`)
	require.NoError(t, err)

	return db
}

// teardownTestDB closes the database connection
func teardownTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

// Helper functions for nullable fields
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
