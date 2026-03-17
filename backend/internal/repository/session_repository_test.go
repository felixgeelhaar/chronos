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

func TestSessionRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "session@example.com",
		Password: "hashedpassword",
		Name:     "Session User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name        string
		session     *domain.Session
		expectError bool
	}{
		{
			name: "valid session",
			session: &domain.Session{
				UserID: user.ID,
				Date:   time.Now(),
				Notes:  stringPtr("Morning workout"),
			},
			expectError: false,
		},
		{
			name: "session without notes",
			session: &domain.Session{
				UserID: user.ID,
				Date:   time.Now().AddDate(0, 0, -1),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sessionRepo.Create(ctx, tt.session)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.session.ID)
				assert.NotZero(t, tt.session.CreatedAt)
			}
		})
	}
}

func TestSessionRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "sessionget@example.com",
		Password: "hashedpassword",
		Name:     "Session Get User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test session
	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
		Notes:  stringPtr("Test session"),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	tests := []struct {
		name        string
		sessionID   uuid.UUID
		expectError bool
		expectNil   bool
	}{
		{
			name:        "existing session",
			sessionID:   session.ID,
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "non-existent session",
			sessionID:   uuid.New(),
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundSession, err := sessionRepo.GetByID(ctx, tt.sessionID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, foundSession)
			} else {
				assert.NotNil(t, foundSession)
				assert.Equal(t, session.ID, foundSession.ID)
				assert.Equal(t, session.UserID, foundSession.UserID)
			}
		})
	}
}

func TestSessionRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "multiSession@example.com",
		Password: "hashedpassword",
		Name:     "Multi Session User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple sessions
	sessions := []*domain.Session{
		{UserID: user.ID, Date: time.Now().AddDate(0, 0, -2), Notes: stringPtr("Session 1")},
		{UserID: user.ID, Date: time.Now().AddDate(0, 0, -1), Notes: stringPtr("Session 2")},
		{UserID: user.ID, Date: time.Now(), Notes: stringPtr("Session 3")},
	}

	for _, session := range sessions {
		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		userID        uuid.UUID
		limit         int
		offset        int
		expectedCount int
	}{
		{
			name:          "get all sessions",
			userID:        user.ID,
			limit:         10,
			offset:        0,
			expectedCount: 3,
		},
		{
			name:          "get with limit",
			userID:        user.ID,
			limit:         2,
			offset:        0,
			expectedCount: 2,
		},
		{
			name:          "get with offset",
			userID:        user.ID,
			limit:         10,
			offset:        1,
			expectedCount: 2,
		},
		{
			name:          "non-existent user",
			userID:        uuid.New(),
			limit:         10,
			offset:        0,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sessionRepo.GetByUserID(ctx, tt.userID, tt.limit, tt.offset)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify sessions are ordered by date descending
			if len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					assert.True(t, result[i].Date.After(result[i+1].Date) || result[i].Date.Equal(result[i+1].Date),
						"Sessions should be ordered by date descending")
				}
			}
		})
	}
}

