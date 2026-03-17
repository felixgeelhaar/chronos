package config

import (
	"fmt"
	"time"

	"github.com/ascend/api/internal/domain"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL            string
	MaxConnections int
	LogLevel       logger.LogLevel
}

// NewDatabase creates a new database connection
func NewDatabase(config DatabaseConfig) (*gorm.DB, error) {
	// Configure GORM logger
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(config.URL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxConnections)
	sqlDB.SetMaxIdleConns(config.MaxConnections / 2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info().Msg("Database connection established")

	return db, nil
}

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	log.Info().Msg("Running database migrations...")

	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Session{},
		&domain.Set{},
		&domain.OneRepMax{},
		&domain.Video{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create additional indexes
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Info().Msg("Database migrations completed successfully")
	return nil
}

// createIndexes creates additional database indexes for performance
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		// User indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC)",

		// Session indexes
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_id_date ON sessions(user_id, date DESC)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_date ON sessions(date DESC)",

		// Set indexes
		"CREATE INDEX IF NOT EXISTS idx_sets_session_id_order ON sets(session_id, set_order)",
		"CREATE INDEX IF NOT EXISTS idx_sets_exercise_name ON sets(exercise_name)",

		// OneRepMax indexes
		"CREATE INDEX IF NOT EXISTS idx_one_rep_maxes_user_exercise ON one_rep_maxes(user_id, exercise_name, date DESC)",
		"CREATE INDEX IF NOT EXISTS idx_one_rep_maxes_exercise ON one_rep_maxes(exercise_name, date DESC)",

		// Video indexes
		"CREATE INDEX IF NOT EXISTS idx_videos_user_id_date ON videos(user_id, date DESC)",
		"CREATE INDEX IF NOT EXISTS idx_videos_session_id ON videos(session_id)",
		"CREATE INDEX IF NOT EXISTS idx_videos_exercise ON videos(exercise_name, date DESC)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			log.Warn().Err(err).Str("index", index).Msg("Failed to create index (may already exist)")
		}
	}

	log.Info().Msg("Database indexes created")
	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Info().Msg("Database connection closed")
	return nil
}
