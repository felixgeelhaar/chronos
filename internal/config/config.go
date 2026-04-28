// Package config provides Chronos configuration. All knobs are env-var
// driven and read once at startup; the rest of the engine receives a
// *Config explicitly so global mutable state stays out of the codebase.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all Chronos configuration. Tier-A knobs (DB, recurrence,
// HTTP) live in this top-level struct; Tier-B detector knobs are added
// next to the Tier-A fields as more detectors land.
type Config struct {
	// Database
	DBType    string // sqlite, postgres, memory
	DBConnStr string // connection string or path

	// Detection — common
	MaxSignalsPerRun   int           // Limit signals per detect run; 0 = unlimited
	ComputationTimeout time.Duration // Overall compute job timeout

	// Detection — Recurrence
	SimilarityThreshold float64 // Minimum cosine similarity (0..1)
	MinSampleSize       int     // Minimum peer cases to emit a recurrence signal

	// Detection — Trend (Tier B)
	TrendMinSlope  float64 // Minimum |slope| of normalised regression
	TrendMinPoints int     // Minimum points to consider a trend

	// Detection — Spike / Drop (Tier B)
	SpikeZScore float64 // Absolute z-score for a positive deviation
	DropZScore  float64 // Absolute z-score for a negative deviation
	SpikeWindow int     // Window size (in points) for the rolling baseline

	// Detection — Stall (Tier B)
	StallMaxStdDev float64 // Maximum stddev of normalised outcome to qualify as stalled
	StallMinPoints int     // Minimum points to consider a stall

	// Detection — Anomaly (cross-entity, the dual of Recurrence)
	AnomalyMaxSimilarity float64 // Max similarity to nearest peer to count as anomalous
	AnomalyMinPeers      int     // Minimum peers required for cross-entity comparison

	// Detection — Seasonality (autocorrelation)
	SeasonalityMinAutocorr float64 // Minimum autocorrelation at any lag to emit
	SeasonalityMinPoints   int     // Minimum observations to consider seasonality
	SeasonalityMinPeriod   int     // Minimum lag (period) to consider

	// Detection — Correlation (cross-series Pearson)
	CorrelationMin       float64 // Minimum |Pearson r| to emit
	CorrelationMinPoints int     // Minimum aligned observations between two series

	// HTTP API
	HTTPPort  int
	HTTPHost  string
	APIToken  string // optional bearer token; empty disables auth on the API
}

// Default returns sensible defaults.
func Default() *Config {
	return &Config{
		DBType:    defaultEnv("CHRONOS_DB_TYPE", "sqlite"),
		DBConnStr: defaultEnv("CHRONOS_DB_CONN", "chronos.db"),

		MaxSignalsPerRun:   defaultEnvInt("CHRONOS_MAX_SIGNALS", 10),
		ComputationTimeout: defaultEnvDuration("CHRONOS_JOB_TIMEOUT", 10*time.Minute),

		SimilarityThreshold: defaultEnvFloat64("CHRONOS_SIM_THRESHOLD", 0.85),
		MinSampleSize:       defaultEnvInt("CHRONOS_MIN_SAMPLE", 2),

		TrendMinSlope:  defaultEnvFloat64("CHRONOS_TREND_MIN_SLOPE", 0.05),
		TrendMinPoints: defaultEnvInt("CHRONOS_TREND_MIN_POINTS", 4),

		SpikeZScore: defaultEnvFloat64("CHRONOS_SPIKE_Z", 2.5),
		DropZScore:  defaultEnvFloat64("CHRONOS_DROP_Z", 2.5),
		SpikeWindow: defaultEnvInt("CHRONOS_SPIKE_WINDOW", 5),

		StallMaxStdDev: defaultEnvFloat64("CHRONOS_STALL_MAX_STDDEV", 0.05),
		StallMinPoints: defaultEnvInt("CHRONOS_STALL_MIN_POINTS", 4),

		AnomalyMaxSimilarity: defaultEnvFloat64("CHRONOS_ANOMALY_MAX_SIM", 0.5),
		AnomalyMinPeers:      defaultEnvInt("CHRONOS_ANOMALY_MIN_PEERS", 2),

		SeasonalityMinAutocorr: defaultEnvFloat64("CHRONOS_SEASONALITY_MIN_AUTOCORR", 0.5),
		SeasonalityMinPoints:   defaultEnvInt("CHRONOS_SEASONALITY_MIN_POINTS", 12),
		SeasonalityMinPeriod:   defaultEnvInt("CHRONOS_SEASONALITY_MIN_PERIOD", 2),

		CorrelationMin:       defaultEnvFloat64("CHRONOS_CORRELATION_MIN", 0.7),
		CorrelationMinPoints: defaultEnvInt("CHRONOS_CORRELATION_MIN_POINTS", 5),

		HTTPPort: defaultEnvInt("CHRONOS_HTTP_PORT", 7778),
		HTTPHost: defaultEnv("CHRONOS_HTTP_HOST", "127.0.0.1"),
		APIToken: defaultEnv("CHRONOS_API_TOKEN", ""),
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
	if c.SpikeZScore < 0 {
		return fmt.Errorf("spike z-score must be >= 0, got %f", c.SpikeZScore)
	}
	if c.DropZScore < 0 {
		return fmt.Errorf("drop z-score must be >= 0, got %f", c.DropZScore)
	}
	if c.StallMaxStdDev < 0 {
		return fmt.Errorf("stall max stddev must be >= 0, got %f", c.StallMaxStdDev)
	}
	if c.AnomalyMaxSimilarity < -1 || c.AnomalyMaxSimilarity > 1 {
		return fmt.Errorf("anomaly max similarity must be in [-1, 1], got %f", c.AnomalyMaxSimilarity)
	}
	if c.SeasonalityMinAutocorr < -1 || c.SeasonalityMinAutocorr > 1 {
		return fmt.Errorf("seasonality min autocorrelation must be in [-1, 1], got %f", c.SeasonalityMinAutocorr)
	}
	if c.CorrelationMin < 0 || c.CorrelationMin > 1 {
		return fmt.Errorf("correlation min must be in [0, 1], got %f", c.CorrelationMin)
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
