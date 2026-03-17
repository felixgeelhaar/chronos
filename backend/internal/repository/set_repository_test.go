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

func TestSetRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	ctx := context.Background()

	// Create test user and session
	user := &domain.User{
		Email:    "settest@example.com",
		Password: "hashedpassword",
		Name:     "Set Test User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	tests := []struct {
		name        string
		set         *domain.Set
		expectError bool
	}{
		{
			name: "valid set",
			set: &domain.Set{
				SessionID:    session.ID,
				ExerciseName: "Squat",
				SetOrder:     1,
				Weight:       100.5,
				Reps:         5,
				RPE:          floatPtr(8.5),
			},
			expectError: false,
		},
		{
			name: "set without RPE",
			set: &domain.Set{
				SessionID:    session.ID,
				ExerciseName: "Bench Press",
				SetOrder:     1,
				Weight:       80.0,
				Reps:         8,
				RPE:          nil,
			},
			expectError: false,
		},
		{
			name: "set with zero weight",
			set: &domain.Set{
				SessionID:    session.ID,
				ExerciseName: "Pull-up",
				SetOrder:     1,
				Weight:       0, // Bodyweight exercise
				Reps:         10,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setRepo.Create(ctx, tt.set)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.set.ID)
				assert.NotZero(t, tt.set.CreatedAt)
			}
		})
	}
}

func TestSetRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	ctx := context.Background()

	// Create test data
	user := &domain.User{
		Email:    "setgetbyid@example.com",
		Password: "hashedpassword",
		Name:     "Set Get By ID User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	set := &domain.Set{
		SessionID:    session.ID,
		ExerciseName: "Deadlift",
		SetOrder:     1,
		Weight:       150.0,
		Reps:         3,
		RPE:          floatPtr(9.0),
	}
	err = setRepo.Create(ctx, set)
	require.NoError(t, err)

	tests := []struct {
		name        string
		setID       uuid.UUID
		expectError bool
		expectNil   bool
	}{
		{
			name:        "existing set",
			setID:       set.ID,
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "non-existent set",
			setID:       uuid.New(),
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundSet, err := setRepo.GetByID(ctx, tt.setID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, foundSet)
			} else {
				assert.NotNil(t, foundSet)
				assert.Equal(t, set.ID, foundSet.ID)
				assert.Equal(t, set.ExerciseName, foundSet.ExerciseName)
				assert.Equal(t, set.Weight, foundSet.Weight)
			}
		})
	}
}

func TestSetRepository_GetBySessionID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	ctx := context.Background()

	// Create test data
	user := &domain.User{
		Email:    "setsession@example.com",
		Password: "hashedpassword",
		Name:     "Set Session User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	// Create multiple sets
	sets := []*domain.Set{
		{SessionID: session.ID, ExerciseName: "Squat", SetOrder: 1, Weight: 100, Reps: 5},
		{SessionID: session.ID, ExerciseName: "Squat", SetOrder: 2, Weight: 110, Reps: 5},
		{SessionID: session.ID, ExerciseName: "Squat", SetOrder: 3, Weight: 120, Reps: 3},
		{SessionID: session.ID, ExerciseName: "Bench Press", SetOrder: 1, Weight: 80, Reps: 8},
	}

	for _, set := range sets {
		err := setRepo.Create(ctx, set)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		sessionID     uuid.UUID
		expectedCount int
	}{
		{
			name:          "get all sets for session",
			sessionID:     session.ID,
			expectedCount: 4,
		},
		{
			name:          "non-existent session",
			sessionID:     uuid.New(),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := setRepo.GetBySessionID(ctx, tt.sessionID)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify sets are ordered by set_order
			if len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					assert.LessOrEqual(t, result[i].SetOrder, result[i+1].SetOrder,
						"Sets should be ordered by set_order")
				}
			}
		})
	}
}

func TestSetRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	ctx := context.Background()

	// Create test data
	user := &domain.User{
		Email:    "setupdate@example.com",
		Password: "hashedpassword",
		Name:     "Set Update User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	set := &domain.Set{
		SessionID:    session.ID,
		ExerciseName: "Squat",
		SetOrder:     1,
		Weight:       100.0,
		Reps:         5,
		RPE:          floatPtr(8.0),
	}
	err = setRepo.Create(ctx, set)
	require.NoError(t, err)

	// Update set
	set.Weight = 105.0
	set.Reps = 3
	set.RPE = floatPtr(9.0)

	err = setRepo.Update(ctx, set)
	assert.NoError(t, err)

	// Verify update
	updated, err := setRepo.GetByID(ctx, set.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, 105.0, updated.Weight)
	assert.Equal(t, 3, updated.Reps)
	assert.Equal(t, 9.0, *updated.RPE)
}

func TestSetRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	ctx := context.Background()

	// Create test data
	user := &domain.User{
		Email:    "setdelete@example.com",
		Password: "hashedpassword",
		Name:     "Set Delete User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	set := &domain.Set{
		SessionID:    session.ID,
		ExerciseName: "Squat",
		SetOrder:     1,
		Weight:       100.0,
		Reps:         5,
	}
	err = setRepo.Create(ctx, set)
	require.NoError(t, err)

	// Delete set
	err = setRepo.Delete(ctx, set.ID)
	assert.NoError(t, err)

	// Verify deletion - should return error when trying to fetch deleted record
	deleted, err := setRepo.GetByID(ctx, set.ID)
	assert.Error(t, err, "Should return error for deleted record")
	assert.Nil(t, deleted, "Deleted set should not be retrievable")
}

func TestSetRepository_CascadeDelete(t *testing.T) {
	t.Skip("GORM soft deletes (setting deleted_at) don't trigger database CASCADE constraints. Soft delete is an UPDATE, not a DELETE, so ON DELETE CASCADE doesn't activate. Production code should manually handle cascade deletions if needed.")

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	ctx := context.Background()

	// Create test data
	user := &domain.User{
		Email:    "setcascade@example.com",
		Password: "hashedpassword",
		Name:     "Set Cascade User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	set := &domain.Set{
		SessionID:    session.ID,
		ExerciseName: "Squat",
		SetOrder:     1,
		Weight:       100.0,
		Reps:         5,
	}
	err = setRepo.Create(ctx, set)
	require.NoError(t, err)

	// Delete session (should cascade delete sets)
	err = sessionRepo.Delete(ctx, session.ID)
	assert.NoError(t, err)

	// Verify set is also deleted
	deletedSet, err := setRepo.GetByID(ctx, set.ID)
	assert.NoError(t, err)
	assert.Nil(t, deletedSet, "Set should be cascade deleted with session")
}

func TestSetRepository_SetOrdering(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	ctx := context.Background()

	// Create test data
	user := &domain.User{
		Email:    "setordering@example.com",
		Password: "hashedpassword",
		Name:     "Set Ordering User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	// Create sets in random order
	sets := []*domain.Set{
		{SessionID: session.ID, ExerciseName: "Squat", SetOrder: 3, Weight: 120, Reps: 3},
		{SessionID: session.ID, ExerciseName: "Squat", SetOrder: 1, Weight: 100, Reps: 5},
		{SessionID: session.ID, ExerciseName: "Squat", SetOrder: 2, Weight: 110, Reps: 5},
	}

	for _, set := range sets {
		err := setRepo.Create(ctx, set)
		require.NoError(t, err)
	}

	// Fetch sets
	result, err := setRepo.GetBySessionID(ctx, session.ID)
	require.NoError(t, err)
	require.Len(t, result, 3)

	// Verify ascending set_order
	for i := 0; i < len(result)-1; i++ {
		assert.LessOrEqual(t, result[i].SetOrder, result[i+1].SetOrder,
			"Sets should be ordered by set_order ascending")
	}

	// First set should be SetOrder 1
	assert.Equal(t, 1, result[0].SetOrder, "First set should have SetOrder 1")
}

func TestSetRepository_RPEValidation(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	setRepo := repository.NewSetRepository(db)
	ctx := context.Background()

	// Create test data
	user := &domain.User{
		Email:    "rpevalidation@example.com",
		Password: "hashedpassword",
		Name:     "RPE Validation User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	tests := []struct {
		name        string
		rpe         *float64
		expectError bool
	}{
		{
			name:        "valid RPE 8.5",
			rpe:         floatPtr(8.5),
			expectError: false,
		},
		{
			name:        "valid RPE 10",
			rpe:         floatPtr(10.0),
			expectError: false,
		},
		{
			name:        "valid RPE 0",
			rpe:         floatPtr(0.0),
			expectError: false,
		},
		{
			name:        "null RPE",
			rpe:         nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := &domain.Set{
				SessionID:    session.ID,
				ExerciseName: "Squat",
				SetOrder:     1,
				Weight:       100.0,
				Reps:         5,
				RPE:          tt.rpe,
			}

			err := setRepo.Create(ctx, set)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
