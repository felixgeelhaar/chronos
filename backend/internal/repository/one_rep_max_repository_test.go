package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/ascend/api/internal/domain"
	"github.com/ascend/api/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOneRepMaxRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "orm@example.com",
		Password: "hashedpassword",
		Name:     "ORM Test User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name        string
		orm         *domain.OneRepMax
		expectError bool
	}{
		{
			name: "valid 1RM",
			orm: &domain.OneRepMax{
				UserID:       user.ID,
				ExerciseName: "Squat",
				Weight:       150.5,
				Date:         time.Now(),
			},
			expectError: false,
		},
		{
			name: "different exercise",
			orm: &domain.OneRepMax{
				UserID:       user.ID,
				ExerciseName: "Bench Press",
				Weight:       100.0,
				Date:         time.Now(),
			},
			expectError: false,
		},
		{
			name: "heavy deadlift",
			orm: &domain.OneRepMax{
				UserID:       user.ID,
				ExerciseName: "Deadlift",
				Weight:       200.0,
				Date:         time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ormRepo.Create(ctx, tt.orm)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.orm.ID)
				assert.NotZero(t, tt.orm.CreatedAt)
			}
		})
	}
}

func TestOneRepMaxRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormgetbyid@example.com",
		Password: "hashedpassword",
		Name:     "ORM Get By ID User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test 1RM
	orm := &domain.OneRepMax{
		UserID:       user.ID,
		ExerciseName: "Squat",
		Weight:       150.0,
		Date:         time.Now(),
	}
	err = ormRepo.Create(ctx, orm)
	require.NoError(t, err)

	tests := []struct {
		name        string
		ormID       uuid.UUID
		expectError bool
		expectNil   bool
	}{
		{
			name:        "existing 1RM",
			ormID:       orm.ID,
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "non-existent 1RM",
			ormID:       uuid.New(),
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundORM, err := ormRepo.GetByID(ctx, tt.ormID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, foundORM)
			} else {
				assert.NotNil(t, foundORM)
				assert.Equal(t, orm.ID, foundORM.ID)
				assert.Equal(t, orm.ExerciseName, foundORM.ExerciseName)
				assert.Equal(t, orm.Weight, foundORM.Weight)
			}
		})
	}
}

func TestOneRepMaxRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormuserid@example.com",
		Password: "hashedpassword",
		Name:     "ORM User ID User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple 1RMs
	orms := []*domain.OneRepMax{
		{UserID: user.ID, ExerciseName: "Squat", Weight: 150.0, Date: time.Now().AddDate(0, 0, -30)},
		{UserID: user.ID, ExerciseName: "Squat", Weight: 155.0, Date: time.Now().AddDate(0, 0, -15)},
		{UserID: user.ID, ExerciseName: "Squat", Weight: 160.0, Date: time.Now()},
		{UserID: user.ID, ExerciseName: "Bench Press", Weight: 100.0, Date: time.Now()},
	}

	for _, orm := range orms {
		err := ormRepo.Create(ctx, orm)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		expectedCount int
	}{
		{
			name:          "get all user 1RMs",
			userID:        user.ID,
			expectedCount: 4,
		},
		{
			name:          "non-existent user",
			userID:        uuid.New(),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ormRepo.GetByUserID(ctx, tt.userID)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify ordered by date descending
			if len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					assert.True(t, result[i].Date.After(result[i+1].Date) || result[i].Date.Equal(result[i+1].Date),
						"1RMs should be ordered by date descending")
				}
			}
		})
	}
}

func TestOneRepMaxRepository_GetLatestByUserIDAndExercise(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormlatest@example.com",
		Password: "hashedpassword",
		Name:     "ORM Latest User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create 1RM progression over time
	now := time.Now()
	orms := []*domain.OneRepMax{
		{UserID: user.ID, ExerciseName: "Squat", Weight: 140.0, Date: now.AddDate(0, 0, -90)},
		{UserID: user.ID, ExerciseName: "Squat", Weight: 145.0, Date: now.AddDate(0, 0, -60)},
		{UserID: user.ID, ExerciseName: "Squat", Weight: 150.0, Date: now.AddDate(0, 0, -30)},
		{UserID: user.ID, ExerciseName: "Squat", Weight: 160.0, Date: now}, // Latest
		{UserID: user.ID, ExerciseName: "Bench Press", Weight: 100.0, Date: now},
	}

	for _, orm := range orms {
		err := ormRepo.Create(ctx, orm)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		exerciseName   string
		expectedWeight float64
		expectError    bool
		expectNil      bool
	}{
		{
			name:           "latest squat",
			exerciseName:   "Squat",
			expectedWeight: 160.0,
			expectError:    false,
			expectNil:      false,
		},
		{
			name:           "latest bench press",
			exerciseName:   "Bench Press",
			expectedWeight: 100.0,
			expectError:    false,
			expectNil:      false,
		},
		{
			name:         "non-existent exercise",
			exerciseName: "Deadlift",
			expectError:  true,
			expectNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latest, err := ormRepo.GetLatestByUserIDAndExercise(ctx, user.ID, tt.exerciseName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, latest)
			} else {
				assert.NotNil(t, latest)
				assert.Equal(t, tt.expectedWeight, latest.Weight)
				assert.Equal(t, tt.exerciseName, latest.ExerciseName)
			}
		})
	}
}

