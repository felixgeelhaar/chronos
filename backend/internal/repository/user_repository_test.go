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

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		user        *domain.User
		expectError bool
		errorContains string
	}{
		{
			name: "valid user",
			user: &domain.User{
				Email:    "test@example.com",
				Password: "hashedpassword",
				Name:     "Test User",
			},
			expectError: false,
		},
		{
			name: "duplicate email",
			user: &domain.User{
				Email:    "duplicate@example.com",
				Password: "hashedpassword",
				Name:     "Duplicate User",
			},
			expectError: false, // First creation succeeds
		},
		{
			name: "user with body weight",
			user: &domain.User{
				Email:      "weighted@example.com",
				Password:   "hashedpassword",
				Name:       "Weighted User",
				BodyWeight: floatPtr(75.5),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.user)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.user.ID)
				assert.NotZero(t, tt.user.CreatedAt)
				assert.NotZero(t, tt.user.UpdatedAt)
			}
		})
	}

	// Test duplicate email (second attempt)
	t.Run("duplicate email - second attempt", func(t *testing.T) {
		duplicateUser := &domain.User{
			Email:    "duplicate@example.com",
			Password: "hashedpassword",
			Name:     "Another Duplicate",
		}
		err := repo.Create(ctx, duplicateUser)
		assert.Error(t, err, "Should fail on duplicate email")
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "getbyid@example.com",
		Password: "hashedpassword",
		Name:     "Get By ID User",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uuid.UUID
		expectError bool
		expectNil   bool
	}{
		{
			name:        "existing user",
			userID:      user.ID,
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "non-existent user",
			userID:      uuid.New(),
			expectError: true,
			expectNil:   true,
		},
		{
			name:        "zero UUID",
			userID:      uuid.UUID{},
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundUser, err := repo.GetByID(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, foundUser)
			} else {
				assert.NotNil(t, foundUser)
				assert.Equal(t, user.ID, foundUser.ID)
				assert.Equal(t, user.Email, foundUser.Email)
				assert.Equal(t, user.Name, foundUser.Name)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "getbyemail@example.com",
		Password: "hashedpassword",
		Name:     "Get By Email User",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name        string
		email       string
		expectError bool
		expectNil   bool
	}{
		{
			name:        "existing email",
			email:       "getbyemail@example.com",
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "non-existent email",
			email:       "nonexistent@example.com",
			expectError: true,
			expectNil:   true,
		},
		{
			name:        "empty email",
			email:       "",
			expectError: true,
			expectNil:   true,
		},
		{
			name:        "case sensitivity",
			email:       "GetByEmail@example.com", // Different case
			expectError: true,
			expectNil:   true, // Email lookup is case-sensitive in our implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundUser, err := repo.GetByEmail(ctx, tt.email)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, foundUser)
			} else {
				assert.NotNil(t, foundUser)
				assert.Equal(t, user.ID, foundUser.ID)
				assert.Equal(t, user.Email, foundUser.Email)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "update@example.com",
		Password: "hashedpassword",
		Name:     "Original Name",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	originalUpdatedAt := user.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

	tests := []struct {
		name           string
		updateUser     *domain.User
		expectError    bool
		validateFields func(*testing.T, *domain.User)
	}{
		{
			name: "update name",
			updateUser: &domain.User{
				ID:       user.ID,
				Email:    user.Email,
				Password: user.Password,
				Name:     "Updated Name",
			},
			expectError: false,
			validateFields: func(t *testing.T, updated *domain.User) {
				assert.Equal(t, "Updated Name", updated.Name)
				assert.True(t, updated.UpdatedAt.After(originalUpdatedAt))
			},
		},
		{
			name: "update body weight",
			updateUser: &domain.User{
				ID:         user.ID,
				Email:      user.Email,
				Password:   user.Password,
				Name:       "Updated Name",
				BodyWeight: floatPtr(80.5),
			},
			expectError: false,
			validateFields: func(t *testing.T, updated *domain.User) {
				assert.NotNil(t, updated.BodyWeight)
				assert.Equal(t, 80.5, *updated.BodyWeight)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.updateUser)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Fetch updated user
				updated, err := repo.GetByID(ctx, tt.updateUser.ID)
				require.NoError(t, err)
				require.NotNil(t, updated)

				if tt.validateFields != nil {
					tt.validateFields(t, updated)
				}
			}
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "delete@example.com",
		Password: "hashedpassword",
		Name:     "Delete User",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      uuid.UUID
		expectError bool
	}{
		{
			name:        "delete existing user",
			userID:      user.ID,
			expectError: false,
		},
		{
			name:        "delete non-existent user",
			userID:      uuid.New(),
			expectError: false, // Soft delete doesn't error if not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify user is soft-deleted - should return error for deleted user
				deletedUser, err := repo.GetByID(ctx, tt.userID)
				assert.Error(t, err, "Should return error for deleted user")
				assert.Nil(t, deletedUser, "Deleted user should not be retrievable")
			}
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Create multiple test users
	users := []*domain.User{
		{Email: "user1@example.com", Password: "hash", Name: "User 1"},
		{Email: "user2@example.com", Password: "hash", Name: "User 2"},
		{Email: "user3@example.com", Password: "hash", Name: "User 3"},
		{Email: "user4@example.com", Password: "hash", Name: "User 4"},
		{Email: "user5@example.com", Password: "hash", Name: "User 5"},
	}

	for _, user := range users {
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "list all users",
			limit:         10,
			offset:        0,
			expectedCount: 5,
		},
		{
			name:          "list with limit",
			limit:         3,
			offset:        0,
			expectedCount: 3,
		},
		{
			name:          "list with offset",
			limit:         10,
			offset:        2,
			expectedCount: 3,
		},
		{
			name:          "list with limit and offset",
			limit:         2,
			offset:        1,
			expectedCount: 2,
		},
		{
			name:          "offset beyond total",
			limit:         10,
			offset:        10,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.List(ctx, tt.limit, tt.offset)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify all returned users are valid
			for _, user := range result {
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.NotEmpty(t, user.Email)
				assert.NotEmpty(t, user.Name)
			}
		})
	}
}

func TestUserRepository_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "softdelete@example.com",
		Password: "hashedpassword",
		Name:     "Soft Delete User",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Soft delete the user
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	// User should not be found in normal queries - should return error
	foundUser, err := repo.GetByID(ctx, user.ID)
	assert.Error(t, err, "Should return error for soft-deleted user")
	assert.Nil(t, foundUser, "Soft-deleted user should not appear in normal queries")

	// User should not appear in list
	users, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	for _, u := range users {
		assert.NotEqual(t, user.ID, u.ID, "Soft-deleted user should not appear in list")
	}
}

func TestUserRepository_ConcurrentAccess(t *testing.T) {
	t.Skip("In-memory SQLite databases (':memory:') have concurrency limitations. Each goroutine connection may see a different database. Production code uses PostgreSQL which handles concurrency properly.")

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "concurrent@example.com",
		Password: "hashedpassword",
		Name:     "Concurrent User",
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Simulate concurrent reads
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			foundUser, err := repo.GetByID(ctx, user.ID)
			assert.NoError(t, err)
			assert.NotNil(t, foundUser)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
}
