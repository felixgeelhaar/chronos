package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ascend/api/internal/domain"
	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/repository"
	"github.com/ascend/api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// sessionTestDB holds PostgreSQL testcontainer and GORM DB
type sessionTestDB struct {
	db        *gorm.DB
	container *postgrescontainer.PostgresContainer
}

// setupSessionTestDB creates a PostgreSQL testcontainer for session testing
func setupSessionTestDB(t *testing.T) *sessionTestDB {
	t.Helper()

	ctx := context.Background()

	// Start PostgreSQL container
	container, err := postgrescontainer.Run(ctx,
		"postgres:15-alpine",
		postgrescontainer.WithDatabase("test_db"),
		postgrescontainer.WithUsername("test_user"),
		postgrescontainer.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Connect to PostgreSQL with GORM
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables using domain models
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Session{},
		&domain.Set{},
		&domain.Video{},
	)
	require.NoError(t, err)

	return &sessionTestDB{
		db:        db,
		container: container,
	}
}

// teardownSessionTestDB stops the PostgreSQL container and closes connections
func teardownSessionTestDB(t *testing.T, testDB *sessionTestDB) {
	t.Helper()

	ctx := context.Background()

	// Close database connection
	sqlDB, err := testDB.db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	// Terminate container
	require.NoError(t, testDB.container.Terminate(ctx))
}

