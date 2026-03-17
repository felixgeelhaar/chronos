package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ascend/api/internal/config"
	"github.com/ascend/api/internal/handler"
	"github.com/ascend/api/internal/middleware"
	"github.com/ascend/api/internal/repository"
	"github.com/ascend/api/internal/service"
	"github.com/ascend/api/pkg/auth"
	"github.com/ascend/api/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MockS3Client is a mock S3Client for testing
type MockS3Client struct {
	files map[string][]byte // key -> file content
	mu    sync.RWMutex
}

// NewMockS3Client creates a new mock S3Client
func NewMockS3Client() *MockS3Client {
	return &MockS3Client{
		files: make(map[string][]byte),
	}
}

// UploadFile mocks uploading a file to S3
func (m *MockS3Client) UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (*storage.UploadFileResult, error) {
	// Read file content
	buffer := new(bytes.Buffer)
	size, err := io.Copy(buffer, file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Generate mock key and URL
	timestamp := time.Now().Format("2006/01/02")
	uniqueID := uuid.New().String()
	ext := filepath.Ext(filename)
	key := fmt.Sprintf("videos/%s/%s%s", timestamp, uniqueID, ext)
	url := fmt.Sprintf("https://test-bucket.s3.amazonaws.com/%s", key)

	// Store file content
	m.mu.Lock()
	m.files[key] = buffer.Bytes()
	m.mu.Unlock()

	return &storage.UploadFileResult{
		Key:      key,
		URL:      url,
		FileSize: size,
	}, nil
}

// DeleteFile mocks deleting a file from S3
func (m *MockS3Client) DeleteFile(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.files, key)
	return nil
}

// GetFile retrieves a file from the mock storage (helper for testing)
func (m *MockS3Client) GetFile(key string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	content, exists := m.files[key]
	return content, exists
}

// TestServer encapsulates the test server and its dependencies
type TestServer struct {
	Router           *gin.Engine
	DB               *gorm.DB
	JWTService       *auth.JWTService
	AuthService      *service.AuthService
	SessionService   *service.SessionService
	AnalyticsService *service.AnalyticsService
	OneRepMaxService *service.OneRepMaxService
	VideoService     *service.VideoService
	MockS3           *MockS3Client
	Container        testcontainers.Container
	TestServer       *httptest.Server
}

// SetupTestServer creates a test server with a real PostgreSQL database using testcontainers
func SetupTestServer(t *testing.T) *TestServer {
	gin.SetMode(gin.TestMode)

	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "test_db",
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Get container host and port
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	mappedPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	// Build database URL
	dbURL := fmt.Sprintf("postgres://test_user:test_password@%s:%s/test_db?sslmode=disable",
		host, mappedPort.Port())

	// Connect to database
	db, err := config.NewDatabase(config.DatabaseConfig{
		URL:            dbURL,
		MaxConnections: 10,
		LogLevel:       logger.Silent, // Use silent mode for tests
	})
	if err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := config.RunMigrations(db); err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize JWT service
	jwtService := auth.NewJWTService(auth.JWTConfig{
		AccessSecret:  "test-access-secret-key-for-integration-tests",
		RefreshSecret: "test-refresh-secret-key-for-integration-tests",
		AccessExpiry:  15 * time.Minute,
		RefreshExpiry: 168 * time.Hour,
	})

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	videoRepo := repository.NewVideoRepository(db)

	// Initialize mock S3 client
	mockS3 := NewMockS3Client()

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtService, 15*time.Minute)
	sessionService := service.NewSessionService(sessionRepo, setRepo, db)
	analyticsService := service.NewAnalyticsService(sessionRepo, setRepo, oneRepMaxRepo)
	oneRepMaxService := service.NewOneRepMaxService(oneRepMaxRepo)
	videoService := service.NewVideoService(videoRepo, mockS3, nil) // nil queue for testing

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsService)
	oneRepMaxHandler := handler.NewOneRepMaxHandler(oneRepMaxService)
	videoHandler := handler.NewVideoHandler(videoService)

	// Setup router
	router := gin.New()

	// Apply middleware (minimal set for testing)
	router.Use(middleware.SecurityHeaders("test"))
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.ErrorHandler())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Public routes - auth endpoints
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.Auth(jwtService))
		{
			// User profile
			protected.GET("/auth/me", authHandler.GetMe)

			// Session routes
			protected.POST("/sessions", sessionHandler.CreateSession)
			protected.GET("/sessions", sessionHandler.ListSessions)
			protected.GET("/sessions/:id", sessionHandler.GetSession)
			protected.PUT("/sessions/:id", sessionHandler.UpdateSession)
			protected.DELETE("/sessions/:id", sessionHandler.DeleteSession)

			// Analytics routes
			protected.GET("/analytics/exercise/:name", analyticsHandler.GetExerciseHistory)
			protected.GET("/analytics/acwr", analyticsHandler.GetACWR)
			protected.GET("/analytics/volume", analyticsHandler.GetVolumeProgress)
			protected.GET("/analytics/summary", analyticsHandler.GetProgressSummary)

			// One Rep Max routes
			protected.POST("/one-rep-maxes", oneRepMaxHandler.CreateOneRepMax)
			protected.GET("/one-rep-maxes", oneRepMaxHandler.ListOneRepMaxes)
			protected.GET("/one-rep-maxes/:id", oneRepMaxHandler.GetOneRepMax)
			protected.GET("/one-rep-maxes/exercise/:name/history", oneRepMaxHandler.GetOneRepMaxHistory)
			protected.PUT("/one-rep-maxes/:id", oneRepMaxHandler.UpdateOneRepMax)
			protected.DELETE("/one-rep-maxes/:id", oneRepMaxHandler.DeleteOneRepMax)

			// Video routes
			protected.POST("/videos", videoHandler.UploadVideo)
			protected.GET("/videos", videoHandler.ListVideos)
			protected.GET("/videos/:id", videoHandler.GetVideo)
			protected.PUT("/videos/:id", videoHandler.UpdateVideo)
			protected.DELETE("/videos/:id", videoHandler.DeleteVideo)
			protected.POST("/videos/:id/presigned-url", videoHandler.GeneratePresignedURL)
			protected.GET("/sessions/:id/videos", videoHandler.ListVideosBySession)
		}
	}

	// Create test server
	testServer := httptest.NewServer(router)

	return &TestServer{
		Router:           router,
		DB:               db,
		JWTService:       jwtService,
		AuthService:      authService,
		SessionService:   sessionService,
		AnalyticsService: analyticsService,
		OneRepMaxService: oneRepMaxService,
		VideoService:     videoService,
		MockS3:           mockS3,
		Container:        postgresContainer,
		TestServer:       testServer,
	}
}

// Teardown cleans up the test server and database container
func (ts *TestServer) Teardown(t *testing.T) {
	// Close test server
	if ts.TestServer != nil {
		ts.TestServer.Close()
	}

	// Close database connection
	if ts.DB != nil {
		sqlDB, err := ts.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// Terminate container
	if ts.Container != nil {
		ctx := context.Background()
		if err := ts.Container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}
}

// CleanDatabase removes all data from the database while preserving schema
func (ts *TestServer) CleanDatabase(t *testing.T) {
	// Delete all data in reverse order of foreign key dependencies
	tables := []string{
		"videos",
		"sets",
		"sessions",
		"one_rep_maxes",
		"users",
	}

	for _, table := range tables {
		if err := ts.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			t.Fatalf("Failed to clean table %s: %v", table, err)
		}
	}
}
