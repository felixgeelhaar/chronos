package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db            *gorm.DB
	startTime     time.Time
	version       string
	environment   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB, version, environment string) *HealthHandler {
	return &HealthHandler{
		db:          db,
		startTime:   time.Now(),
		version:     version,
		environment: environment,
	}
}

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      HealthStatus           `json:"status"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Timestamp   string                 `json:"timestamp"`
	Uptime      string                 `json:"uptime"`
	Checks      map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of an individual health check
type CheckResult struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Latency string       `json:"latency,omitempty"`
}

// LivenessProbe indicates if the application is running
// Kubernetes liveness probe: /health/live
// Returns 200 if app is alive, 503 if dead (triggers restart)
func (h *HealthHandler) LivenessProbe(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "alive",
		"service": "ascend-api",
		"version": h.version,
		"uptime":  time.Since(h.startTime).String(),
	})
}

// ReadinessProbe indicates if the application is ready to serve traffic
// Kubernetes readiness probe: /health/ready
// Returns 200 if ready, 503 if not ready (removes from load balancer)
func (h *HealthHandler) ReadinessProbe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)
	overallStatus := HealthStatusHealthy

	// Check database connectivity
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status != HealthStatusHealthy {
		overallStatus = HealthStatusUnhealthy
	}

	// If unhealthy, return 503 Service Unavailable
	statusCode := http.StatusOK
	if overallStatus == HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:      overallStatus,
		Service:     "ascend-api",
		Version:     h.version,
		Environment: h.environment,
		Timestamp:   time.Now().Format(time.RFC3339),
		Uptime:      time.Since(h.startTime).String(),
		Checks:      checks,
	})
}

// StartupProbe indicates if the application has started successfully
// Kubernetes startup probe: /health/startup
// Used during initial startup, can have longer timeout
func (h *HealthHandler) StartupProbe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)
	overallStatus := HealthStatusHealthy

	// Check database connectivity with longer timeout for startup
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status != HealthStatusHealthy {
		overallStatus = HealthStatusUnhealthy
	}

	// Check database migrations
	migrationCheck := h.checkMigrations(ctx)
	checks["migrations"] = migrationCheck
	if migrationCheck.Status != HealthStatusHealthy {
		overallStatus = HealthStatusUnhealthy
	}

	statusCode := http.StatusOK
	if overallStatus == HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, HealthResponse{
		Status:      overallStatus,
		Service:     "ascend-api",
		Version:     h.version,
		Environment: h.environment,
		Timestamp:   time.Now().Format(time.RFC3339),
		Uptime:      time.Since(h.startTime).String(),
		Checks:      checks,
	})
}

// HealthCheck provides comprehensive health status
// GET /health
// Detailed health check with all dependencies
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)
	overallStatus := HealthStatusHealthy

	// Database check
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status == HealthStatusUnhealthy {
		overallStatus = HealthStatusUnhealthy
	} else if dbCheck.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
		overallStatus = HealthStatusDegraded
	}

	// Database pool check
	poolCheck := h.checkDatabasePool(ctx)
	checks["database_pool"] = poolCheck
	if poolCheck.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
		overallStatus = HealthStatusDegraded
	}

	statusCode := http.StatusOK
	if overallStatus == HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if overallStatus == HealthStatusDegraded {
		statusCode = http.StatusOK // Still return 200, but indicate degraded in body
	}

	c.JSON(statusCode, HealthResponse{
		Status:      overallStatus,
		Service:     "ascend-api",
		Version:     h.version,
		Environment: h.environment,
		Timestamp:   time.Now().Format(time.RFC3339),
		Uptime:      time.Since(h.startTime).String(),
		Checks:      checks,
	})
}

// checkDatabase verifies database connectivity
func (h *HealthHandler) checkDatabase(ctx context.Context) CheckResult {
	start := time.Now()

	sqlDB, err := h.db.DB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get database handle")
		return CheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "Failed to get database handle",
			Latency: time.Since(start).String(),
		}
	}

	// Ping database with timeout
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Error().Err(err).Msg("Database ping failed")
		return CheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "Database unreachable",
			Latency: time.Since(start).String(),
		}
	}

	latency := time.Since(start)

	// Warn if latency is high (> 100ms)
	if latency > 100*time.Millisecond {
		return CheckResult{
			Status:  HealthStatusDegraded,
			Message: "High database latency",
			Latency: latency.String(),
		}
	}

	return CheckResult{
		Status:  HealthStatusHealthy,
		Message: "Database reachable",
		Latency: latency.String(),
	}
}

// checkDatabasePool verifies database connection pool health
func (h *HealthHandler) checkDatabasePool(ctx context.Context) CheckResult {
	sqlDB, err := h.db.DB()
	if err != nil {
		return CheckResult{
			Status:  HealthStatusUnhealthy,
			Message: "Failed to get database handle",
		}
	}

	stats := sqlDB.Stats()

	// Check if too many connections are in use
	maxConns := stats.MaxOpenConnections
	inUse := stats.InUse

	usagePercent := float64(inUse) / float64(maxConns) * 100

	if usagePercent > 90 {
		return CheckResult{
			Status:  HealthStatusDegraded,
			Message: "Connection pool nearly exhausted",
		}
	}

	// Check for connection wait time
	if stats.WaitCount > 0 {
		avgWaitTime := stats.WaitDuration / time.Duration(stats.WaitCount)
		if avgWaitTime > 100*time.Millisecond {
			return CheckResult{
				Status:  HealthStatusDegraded,
				Message: "High connection pool wait times",
			}
		}
	}

	return CheckResult{
		Status:  HealthStatusHealthy,
		Message: "Connection pool healthy",
	}
}

// checkMigrations verifies database schema is up to date
func (h *HealthHandler) checkMigrations(ctx context.Context) CheckResult {
	// Check if critical tables exist
	tables := []string{"users", "sessions", "sets", "one_rep_maxes", "videos"}

	for _, table := range tables {
		if !h.db.Migrator().HasTable(table) {
			return CheckResult{
				Status:  HealthStatusUnhealthy,
				Message: "Missing required database table: " + table,
			}
		}
	}

	return CheckResult{
		Status:  HealthStatusHealthy,
		Message: "Database schema initialized",
	}
}

// MetricsHandler returns Prometheus-compatible metrics
// GET /metrics
func (h *HealthHandler) MetricsHandler(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get database stats")
		return
	}

	stats := sqlDB.Stats()

	// Prometheus format metrics
	metrics := `# HELP ascend_api_uptime_seconds Time since application started
# TYPE ascend_api_uptime_seconds gauge
ascend_api_uptime_seconds ` + time.Since(h.startTime).String() + `

# HELP ascend_db_open_connections Number of open database connections
# TYPE ascend_db_open_connections gauge
ascend_db_open_connections ` + string(rune(stats.OpenConnections)) + `

# HELP ascend_db_in_use_connections Number of connections currently in use
# TYPE ascend_db_in_use_connections gauge
ascend_db_in_use_connections ` + string(rune(stats.InUse)) + `

# HELP ascend_db_idle_connections Number of idle connections
# TYPE ascend_db_idle_connections gauge
ascend_db_idle_connections ` + string(rune(stats.Idle)) + `

# HELP ascend_db_wait_count_total Total number of connections waited for
# TYPE ascend_db_wait_count_total counter
ascend_db_wait_count_total ` + string(rune(int(stats.WaitCount))) + `
`

	c.String(http.StatusOK, metrics)
}
