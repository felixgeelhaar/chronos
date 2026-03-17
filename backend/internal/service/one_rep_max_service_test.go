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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupOneRepMaxTestDB creates an in-memory SQLite database for one rep max testing
func setupOneRepMaxTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	// Set connection pool to 1 to avoid SQLite in-memory database isolation issues
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Enable foreign key constraints
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

	return db
}

// teardownOneRepMaxTestDB closes the database connection
func teardownOneRepMaxTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func TestOneRepMaxService_CreateOneRepMax(t *testing.T) {
	db := setupOneRepMaxTestDB(t)
	defer teardownOneRepMaxTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	oneRepMaxService := service.NewOneRepMaxService(oneRepMaxRepo)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "orm@example.com",
		Password: "hashedpassword",
		Name:     "ORM User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	now := time.Now()

	tests := []struct {
		name        string
		userID      uuid.UUID
		request     *dto.CreateOneRepMaxRequest
		expectError bool
		errorMsg    string
		validateFn  func(*testing.T, *dto.OneRepMaxResponse)
	}{
		{
			name:   "successful 1RM creation",
			userID: user.ID,
			request: &dto.CreateOneRepMaxRequest{
				ExerciseName: "Squat",
				Weight:       150.0,
				Date:         now,
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxResponse) {
				assert.Equal(t, user.ID, resp.UserID)
				assert.Equal(t, "Squat", resp.ExerciseName)
				assert.Equal(t, 150.0, resp.Weight)
				assert.Equal(t, now.Unix(), resp.Date.Unix())
			},
		},
		{
			name:   "create 1RM with different exercise",
			userID: user.ID,
			request: &dto.CreateOneRepMaxRequest{
				ExerciseName: "Bench Press",
				Weight:       100.5,
				Date:         now,
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxResponse) {
				assert.Equal(t, "Bench Press", resp.ExerciseName)
				assert.Equal(t, 100.5, resp.Weight)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := oneRepMaxService.CreateOneRepMax(ctx, tt.userID, tt.request)

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

func TestOneRepMaxService_GetOneRepMax(t *testing.T) {
	db := setupOneRepMaxTestDB(t)
	defer teardownOneRepMaxTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	oneRepMaxService := service.NewOneRepMaxService(oneRepMaxRepo)
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

	// Create test 1RM for user1
	oneRepMax := &domain.OneRepMax{
		UserID:       user1.ID,
		ExerciseName: "Squat",
		Weight:       150.0,
		Date:         time.Now(),
	}
	err = oneRepMaxRepo.Create(ctx, oneRepMax)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uuid.UUID
		id          uuid.UUID
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful retrieval of own 1RM",
			userID:      user1.ID,
			id:          oneRepMax.ID,
			expectError: false,
		},
		{
			name:        "unauthorized access to another user's 1RM",
			userID:      user2.ID,
			id:          oneRepMax.ID,
			expectError: true,
			errorMsg:    "unauthorized",
		},
		{
			name:        "non-existent 1RM",
			userID:      user1.ID,
			id:          uuid.New(),
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := oneRepMaxService.GetOneRepMax(ctx, tt.userID, tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, oneRepMax.ID, response.ID)
				assert.Equal(t, user1.ID, response.UserID)
			}
		})
	}
}