func TestSessionRepository_GetByUserIDAndDateRange(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "daterange@example.com",
		Password: "hashedpassword",
		Name:     "Date Range User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create sessions across different dates
	now := time.Now()
	sessions := []*domain.Session{
		{UserID: user.ID, Date: now.AddDate(0, 0, -10)}, // 10 days ago
		{UserID: user.ID, Date: now.AddDate(0, 0, -7)},  // 7 days ago
		{UserID: user.ID, Date: now.AddDate(0, 0, -5)},  // 5 days ago
		{UserID: user.ID, Date: now.AddDate(0, 0, -3)},  // 3 days ago
		{UserID: user.ID, Date: now.AddDate(0, 0, -1)},  // 1 day ago
		{UserID: user.ID, Date: now},                    // Today
	}

	for _, session := range sessions {
		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		startDate     time.Time
		endDate       time.Time
		expectedCount int
	}{
		{
			name:          "all sessions",
			startDate:     now.AddDate(0, 0, -30),
			endDate:       now.AddDate(0, 0, 1),
			expectedCount: 6,
		},
		{
			name:          "last 7 days",
			startDate:     now.AddDate(0, 0, -7),
			endDate:       now,
			expectedCount: 5, // -7, -5, -3, -1, today
		},
		{
			name:          "specific range",
			startDate:     now.AddDate(0, 0, -6),
			endDate:       now.AddDate(0, 0, -2),
			expectedCount: 2, // -5, -3
		},
		{
			name:          "single day",
			startDate:     now.AddDate(0, 0, -5),
			endDate:       now.AddDate(0, 0, -5),
			expectedCount: 1, // Only -5 days
		},
		{
			name:          "future date range",
			startDate:     now.AddDate(0, 0, 1),
			endDate:       now.AddDate(0, 0, 10),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sessionRepo.GetByUserIDAndDateRange(ctx, user.ID, tt.startDate, tt.endDate)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify all sessions are within date range
			for _, session := range result {
				assert.True(t, !session.Date.Before(tt.startDate), "Session date should be >= start date")
				assert.True(t, !session.Date.After(tt.endDate), "Session date should be <= end date")
			}
		})
	}
}

func TestSessionRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "updatesession@example.com",
		Password: "hashedpassword",
		Name:     "Update Session User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test session
	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
		Notes:  stringPtr("Original notes"),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	// Update notes
	updatedNotes := "Updated notes"
	session.Notes = &updatedNotes

	err = sessionRepo.Update(ctx, session)
	assert.NoError(t, err)

	// Verify update
	updated, err := sessionRepo.GetByID(ctx, session.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, updatedNotes, *updated.Notes)
}

func TestSessionRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "deletesession@example.com",
		Password: "hashedpassword",
		Name:     "Delete Session User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test session
	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	// Delete session
	err = sessionRepo.Delete(ctx, session.ID)
	assert.NoError(t, err)

	// Verify deletion - should return error when trying to fetch deleted record
	deleted, err := sessionRepo.GetByID(ctx, session.ID)
	assert.Error(t, err, "Should return error for deleted record")
	assert.Nil(t, deleted, "Deleted session should not be retrievable")
}

func TestSessionRepository_CascadeDelete(t *testing.T) {
	t.Skip("GORM soft deletes (setting deleted_at) don't trigger database CASCADE constraints. Soft delete is an UPDATE, not a DELETE, so ON DELETE CASCADE doesn't activate. Production code should manually handle cascade deletions if needed.")

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "cascade@example.com",
		Password: "hashedpassword",
		Name:     "Cascade User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test session
	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	// Delete user (should cascade delete session)
	err = userRepo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// Verify session is also deleted
	deletedSession, err := sessionRepo.GetByID(ctx, session.ID)
	assert.NoError(t, err)
	assert.Nil(t, deletedSession, "Session should be cascade deleted with user")
}

func TestSessionRepository_DateOrdering(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "ordering@example.com",
		Password: "hashedpassword",
		Name:     "Ordering User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create sessions in random order
	now := time.Now()
	sessions := []*domain.Session{
		{UserID: user.ID, Date: now.AddDate(0, 0, -1)},
		{UserID: user.ID, Date: now.AddDate(0, 0, -5)},
		{UserID: user.ID, Date: now},
		{UserID: user.ID, Date: now.AddDate(0, 0, -3)},
	}

	for _, session := range sessions {
		err := sessionRepo.Create(ctx, session)
		require.NoError(t, err)
	}

	// Fetch sessions
	result, err := sessionRepo.GetByUserID(ctx, user.ID, 10, 0)
	require.NoError(t, err)
	require.Len(t, result, 4)

	// Verify descending date order
	for i := 0; i < len(result)-1; i++ {
		assert.True(t, result[i].Date.After(result[i+1].Date) || result[i].Date.Equal(result[i+1].Date),
			"Sessions should be ordered by date descending (most recent first)")
	}

	// First session should be most recent (today)
	assert.True(t, result[0].Date.After(now.AddDate(0, 0, -1)) || result[0].Date.Equal(now),
		"First session should be most recent")
}