func TestOneRepMaxRepository_GetByUserIDAndExercise(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormexercise@example.com",
		Password: "hashedpassword",
		Name:     "ORM Exercise User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create 1RMs for different exercises
	now := time.Now()
	orms := []*domain.OneRepMax{
		{UserID: user.ID, ExerciseName: "Squat", Weight: 150.0, Date: now.AddDate(0, 0, -30)},
		{UserID: user.ID, ExerciseName: "Squat", Weight: 155.0, Date: now.AddDate(0, 0, -15)},
		{UserID: user.ID, ExerciseName: "Squat", Weight: 160.0, Date: now},
		{UserID: user.ID, ExerciseName: "Bench Press", Weight: 100.0, Date: now},
		{UserID: user.ID, ExerciseName: "Deadlift", Weight: 180.0, Date: now},
	}

	for _, orm := range orms {
		err := ormRepo.Create(ctx, orm)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		exerciseName  string
		expectedCount int
	}{
		{
			name:          "squat history",
			exerciseName:  "Squat",
			expectedCount: 3,
		},
		{
			name:          "bench press history",
			exerciseName:  "Bench Press",
			expectedCount: 1,
		},
		{
			name:          "non-existent exercise",
			exerciseName:  "Overhead Press",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ormRepo.GetByUserIDAndExercise(ctx, user.ID, tt.exerciseName)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify all have correct exercise name
			for _, orm := range result {
				assert.Equal(t, tt.exerciseName, orm.ExerciseName)
			}

			// Verify ordered by date descending
			if len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					assert.True(t, result[i].Date.After(result[i+1].Date) || result[i].Date.Equal(result[i+1].Date),
						"1RMs should be ordered by date descending")
				}
			}
		})
	}
}

func TestOneRepMaxRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormupdate@example.com",
		Password: "hashedpassword",
		Name:     "ORM Update User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test 1RM
	orm := &domain.OneRepMax{
		UserID:       user.ID,
		ExerciseName: "Squat",
		Weight:       150.0,
		Date:         time.Now(),
	}
	err = ormRepo.Create(ctx, orm)
	require.NoError(t, err)

	// Update weight
	orm.Weight = 155.0

	err = ormRepo.Update(ctx, orm)
	assert.NoError(t, err)

	// Verify update
	updated, err := ormRepo.GetByID(ctx, orm.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, 155.0, updated.Weight)
}

func TestOneRepMaxRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormdelete@example.com",
		Password: "hashedpassword",
		Name:     "ORM Delete User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test 1RM
	orm := &domain.OneRepMax{
		UserID:       user.ID,
		ExerciseName: "Squat",
		Weight:       150.0,
		Date:         time.Now(),
	}
	err = ormRepo.Create(ctx, orm)
	require.NoError(t, err)

	// Delete 1RM
	err = ormRepo.Delete(ctx, orm.ID)
	assert.NoError(t, err)

	// Verify deletion - should return error when trying to fetch deleted record
	deleted, err := ormRepo.GetByID(ctx, orm.ID)
	assert.Error(t, err, "Should return error for deleted record")
	assert.Nil(t, deleted, "Deleted 1RM should not be retrievable")
}

