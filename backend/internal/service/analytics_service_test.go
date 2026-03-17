package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/ascend/api/internal/domain"
	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/repository"
	"github.com/ascend/api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAnalyticsTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	// Set connection pool to 1 for SQLite
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Enable foreign key constraints
	_, err = sqlDB.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Create all required tables
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

	_, err = sqlDB.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			date DATETIME NOT NULL,
			notes TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	require.NoError(t, err)

	_, err = sqlDB.Exec(`
		CREATE TABLE sets (
			id TEXT PRIMARY KEY NOT NULL,
			session_id TEXT NOT NULL,
			exercise_name TEXT NOT NULL,
			weight REAL NOT NULL,
			reps INTEGER NOT NULL,
			rpe REAL,
			notes TEXT,
			set_order INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)
	`)
	require.NoError(t, err)

	_, err = sqlDB.Exec(`
		CREATE TABLE one_rep_maxes (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			exercise_name TEXT NOT NULL,
			weight REAL NOT NULL,
			date DATETIME NOT NULL,
			notes TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	require.NoError(t, err)

	return db
}

func TestAnalyticsService_GetExerciseHistory(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer teardownTestDB(t, db)

	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	analyticsService := service.NewAnalyticsService(sessionRepo, setRepo, oneRepMaxRepo)

	ctx := context.Background()

	// Create user
	user := &domain.User{
		Email:    "analytics@example.com",
		Password: "hashedpassword",
		Name:     "Analytics User",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	now := time.Now()
	startDate := now.AddDate(0, 0, -30) // 30 days ago
	endDate := now

	tests := []struct {
		name         string
		setupFunc    func()
		exerciseName string
		expectError  bool
		validateFn   func(*testing.T, interface{})
	}{
		{
			name: "exercise with multiple sets across sessions",
			setupFunc: func() {
				// Create 3 sessions with squat sets
				for i := 0; i < 3; i++ {
					session := &domain.Session{
						UserID: user.ID,
						Date:   now.AddDate(0, 0, -i*5),
					}
					err := sessionRepo.Create(ctx, session)
					require.NoError(t, err)

					// Add 3 sets per session
					for j := 0; j < 3; j++ {
						set := &domain.Set{
							SessionID:    session.ID,
							ExerciseName: "Squat",
							Weight:       100.0 + float64(i*10),
							Reps:         5,
							SetOrder:     j + 1,
						}
						err = setRepo.Create(ctx, set)
						require.NoError(t, err)
					}
				}

				// Create one rep max
				orm := &domain.OneRepMax{
					UserID:       user.ID,
					ExerciseName: "Squat",
					Weight:       140.0,
					Date:         now.AddDate(0, 0, -10),
				}
				err := oneRepMaxRepo.Create(ctx, orm)
				require.NoError(t, err)
			},
			exerciseName: "Squat",
			expectError:  false,
			validateFn: func(t *testing.T, result interface{}) {
				resp := result.(*dto.ExerciseHistoryResponse)
				assert.Equal(t, "Squat", resp.ExerciseName)
				assert.Len(t, resp.Records, 9) // 3 sessions * 3 sets
				assert.Equal(t, 9, resp.TotalSets)
				// Volume = (100*5 + 100*5 + 100*5) + (110*5 + 110*5 + 110*5) + (120*5 + 120*5 + 120*5)
				// = 1500 + 1650 + 1800 = 4950
				assert.Equal(t, 4950.0, resp.TotalVolume)
				require.NotNil(t, resp.OneRepMax)
				assert.Equal(t, 140.0, resp.OneRepMax.Weight)
			},
		},
		{
			name: "exercise with no history",
			setupFunc: func() {
				// No sessions created for this test
			},
			exerciseName: "Deadlift",
			expectError:  false,
			validateFn: func(t *testing.T, result interface{}) {
				resp := result.(*dto.ExerciseHistoryResponse)
				assert.Equal(t, "Deadlift", resp.ExerciseName)
				assert.Len(t, resp.Records, 0)
				assert.Equal(t, 0, resp.TotalSets)
				assert.Equal(t, 0.0, resp.TotalVolume)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()

			result, err := analyticsService.GetExerciseHistory(ctx, user.ID, tt.exerciseName, startDate, endDate)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateFn != nil {
					tt.validateFn(t, result)
				}
			}
		})
	}
}

func TestAnalyticsService_CalculateACWR(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer teardownTestDB(t, db)

	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	analyticsService := service.NewAnalyticsService(sessionRepo, setRepo, oneRepMaxRepo)

	ctx := context.Background()

	// Create user
	user := &domain.User{
		Email:    "acwr@example.com",
		Password: "hashedpassword",
		Name:     "ACWR User",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	now := time.Now()

	tests := []struct {
		name         string
		setupFunc    func()
		exerciseName string
		expectError  bool
		validateFn   func(*testing.T, interface{})
	}{
		{
			name: "optimal ACWR (0.8-1.3)",
			setupFunc: func() {
				// Create consistent training for 28 days
				// Chronic: 500 volume/day for 28 days = 500 average
				// Acute: 500 volume/day for 7 days = 500 average
				// ACWR = 500/500 = 1.0 (optimal)
				for i := 0; i < 28; i++ {
					session := &domain.Session{
						UserID: user.ID,
						Date:   now.AddDate(0, 0, -i),
					}
					err := sessionRepo.Create(ctx, session)
					require.NoError(t, err)

					set := &domain.Set{
						SessionID:    session.ID,
						ExerciseName: "Bench Press",
						Weight:       100.0,
						Reps:         5,
						SetOrder:     1,
					}
					err = setRepo.Create(ctx, set)
					require.NoError(t, err)
				}
			},
			exerciseName: "Bench Press",
			expectError:  false,
			validateFn: func(t *testing.T, result interface{}) {
				resp := result.(*dto.ACWRResponse)
				assert.Equal(t, "Bench Press", resp.ExerciseName)
				assert.InDelta(t, 1.0, resp.CurrentACWR, 0.05) // Allow 5% tolerance due to date ranges
				assert.Equal(t, "optimal", resp.Status)
				assert.Contains(t, resp.Recommendation, "optimal range")
			},
		},
		{
			name: "high risk ACWR (>1.3)",
			setupFunc: func() {
				// Chronic: low volume for weeks 2-4, high volume for week 1
				// This creates a spike scenario
				// Week 4-2 (days 28-8): 200 volume/day
				for i := 28; i > 7; i-- {
					session := &domain.Session{
						UserID: user.ID,
						Date:   now.AddDate(0, 0, -i),
					}
					err := sessionRepo.Create(ctx, session)
					require.NoError(t, err)

					set := &domain.Set{
						SessionID:    session.ID,
						ExerciseName: "Overhead Press",
						Weight:       40.0,
						Reps:         5,
						SetOrder:     1,
					}
					err = setRepo.Create(ctx, set)
					require.NoError(t, err)
				}

				// Week 1 (days 7-0): 800 volume/day (spike)
				for i := 7; i >= 0; i-- {
					session := &domain.Session{
						UserID: user.ID,
						Date:   now.AddDate(0, 0, -i),
					}
					err := sessionRepo.Create(ctx, session)
					require.NoError(t, err)

					set := &domain.Set{
						SessionID:    session.ID,
						ExerciseName: "Overhead Press",
						Weight:       160.0,
						Reps:         5,
						SetOrder:     1,
					}
					err = setRepo.Create(ctx, set)
					require.NoError(t, err)
				}
			},
			exerciseName: "Overhead Press",
			expectError:  false,
			validateFn: func(t *testing.T, result interface{}) {
				resp := result.(*dto.ACWRResponse)
				assert.Equal(t, "Overhead Press", resp.ExerciseName)
				assert.Greater(t, resp.CurrentACWR, 1.3)
				assert.Equal(t, "high_risk", resp.Status)
				assert.Contains(t, resp.Recommendation, "too high")
			},
		},
		{
			name: "undertraining ACWR (<0.8)",
			setupFunc: func() {
				// High volume for weeks 2-4, low volume for week 1
				// This creates detraining scenario
				// Week 4-2 (days 28-8): 800 volume/day
				for i := 28; i > 7; i-- {
					session := &domain.Session{
						UserID: user.ID,
						Date:   now.AddDate(0, 0, -i),
					}
					err := sessionRepo.Create(ctx, session)
					require.NoError(t, err)

					set := &domain.Set{
						SessionID:    session.ID,
						ExerciseName: "Pull-ups",
						Weight:       160.0,
						Reps:         5,
						SetOrder:     1,
					}
					err = setRepo.Create(ctx, set)
					require.NoError(t, err)
				}

				// Week 1 (days 7-0): 200 volume/day (low)
				for i := 7; i >= 0; i-- {
					session := &domain.Session{
						UserID: user.ID,
						Date:   now.AddDate(0, 0, -i),
					}
					err := sessionRepo.Create(ctx, session)
					require.NoError(t, err)

					set := &domain.Set{
						SessionID:    session.ID,
						ExerciseName: "Pull-ups",
						Weight:       40.0,
						Reps:         5,
						SetOrder:     1,
					}
					err = setRepo.Create(ctx, set)
					require.NoError(t, err)
				}
			},
			exerciseName: "Pull-ups",
			expectError:  false,
			validateFn: func(t *testing.T, result interface{}) {
				resp := result.(*dto.ACWRResponse)
				assert.Equal(t, "Pull-ups", resp.ExerciseName)
				assert.Less(t, resp.CurrentACWR, 0.8)
				assert.Equal(t, "undertraining", resp.Status)
				assert.Contains(t, resp.Recommendation, "increasing")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up previous test data
			db.Exec("DELETE FROM sets")
			db.Exec("DELETE FROM sessions")

			tt.setupFunc()

			result, err := analyticsService.CalculateACWR(ctx, user.ID, tt.exerciseName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateFn != nil {
					tt.validateFn(t, result)
				}
			}
		})
	}
}

func TestAnalyticsService_GetVolumeProgress(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer teardownTestDB(t, db)

	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	analyticsService := service.NewAnalyticsService(sessionRepo, setRepo, oneRepMaxRepo)

	ctx := context.Background()

	// Create user
	user := &domain.User{
		Email:    "volume@example.com",
		Password: "hashedpassword",
		Name:     "Volume User",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	now := time.Now()

	// Create test data: increasing volume over 14 days
	for i := 14; i >= 0; i-- {
		session := &domain.Session{
			UserID: user.ID,
			Date:   now.AddDate(0, 0, -i),
		}
		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)

		// Increasing weight over time
		weight := 100.0 + float64(14-i)*5.0
		set := &domain.Set{
			SessionID:    session.ID,
			ExerciseName: "Row",
			Weight:       weight,
			Reps:         5,
			SetOrder:     1,
		}
		err = setRepo.Create(ctx, set)
		require.NoError(t, err)
	}

	tests := []struct {
		name         string
		exerciseName string
		period       string
		expectError  bool
		validateFn   func(*testing.T, interface{})
	}{
		{
			name:         "week period",
			exerciseName: "Row",
			period:       "week",
			expectError:  false,
			validateFn: func(t *testing.T, result interface{}) {
				resp := result.(*dto.VolumeProgressResponse)
				assert.Equal(t, "Row", resp.ExerciseName)
				assert.Equal(t, "week", resp.Period)
				assert.GreaterOrEqual(t, len(resp.DataPoints), 7)
				assert.Greater(t, resp.TotalVolume, 0.0)
				assert.Greater(t, resp.AverageVolume, 0.0)
				assert.NotEmpty(t, resp.Trend) // Trend is calculated (increasing/decreasing/stable)
			},
		},
		{
			name:         "month period",
			exerciseName: "Row",
			period:       "month",
			expectError:  false,
			validateFn: func(t *testing.T, result interface{}) {
				resp := result.(*dto.VolumeProgressResponse)
				assert.Equal(t, "Row", resp.ExerciseName)
				assert.Equal(t, "month", resp.Period)
				assert.GreaterOrEqual(t, len(resp.DataPoints), 14)
				assert.Greater(t, resp.TotalVolume, 0.0)
			},
		},
		{
			name:         "invalid period defaults to month",
			exerciseName: "Row",
			period:       "invalid",
			expectError:  false,
			validateFn: func(t *testing.T, result interface{}) {
				resp := result.(*dto.VolumeProgressResponse)
				assert.Equal(t, "month", resp.Period) // Should default to month
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyticsService.GetVolumeProgress(ctx, user.ID, tt.exerciseName, tt.period)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				if tt.validateFn != nil {
					tt.validateFn(t, result)
				}
			}
		})
	}
}

func TestAnalyticsService_GetProgressSummary(t *testing.T) {
	db := setupAnalyticsTestDB(t)
	defer teardownTestDB(t, db)

	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	oneRepMaxRepo := repository.NewOneRepMaxRepository(db)
	analyticsService := service.NewAnalyticsService(sessionRepo, setRepo, oneRepMaxRepo)

	ctx := context.Background()

	// Create user
	user := &domain.User{
		Email:    "summary@example.com",
		Password: "hashedpassword",
		Name:     "Summary User",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	now := time.Now()
	startDate := now.AddDate(0, 0, -30)
	endDate := now

	// Create 5 sessions with mixed exercises
	exercises := []string{"Squat", "Bench Press", "Deadlift"}
	for i := 0; i < 5; i++ {
		session := &domain.Session{
			UserID: user.ID,
			Date:   now.AddDate(0, 0, -i*6),
		}
		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)

		// Add sets for each exercise
		for _, exercise := range exercises {
			for j := 0; j < 3; j++ {
				set := &domain.Set{
					SessionID:    session.ID,
					ExerciseName: exercise,
					Weight:       100.0 + float64(i*10),
					Reps:         5,
					SetOrder:     j + 1,
				}
				err = setRepo.Create(ctx, set)
				require.NoError(t, err)
			}
		}
	}

	// Create one rep maxes
	for _, exercise := range exercises {
		orm := &domain.OneRepMax{
			UserID:       user.ID,
			ExerciseName: exercise,
			Weight:       150.0,
			Date:         now.AddDate(0, 0, -15),
		}
		err := oneRepMaxRepo.Create(ctx, orm)
		require.NoError(t, err)
	}

	result, err := analyticsService.GetProgressSummary(ctx, user.ID, startDate, endDate)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 5, result.TotalSessions)
	assert.Equal(t, 45, result.TotalSets) // 5 sessions * 3 exercises * 3 sets
	assert.Greater(t, result.TotalVolume, 0.0)
	assert.Len(t, result.TopExercises, 3)
	assert.Len(t, result.OneRepMaxRecords, 3)

	// Verify top exercises have volume data
	for _, exercise := range result.TopExercises {
		assert.Greater(t, exercise.TotalVolume, 0.0)
		assert.Greater(t, exercise.TotalSets, 0)
		assert.Greater(t, exercise.MaxWeight, 0.0)
	}
}