func TestSessionService_CreateSession(t *testing.T) {
	testDB := setupSessionTestDB(t)
	defer teardownSessionTestDB(t, testDB)

	userRepo := repository.NewUserRepository(testDB.db)
	sessionRepo := repository.NewSessionRepository(testDB.db)
	setRepo := repository.NewSetRepository(testDB.db)
	sessionService := service.NewSessionService(sessionRepo, setRepo, testDB.db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "session@example.com",
		Password: "hashedpassword",
		Name:     "Session User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	now := time.Now()

	tests := []struct {
		name        string
		userID      uuid.UUID
		request     *dto.CreateSessionRequest
		expectError bool
		errorMsg    string
		validateFn  func(*testing.T, *dto.SessionResponse)
	}{
		{
			name:   "successful session creation without sets",
			userID: user.ID,
			request: &dto.CreateSessionRequest{
				Date:  now,
				Notes: stringPtr("Morning workout"),
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionResponse) {
				assert.Equal(t, user.ID, resp.UserID)
				// Compare dates truncated to day (PostgreSQL may handle timestamps differently)
				assert.Equal(t, now.Truncate(24*time.Hour).Unix(), resp.Date.Truncate(24*time.Hour).Unix())
				assert.Equal(t, "Morning workout", *resp.Notes)
				assert.Empty(t, resp.Sets)
			},
		},
		{
			name:   "successful session creation with multiple sets",
			userID: user.ID,
			request: &dto.CreateSessionRequest{
				Date:  now,
				Notes: stringPtr("Leg day"),
				Sets: []dto.CreateSetInput{
					{
						ExerciseName: "Squat",
						Weight:       100.0,
						Reps:         5,
						RPE:          floatPtr(8.5),
						Notes:        stringPtr("Felt good"),
						SetOrder:     1,
					},
					{
						ExerciseName: "Squat",
						Weight:       110.0,
						Reps:         3,
						RPE:          floatPtr(9.0),
						SetOrder:     2,
					},
					{
						ExerciseName: "Deadlift",
						Weight:       140.0,
						Reps:         5,
						RPE:          floatPtr(8.0),
						SetOrder:     3,
					},
				},
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionResponse) {
				assert.Equal(t, user.ID, resp.UserID)
				assert.Equal(t, "Leg day", *resp.Notes)
				assert.Len(t, resp.Sets, 3)
				assert.Equal(t, "Squat", resp.Sets[0].ExerciseName)
				assert.Equal(t, 100.0, resp.Sets[0].Weight)
				assert.Equal(t, 5, resp.Sets[0].Reps)
				assert.Equal(t, 8.5, *resp.Sets[0].RPE)
			},
		},
		{
			name:   "session creation with minimal set data",
			userID: user.ID,
			request: &dto.CreateSessionRequest{
				Date: now,
				Sets: []dto.CreateSetInput{
					{
						ExerciseName: "Bench Press",
						Weight:       80.0,
						Reps:         8,
						SetOrder:     1,
					},
				},
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionResponse) {
				assert.Len(t, resp.Sets, 1)
				assert.Nil(t, resp.Sets[0].RPE)
				assert.Nil(t, resp.Sets[0].Notes)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := sessionService.CreateSession(ctx, tt.userID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.NotEqual(t, uuid.Nil, response.ID)
				if tt.validateFn != nil {
					tt.validateFn(t, response)
				}
			}
		})
	}
}

func TestSessionService_GetSession(t *testing.T) {
	testDB := setupSessionTestDB(t)
	defer teardownSessionTestDB(t, testDB)

	userRepo := repository.NewUserRepository(testDB.db)
	sessionRepo := repository.NewSessionRepository(testDB.db)
	setRepo := repository.NewSetRepository(testDB.db)
	sessionService := service.NewSessionService(sessionRepo, setRepo, testDB.db)
	ctx := context.Background()

	// Create test users
	user1 := &domain.User{
		Email:    "user1@example.com",
		Password: "hashedpassword",
		Name:     "User 1",
	}
	err := userRepo.Create(ctx, user1)
	require.NoError(t, err)

	user2 := &domain.User{
		Email:    "user2@example.com",
		Password: "hashedpassword",
		Name:     "User 2",
	}
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	// Create test session for user1
	now := time.Now()
	session := &domain.Session{
		UserID: user1.ID,
		Date:   now,
		Notes:  stringPtr("Test session"),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	// Create sets for the session
	sets := []*domain.Set{
		{
			SessionID:    session.ID,
			ExerciseName: "Squat",
			Weight:       100.0,
			Reps:         5,
			SetOrder:     1,
		},
		{
			SessionID:    session.ID,
			ExerciseName: "Deadlift",
			Weight:       140.0,
			Reps:         3,
			SetOrder:     2,
		},
	}
	err = setRepo.CreateBulk(ctx, sets)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uuid.UUID
		sessionID   uuid.UUID
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful retrieval of own session",
			userID:      user1.ID,
			sessionID:   session.ID,
			expectError: false,
		},
		{
			name:        "unauthorized access to another user's session",
			userID:      user2.ID,
			sessionID:   session.ID,
			expectError: true,
			errorMsg:    "unauthorized",
		},
		{
			name:        "non-existent session",
			userID:      user1.ID,
			sessionID:   uuid.New(),
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := sessionService.GetSession(ctx, tt.userID, tt.sessionID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, session.ID, response.ID)
				assert.Equal(t, user1.ID, response.UserID)
				assert.Len(t, response.Sets, 2)
			}
		})
	}
}