func TestOneRepMaxService_ListOneRepMaxes(t *testing.T) {
	db := setupOneRepMaxTestDB(t)
	defer teardownOneRepMaxTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	oneRepMaxService := service.NewOneRepMaxService(oneRepMaxRepo)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "list@example.com",
		Password: "hashedpassword",
		Name:     "List User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple test 1RM records
	now := time.Now()
	for i := 0; i < 5; i++ {
		oneRepMax := &domain.OneRepMax{
			UserID:       user.ID,
			ExerciseName: "Squat",
			Weight:       100.0 + float64(i*10),
			Date:         now.Add(time.Duration(-i) * 24 * time.Hour),
		}
		err = oneRepMaxRepo.Create(ctx, oneRepMax)
		require.NoError(t, err)
	}

	tests := []struct {
		name        string
		userID      uuid.UUID
		page        int
		pageSize    int
		expectError bool
		validateFn  func(*testing.T, *dto.OneRepMaxListResponse)
	}{
		{
			name:        "list 1RM records with default pagination",
			userID:      user.ID,
			page:        1,
			pageSize:    20,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxListResponse) {
				assert.Len(t, resp.Records, 5)
				assert.Equal(t, 5, resp.Total)
				assert.Equal(t, 1, resp.Page)
				assert.Equal(t, 20, resp.PageSize)
			},
		},
		{
			name:        "list 1RM records with custom page size",
			userID:      user.ID,
			page:        1,
			pageSize:    2,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxListResponse) {
				assert.Len(t, resp.Records, 2)
				assert.Equal(t, 5, resp.Total)
				assert.Equal(t, 2, resp.PageSize)
			},
		},
		{
			name:        "list 1RM records page 2",
			userID:      user.ID,
			page:        2,
			pageSize:    2,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxListResponse) {
				assert.Len(t, resp.Records, 2)
				assert.Equal(t, 2, resp.Page)
			},
		},
		{
			name:        "list 1RM records with invalid page (< 1) - defaults to 1",
			userID:      user.ID,
			page:        0,
			pageSize:    20,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxListResponse) {
				assert.Equal(t, 1, resp.Page)
			},
		},
		{
			name:        "list 1RM records with invalid pageSize (< 1) - defaults to 20",
			userID:      user.ID,
			page:        1,
			pageSize:    0,
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxListResponse) {
				assert.Equal(t, 20, resp.PageSize)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := oneRepMaxService.ListOneRepMaxes(ctx, tt.userID, tt.page, tt.pageSize)

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

func TestOneRepMaxService_GetOneRepMaxHistory(t *testing.T) {
	db := setupOneRepMaxTestDB(t)
	defer teardownOneRepMaxTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	oneRepMaxService := service.NewOneRepMaxService(oneRepMaxRepo)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "history@example.com",
		Password: "hashedpassword",
		Name:     "History User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create history of 1RM records for Squat
	now := time.Now()
	squatWeights := []float64{100.0, 110.0, 120.0, 130.0, 140.0}
	for i, weight := range squatWeights {
		oneRepMax := &domain.OneRepMax{
			UserID:       user.ID,
			ExerciseName: "Squat",
			Weight:       weight,
			Date:         now.Add(time.Duration(-i) * 24 * time.Hour),
		}
		err = oneRepMaxRepo.Create(ctx, oneRepMax)
		require.NoError(t, err)
	}

	// Create some records for different exercise
	benchPress := &domain.OneRepMax{
		UserID:       user.ID,
		ExerciseName: "Bench Press",
		Weight:       80.0,
		Date:         now,
	}
	err = oneRepMaxRepo.Create(ctx, benchPress)
	require.NoError(t, err)

	tests := []struct {
		name         string
		userID       uuid.UUID
		exerciseName string
		expectError  bool
		validateFn   func(*testing.T, *dto.OneRepMaxHistoryResponse)
	}{
		{
			name:         "get history for exercise with multiple records",
			userID:       user.ID,
			exerciseName: "Squat",
			expectError:  false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxHistoryResponse) {
				assert.Equal(t, "Squat", resp.ExerciseName)
				assert.Len(t, resp.History, 5)
				assert.Equal(t, 140.0, resp.PersonalBest)
				require.NotNil(t, resp.CurrentRecord)
				assert.Equal(t, 100.0, resp.CurrentRecord.Weight) // Most recent
				// Improvement from 140.0 (earliest) to 100.0 (current) = -28.57%
				assert.InDelta(t, -28.57, resp.Improvement, 0.1)
			},
		},
		{
			name:         "get history for exercise with single record",
			userID:       user.ID,
			exerciseName: "Bench Press",
			expectError:  false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxHistoryResponse) {
				assert.Equal(t, "Bench Press", resp.ExerciseName)
				assert.Len(t, resp.History, 1)
				assert.Equal(t, 80.0, resp.PersonalBest)
				require.NotNil(t, resp.CurrentRecord)
				assert.Equal(t, 0.0, resp.Improvement) // No improvement for single record
			},
		},
		{
			name:         "get history for non-existent exercise",
			userID:       user.ID,
			exerciseName: "Deadlift",
			expectError:  false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxHistoryResponse) {
				assert.Equal(t, "Deadlift", resp.ExerciseName)
				assert.Empty(t, resp.History)
				assert.Nil(t, resp.CurrentRecord)
				assert.Equal(t, 0.0, resp.PersonalBest)
				assert.Equal(t, 0.0, resp.Improvement)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := oneRepMaxService.GetOneRepMaxHistory(ctx, tt.userID, tt.exerciseName)

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

func TestOneRepMaxService_UpdateOneRepMax(t *testing.T) {
	db := setupOneRepMaxTestDB(t)
	defer teardownOneRepMaxTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	oneRepMaxService := service.NewOneRepMaxService(oneRepMaxRepo)
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
	newWeight := 160.0

	tests := []struct {
		name        string
		setupFunc   func() uuid.UUID
		userID      uuid.UUID
		request     *dto.UpdateOneRepMaxRequest
		expectError bool
		errorMsg    string
		validateFn  func(*testing.T, *dto.OneRepMaxResponse)
	}{
		{
			name: "update weight only",
			setupFunc: func() uuid.UUID {
				oneRepMax := &domain.OneRepMax{
					UserID:       user1.ID,
					ExerciseName: "Squat",
					Weight:       150.0,
					Date:         now,
				}
				err := oneRepMaxRepo.Create(ctx, oneRepMax)
				require.NoError(t, err)
				return oneRepMax.ID
			},
			userID: user1.ID,
			request: &dto.UpdateOneRepMaxRequest{
				Weight: &newWeight,
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxResponse) {
				assert.Equal(t, 160.0, resp.Weight)
				assert.Equal(t, now.Unix(), resp.Date.Unix())
			},
		},
		{
			name: "update date only",
			setupFunc: func() uuid.UUID {
				oneRepMax := &domain.OneRepMax{
					UserID:       user1.ID,
					ExerciseName: "Bench Press",
					Weight:       100.0,
					Date:         now,
				}
				err := oneRepMaxRepo.Create(ctx, oneRepMax)
				require.NoError(t, err)
				return oneRepMax.ID
			},
			userID: user1.ID,
			request: &dto.UpdateOneRepMaxRequest{
				Date: &newDate,
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxResponse) {
				assert.Equal(t, 100.0, resp.Weight)
				assert.Equal(t, newDate.Unix(), resp.Date.Unix())
			},
		},
		{
			name: "update both weight and date",
			setupFunc: func() uuid.UUID {
				oneRepMax := &domain.OneRepMax{
					UserID:       user1.ID,
					ExerciseName: "Deadlift",
					Weight:       140.0,
					Date:         now,
				}
				err := oneRepMaxRepo.Create(ctx, oneRepMax)
				require.NoError(t, err)
				return oneRepMax.ID
			},
			userID: user1.ID,
			request: &dto.UpdateOneRepMaxRequest{
				Weight: &newWeight,
				Date:   &newDate,
			},
			expectError: false,
			validateFn: func(t *testing.T, resp *dto.OneRepMaxResponse) {
				assert.Equal(t, 160.0, resp.Weight)
				assert.Equal(t, newDate.Unix(), resp.Date.Unix())
			},
		},
		{
			name: "unauthorized access to another user's 1RM",
			setupFunc: func() uuid.UUID {
				oneRepMax := &domain.OneRepMax{
					UserID:       user1.ID,
					ExerciseName: "Squat",
					Weight:       150.0,
					Date:         now,
				}
				err := oneRepMaxRepo.Create(ctx, oneRepMax)
				require.NoError(t, err)
				return oneRepMax.ID
			},
			userID: user2.ID,
			request: &dto.UpdateOneRepMaxRequest{
				Weight: &newWeight,
			},
			expectError: true,
			errorMsg:    "unauthorized",
		},
		{
			name: "non-existent 1RM",
			setupFunc: func() uuid.UUID {
				return uuid.New()
			},
			userID: user1.ID,
			request: &dto.UpdateOneRepMaxRequest{
				Weight: &newWeight,
			},
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := tt.setupFunc()
			response, err := oneRepMaxService.UpdateOneRepMax(ctx, tt.userID, id, tt.request)

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

func TestOneRepMaxService_DeleteOneRepMax(t *testing.T) {
	db := setupOneRepMaxTestDB(t)
	defer teardownOneRepMaxTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	oneRepMaxService := service.NewOneRepMaxService(oneRepMaxRepo)
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

	now := time.Now()

	tests := []struct {
		name        string
		setupFunc   func() uuid.UUID
		userID      uuid.UUID
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful deletion of own 1RM",
			setupFunc: func() uuid.UUID {
				oneRepMax := &domain.OneRepMax{
					UserID:       user1.ID,
					ExerciseName: "Squat",
					Weight:       150.0,
					Date:         now,
				}
				err := oneRepMaxRepo.Create(ctx, oneRepMax)
				require.NoError(t, err)
				return oneRepMax.ID
			},
			userID:      user1.ID,
			expectError: false,
		},
		{
			name: "unauthorized access to another user's 1RM",
			setupFunc: func() uuid.UUID {
				oneRepMax := &domain.OneRepMax{
					UserID:       user1.ID,
					ExerciseName: "Bench Press",
					Weight:       100.0,
					Date:         now,
				}
				err := oneRepMaxRepo.Create(ctx, oneRepMax)
				require.NoError(t, err)
				return oneRepMax.ID
			},
			userID:      user2.ID,
			expectError: true,
			errorMsg:    "unauthorized",
		},
		{
			name: "non-existent 1RM",
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
			id := tt.setupFunc()
			err := oneRepMaxService.DeleteOneRepMax(ctx, tt.userID, id)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify 1RM is deleted
				_, err := oneRepMaxService.GetOneRepMax(ctx, tt.userID, id)
				assert.Error(t, err, "1RM should not be retrievable after deletion")
			}
		})
	}
}