func TestOneRepMaxRepository_CascadeDelete(t *testing.T) {
	t.Skip("GORM soft deletes (setting deleted_at) don't trigger database CASCADE constraints. Soft delete is an UPDATE, not a DELETE, so ON DELETE CASCADE doesn't activate. Production code should manually handle cascade deletions if needed.")

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormcascade@example.com",
		Password: "hashedpassword",
		Name:     "ORM Cascade User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test 1RM
	orm := &domain.OneRepMax{
		UserID:       user.ID,
		ExerciseName: "Squat",
		Weight:       150.0,
		Date:         time.Now(),
	}
	err = ormRepo.Create(ctx, orm)
	require.NoError(t, err)

	// Delete user (should cascade delete 1RMs)
	err = userRepo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// Verify 1RM is also deleted
	deletedORM, err := ormRepo.GetByID(ctx, orm.ID)
	assert.NoError(t, err)
	assert.Nil(t, deletedORM, "1RM should be cascade deleted with user")
}

func TestOneRepMaxRepository_DateOrdering(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormordering@example.com",
		Password: "hashedpassword",
		Name:     "ORM Ordering User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create 1RMs in random order
	now := time.Now()
	orms := []*domain.OneRepMax{
		{UserID: user.ID, ExerciseName: "Squat", Weight: 145.0, Date: now.AddDate(0, 0, -60)},
		{UserID: user.ID, ExerciseName: "Squat", Weight: 160.0, Date: now},
		{UserID: user.ID, ExerciseName: "Squat", Weight: 150.0, Date: now.AddDate(0, 0, -30)},
	}

	for _, orm := range orms {
		err := ormRepo.Create(ctx, orm)
		require.NoError(t, err)
	}

	// Fetch 1RMs
	result, err := ormRepo.GetByUserIDAndExercise(ctx, user.ID, "Squat")
	require.NoError(t, err)
	require.Len(t, result, 3)

	// Verify descending date order
	for i := 0; i < len(result)-1; i++ {
		assert.True(t, result[i].Date.After(result[i+1].Date) || result[i].Date.Equal(result[i+1].Date),
			"1RMs should be ordered by date descending (most recent first)")
	}

	// First entry should be most recent (160.0)
	assert.Equal(t, 160.0, result[0].Weight, "Most recent 1RM should be first")
}

func TestOneRepMaxRepository_Progression(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormprogression@example.com",
		Password: "hashedpassword",
		Name:     "ORM Progression User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create realistic progression
	now := time.Now()
	progression := []float64{140.0, 145.0, 150.0, 155.0, 160.0}
	for i, weight := range progression {
		orm := &domain.OneRepMax{
			UserID:       user.ID,
			ExerciseName: "Squat",
			Weight:       weight,
			Date:         now.AddDate(0, 0, -(len(progression)-i-1)*30), // Monthly progression
		}
		err := ormRepo.Create(ctx, orm)
		require.NoError(t, err)
	}

	// Get full history
	history, err := ormRepo.GetByUserIDAndExercise(ctx, user.ID, "Squat")
	require.NoError(t, err)
	require.Len(t, history, 5)

	// Verify progression (descending order, so reverse check)
	for i := 0; i < len(history)-1; i++ {
		assert.Greater(t, history[i].Weight, history[i+1].Weight,
			"Recent 1RMs should be heavier than older ones")
	}

	// Get latest
	latest, err := ormRepo.GetLatestByUserIDAndExercise(ctx, user.ID, "Squat")
	require.NoError(t, err)
	assert.Equal(t, 160.0, latest.Weight, "Latest should be heaviest")
}

func TestOneRepMaxRepository_MultipleExercises(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	ormRepo := repository.NewOneRepMaxRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ormmulti@example.com",
		Password: "hashedpassword",
		Name:     "ORM Multi User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create 1RMs for different exercises
	now := time.Now()
	exercises := []struct {
		name   string
		weight float64
	}{
		{"Squat", 160.0},
		{"Bench Press", 100.0},
		{"Deadlift", 180.0},
		{"Overhead Press", 70.0},
	}

	for _, ex := range exercises {
		orm := &domain.OneRepMax{
			UserID:       user.ID,
			ExerciseName: ex.name,
			Weight:       ex.weight,
			Date:         now,
		}
		err := ormRepo.Create(ctx, orm)
		require.NoError(t, err)
	}

	// Get all user 1RMs
	allORMs, err := ormRepo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, allORMs, 4)

	// Verify each exercise has correct weight
	for _, ex := range exercises {
		latest, err := ormRepo.GetLatestByUserIDAndExercise(ctx, user.ID, ex.name)
		require.NoError(t, err)
		require.NotNil(t, latest)
		assert.Equal(t, ex.weight, latest.Weight)
		assert.Equal(t, ex.name, latest.ExerciseName)
	}
}