func TestSessionService_ListSessions(t *testing.T) {
	testDB := setupSessionTestDB(t)
	defer teardownSessionTestDB(t, testDB)

	userRepo := repository.NewUserRepository(testDB.db)
	sessionRepo := repository.NewSessionRepository(testDB.db)
	setRepo := repository.NewSetRepository(testDB.db)
	sessionService := service.NewSessionService(sessionRepo, setRepo, testDB.db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "list@example.com",
		Password: "hashedpassword",
		Name:     "List User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple test sessions
	now := time.Now()
	for i := 0; i < 5; i++ {
		session := &domain.Session{
			UserID: user.ID,
			Date:   now.Add(time.Duration(-i) * 24 * time.Hour),
			Notes:  stringPtr("Session " + string(rune('A'+i))),
		}
		err = sessionRepo.Create(ctx, session)
		require.NoError(t, err)
	}

	tests := []struct {
		name         string
		userID       uuid.UUID
		page         int
		pageSize     int
		expectError  bool
		validateFn   func(*testing.T, *dto.SessionListResponse)
	}{
		{
			name:        "list sessions with default pagination",
			userID:      user.ID,
			page:        1,
			pageSize:    20,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionListResponse) {
				assert.Len(t, resp.Sessions, 5)
				assert.Equal(t, 5, resp.Total)
				assert.Equal(t, 1, resp.Page)
				assert.Equal(t, 20, resp.PageSize)
			},
		},
		{
			name:        "list sessions with custom page size",
			userID:      user.ID,
			page:        1,
			pageSize:    2,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionListResponse) {
				assert.LessOrEqual(t, len(resp.Sessions), 2)
				assert.Equal(t, 2, resp.PageSize)
			},
		},
		{
			name:        "list sessions with invalid page (< 1) - defaults to 1",
			userID:      user.ID,
			page:        0,
			pageSize:    20,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionListResponse) {
				assert.Equal(t, 1, resp.Page)
			},
		},
		{
			name:        "list sessions with invalid pageSize (< 1) - defaults to 20",
			userID:      user.ID,
			page:        1,
			pageSize:    0,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionListResponse) {
				assert.Equal(t, 20, resp.PageSize)
			},
		},
		{
			name:        "list sessions with pageSize > 100 - capped at 20",
			userID:      user.ID,
			page:        1,
			pageSize:    150,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionListResponse) {
				assert.Equal(t, 20, resp.PageSize)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := sessionService.ListSessions(ctx, tt.userID, tt.page, tt.pageSize)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				if tt.validateFn != nil {
					tt.validateFn(t, response)
				}
			}
		})
	}
}

func TestSessionService_UpdateSession(t *testing.T) {
	testDB := setupSessionTestDB(t)
	defer teardownSessionTestDB(t, testDB)

	userRepo := repository.NewUserRepository(testDB.db)
	sessionRepo := repository.NewSessionRepository(testDB.db)
	setRepo := repository.NewSetRepository(testDB.db)
	sessionService := service.NewSessionService(sessionRepo, setRepo, testDB.db)
	ctx := context.Background()

	// Create test users
	user1 := &domain.User{
		Email:    "update1@example.com",
		Password: "hashedpassword",
		Name:     "Update User 1",
	}
	err := userRepo.Create(ctx, user1)
	require.NoError(t, err)

	user2 := &domain.User{
		Email:    "update2@example.com",
		Password: "hashedpassword",
		Name:     "Update User 2",
	}
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	now := time.Now()
	newDate := now.Add(24 * time.Hour)

	tests := []struct {
		name        string
		setupFunc   func() uuid.UUID
		userID      uuid.UUID
		request     *dto.UpdateSessionRequest
		expectError bool
		errorMsg    string
		validateFn  func(*testing.T, *dto.SessionResponse)
	}{
		{
			name: "update session date only",
			setupFunc: func() uuid.UUID {
				session := &domain.Session{
					UserID: user1.ID,
					Date:   now,
					Notes:  stringPtr("Original notes"),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)
				return session.ID
			},
			userID: user1.ID,
			request: &dto.UpdateSessionRequest{
				Date: &newDate,
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionResponse) {
				// Compare dates truncated to day (PostgreSQL may handle timestamps differently)
				assert.Equal(t, newDate.Truncate(24*time.Hour).Unix(), resp.Date.Truncate(24*time.Hour).Unix())
				assert.Equal(t, "Original notes", *resp.Notes)
			},
		},
		{
			name: "update session notes only",
			setupFunc: func() uuid.UUID {
				session := &domain.Session{
					UserID: user1.ID,
					Date:   now,
					Notes:  stringPtr("Original notes"),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)
				return session.ID
			},
			userID: user1.ID,
			request: &dto.UpdateSessionRequest{
				Notes: stringPtr("Updated notes"),
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionResponse) {
				assert.Equal(t, "Updated notes", *resp.Notes)
			},
		},
		{
			name: "update session with new sets (replace all)",
			setupFunc: func() uuid.UUID {
				session := &domain.Session{
					UserID: user1.ID,
					Date:   now,
					Notes:  stringPtr("Original notes"),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)

				// Create initial sets
				initialSets := []*domain.Set{
					{
						SessionID:    session.ID,
						ExerciseName: "Squat",
						Weight:       100.0,
						Reps:         5,
						SetOrder:     1,
					},
				}
				err = setRepo.CreateBulk(ctx, initialSets)
				require.NoError(t, err)

				return session.ID
			},
			userID: user1.ID,
			request: &dto.UpdateSessionRequest{
				Sets: []dto.CreateSetInput{
					{
						ExerciseName: "Deadlift",
						Weight:       140.0,
						Reps:         3,
						SetOrder:     1,
					},
					{
						ExerciseName: "Deadlift",
						Weight:       150.0,
						Reps:         2,
						SetOrder:     2,
					},
				},
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.SessionResponse) {
				assert.Len(t, resp.Sets, 2)
				assert.Equal(t, "Deadlift", resp.Sets[0].ExerciseName)
				assert.Equal(t, 140.0, resp.Sets[0].Weight)
			},
		},
		{
			name: "unauthorized access to another user's session",
			setupFunc: func() uuid.UUID {
				session := &domain.Session{
					UserID: user1.ID,
					Date:   now,
					Notes:  stringPtr("Protected session"),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)
				return session.ID
			},
			userID: user2.ID,
			request: &dto.UpdateSessionRequest{
				Notes: stringPtr("Hacked"),
			},
			expectError: true,
			errorMsg:    "unauthorized",
		},
		{
			name: "non-existent session",
			setupFunc: func() uuid.UUID {
				return uuid.New()
			},
			userID: user1.ID,
			request: &dto.UpdateSessionRequest{
				Notes: stringPtr("New notes"),
			},
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID := tt.setupFunc()
			response, err := sessionService.UpdateSession(ctx, tt.userID, sessionID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				if tt.validateFn != nil {
					tt.validateFn(t, response)
				}
			}
		})
	}
}

