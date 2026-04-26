// Package config provides Chronos configuration.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all Chronos configuration.
type Config struct {
	// Database
	DBType     string // sqlite, postgres, memory
	DBConnStr  string // connection string or path

	// Computation
	SimilarityThreshold float64       // Minimum cosine similarity (0.0-1.0)
	MinSampleSize       int           // Minimum similar cases to generate insight
	MaxInsightsPerRun   int           // Limit insights per computation
	ComputationTimeout  time.Duration // Overall job timeout

	// HTTP API
	HTTPPort int
	HTTPHost string
}

// Default returns sensible defaults.
func Default() *Config {
	return &Config{
		DBType:              defaultEnv("CHRONOS_DB_TYPE", "sqlite"),
		DBConnStr:           defaultEnv("CHRONOS_DB_CONN", "chronos.db"),
		SimilarityThreshold: defaultEnvFloat64("CHRONOS_SIM_THRESHOLD", 0.85),
		MinSampleSize:       defaultEnvInt("CHRONOS_MIN_SAMPLE", 2),
		MaxInsightsPerRun:   defaultEnvInt("CHRONOS_MAX_INSIGHTS", 10),
		ComputationTimeout:  defaultEnvDuration("CHRONOS_JOB_TIMEOUT", 10*time.Minute),
		HTTPPort:            defaultEnvInt("CHRONOS_HTTP_PORT", 7778),
		HTTPHost:            defaultEnv("CHRONOS_HTTP_HOST", "127.0.0.1"),
	}
}

// Validate checks configuration sanity.
func (c *Config) Validate() error {
	if c.DBType == "" {
		return fmt.Errorf("db_type is required")
	}
	if c.DBConnStr == "" {
		return fmt.Errorf("db_conn is required")
	}
	if c.SimilarityThreshold < 0 || c.SimilarityThreshold > 1 {
		return fmt.Errorf("similarity threshold must be between 0 and 1, got %f", c.SimilarityThreshold)
	}
	if c.MinSampleSize < 1 {
		return fmt.Errorf("min sample size must be at least 1, got %d", c.MinSampleSize)
	}
	return nil
}

func defaultEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func defaultEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func defaultEnvFloat64(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func defaultEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