func TestSessionService_DeleteSession(t *testing.T) {
	testDB := setupSessionTestDB(t)
	defer teardownSessionTestDB(t, testDB)

	userRepo := repository.NewUserRepository(testDB.db)
	sessionRepo := repository.NewSessionRepository(testDB.db)
	setRepo := repository.NewSetRepository(testDB.db)
	sessionService := service.NewSessionService(sessionRepo, setRepo, testDB.db)
	ctx := context.Background()

	// Create test users
	user1 := &domain.User{
		Email:    "delete1@example.com",
		Password: "hashedpassword",
		Name:     "Delete User 1",
	}
	err := userRepo.Create(ctx, user1)
	require.NoError(t, err)

	user2 := &domain.User{
		Email:    "delete2@example.com",
		Password: "hashedpassword",
		Name:     "Delete User 2",
	}
	err = userRepo.Create(ctx, user2)
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupFunc   func() uuid.UUID
		userID      uuid.UUID
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful deletion of own session",
			setupFunc: func() uuid.UUID {
				session := &domain.Session{
					UserID: user1.ID,
					Date:   time.Now(),
					Notes:  stringPtr("To be deleted"),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)
				return session.ID
			},
			userID:      user1.ID,
			expectError: false,
		},
		{
			name: "unauthorized access to another user's session",
			setupFunc: func() uuid.UUID {
				session := &domain.Session{
					UserID: user1.ID,
					Date:   time.Now(),
					Notes:  stringPtr("Protected session"),
				}
				err := sessionRepo.Create(ctx, session)
				require.NoError(t, err)
				return session.ID
			},
			userID:      user2.ID,
			expectError: true,
			errorMsg:    "unauthorized",
		},
		{
			name: "non-existent session",
			setupFunc: func() uuid.UUID {
				return uuid.New()
			},
			userID:      user1.ID,
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID := tt.setupFunc()
			err := sessionService.DeleteSession(ctx, tt.userID, sessionID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify session is deleted
				_, err := sessionService.GetSession(ctx, tt.userID, sessionID)
				assert.Error(t, err, "Session should not be retrievable after deletion")
			}
		})
	}
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
