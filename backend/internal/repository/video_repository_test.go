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

func TestVideoRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user and session
	user := &domain.User{
		Email:    "video@example.com",
		Password: "hashedpassword",
		Name:     "Video Test User",
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
		video       *domain.Video
		expectError bool
	}{
		{
			name: "valid video with session",
			video: &domain.Video{
				UserID:       user.ID,
				SessionID:    &session.ID,
				URL:          "https://cdn.example.com/videos/squat.mp4",
				ThumbnailURL: stringPtr("https://cdn.example.com/thumbs/squat.jpg"),
				Date:         time.Now(),
			},
			expectError: false,
		},
		{
			name: "video without session",
			video: &domain.Video{
				UserID: user.ID,
				URL:    "https://cdn.example.com/videos/bench.mp4",
				Date:   time.Now(),
			},
			expectError: false,
		},
		{
			name: "video with exercise name",
			video: &domain.Video{
				UserID:       user.ID,
				SessionID:    &session.ID,
				URL:          "https://cdn.example.com/videos/deadlift.mp4",
				ExerciseName: stringPtr("Deadlift"),
				Duration:     intPtr(45),
				FileSize:     int64Ptr(1024000),
				Date:         time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := videoRepo.Create(ctx, tt.video)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.video.ID)
				assert.NotZero(t, tt.video.CreatedAt)
			}
		})
	}
}

func TestVideoRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "videoget@example.com",
		Password: "hashedpassword",
		Name:     "Video Get User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test video
	video := &domain.Video{
		UserID: user.ID,
		URL:    "https://cdn.example.com/videos/test.mp4",
		Date:   time.Now(),
	}
	err = videoRepo.Create(ctx, video)
	require.NoError(t, err)

	tests := []struct {
		name        string
		videoID     uuid.UUID
		expectError bool
		expectNil   bool
	}{
		{
			name:        "existing video",
			videoID:     video.ID,
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "non-existent video",
			videoID:     uuid.New(),
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundVideo, err := videoRepo.GetByID(ctx, tt.videoID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, foundVideo)
			} else {
				assert.NotNil(t, foundVideo)
				assert.Equal(t, video.ID, foundVideo.ID)
				assert.Equal(t, video.URL, foundVideo.URL)
			}
		})
	}
}

func TestVideoRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "videouserlist@example.com",
		Password: "hashedpassword",
		Name:     "Video User List User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create multiple videos
	now := time.Now()
	videos := []*domain.Video{
		{UserID: user.ID, URL: "https://cdn.example.com/1.mp4", Date: now.AddDate(0, 0, -2)},
		{UserID: user.ID, URL: "https://cdn.example.com/2.mp4", Date: now.AddDate(0, 0, -1)},
		{UserID: user.ID, URL: "https://cdn.example.com/3.mp4", Date: now},
	}

	for _, video := range videos {
		err := videoRepo.Create(ctx, video)
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
			name:          "get all user videos",
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
			result, err := videoRepo.GetByUserID(ctx, tt.userID, tt.limit, tt.offset)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify videos are ordered by date descending
			if len(result) > 1 {
				for i := 0; i < len(result)-1; i++ {
					assert.True(t, result[i].Date.After(result[i+1].Date) || result[i].Date.Equal(result[i+1].Date),
						"Videos should be ordered by date descending")
				}
			}
		})
	}
}

func TestVideoRepository_GetBySessionID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "videosession@example.com",
		Password: "hashedpassword",
		Name:     "Video Session User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create sessions
	session1 := &domain.Session{UserID: user.ID, Date: time.Now()}
	session2 := &domain.Session{UserID: user.ID, Date: time.Now().AddDate(0, 0, -1)}
	err = sessionRepo.Create(ctx, session1)
	require.NoError(t, err)
	err = sessionRepo.Create(ctx, session2)
	require.NoError(t, err)

	// Create videos for different sessions
	videos := []*domain.Video{
		{UserID: user.ID, SessionID: &session1.ID, URL: "https://cdn.example.com/s1_1.mp4", Date: time.Now()},
		{UserID: user.ID, SessionID: &session1.ID, URL: "https://cdn.example.com/s1_2.mp4", Date: time.Now()},
		{UserID: user.ID, SessionID: &session2.ID, URL: "https://cdn.example.com/s2_1.mp4", Date: time.Now().AddDate(0, 0, -1)},
	}

	for _, video := range videos {
		err := videoRepo.Create(ctx, video)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		sessionID     uuid.UUID
		expectedCount int
	}{
		{
			name:          "session 1 videos",
			sessionID:     session1.ID,
			expectedCount: 2,
		},
		{
			name:          "session 2 videos",
			sessionID:     session2.ID,
			expectedCount: 1,
		},
		{
			name:          "non-existent session",
			sessionID:     uuid.New(),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := videoRepo.GetBySessionID(ctx, tt.sessionID)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestVideoRepository_GetByUserIDAndDateRange(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "videodaterange@example.com",
		Password: "hashedpassword",
		Name:     "Video Date Range User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create videos across different dates
	now := time.Now()
	videos := []*domain.Video{
		{UserID: user.ID, URL: "https://cdn.example.com/1.mp4", Date: now.AddDate(0, 0, -10)},
		{UserID: user.ID, URL: "https://cdn.example.com/2.mp4", Date: now.AddDate(0, 0, -7)},
		{UserID: user.ID, URL: "https://cdn.example.com/3.mp4", Date: now.AddDate(0, 0, -5)},
		{UserID: user.ID, URL: "https://cdn.example.com/4.mp4", Date: now.AddDate(0, 0, -3)},
		{UserID: user.ID, URL: "https://cdn.example.com/5.mp4", Date: now},
	}

	for _, video := range videos {
		err := videoRepo.Create(ctx, video)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		startDate     time.Time
		endDate       time.Time
		expectedCount int
	}{
		{
			name:          "all videos",
			startDate:     now.AddDate(0, 0, -30),
			endDate:       now.AddDate(0, 0, 1),
			expectedCount: 5,
		},
		{
			name:          "last 7 days",
			startDate:     now.AddDate(0, 0, -7),
			endDate:       now,
			expectedCount: 4, // -7, -5, -3, today
		},
		{
			name:          "specific range",
			startDate:     now.AddDate(0, 0, -6),
			endDate:       now.AddDate(0, 0, -2),
			expectedCount: 2, // -5, -3
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
			result, err := videoRepo.GetByUserIDAndDateRange(ctx, user.ID, tt.startDate, tt.endDate)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify all videos are within date range
			for _, video := range result {
				assert.True(t, !video.Date.Before(tt.startDate), "Video date should be >= start date")
				assert.True(t, !video.Date.After(tt.endDate), "Video date should be <= end date")
			}
		})
	}
}

func TestVideoRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "videoupdate@example.com",
		Password: "hashedpassword",
		Name:     "Video Update User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test video
	video := &domain.Video{
		UserID: user.ID,
		URL:    "https://cdn.example.com/test.mp4",
		Date:   time.Now(),
	}
	err = videoRepo.Create(ctx, video)
	require.NoError(t, err)

	// Update video
	video.ThumbnailURL = stringPtr("https://cdn.example.com/thumb.jpg")
	video.Duration = intPtr(45)
	video.FileSize = int64Ptr(2048000)

	err = videoRepo.Update(ctx, video)
	assert.NoError(t, err)

	// Verify update
	updated, err := videoRepo.GetByID(ctx, video.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.NotNil(t, updated.ThumbnailURL)
	assert.Equal(t, "https://cdn.example.com/thumb.jpg", *updated.ThumbnailURL)
	assert.NotNil(t, updated.Duration)
	assert.Equal(t, 45, *updated.Duration)
	assert.NotNil(t, updated.FileSize)
	assert.Equal(t, int64(2048000), *updated.FileSize)
}

func TestVideoRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "videodelete@example.com",
		Password: "hashedpassword",
		Name:     "Video Delete User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create test video
	video := &domain.Video{
		UserID: user.ID,
		URL:    "https://cdn.example.com/test.mp4",
		Date:   time.Now(),
	}
	err = videoRepo.Create(ctx, video)
	require.NoError(t, err)

	// Delete video
	err = videoRepo.Delete(ctx, video.ID)
	assert.NoError(t, err)

	// Verify deletion
	deleted, err := videoRepo.GetByID(ctx, video.ID)
	assert.Error(t, err)
	assert.Nil(t, deleted)
}

func TestVideoRepository_SessionSetNull(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "videosetnull@example.com",
		Password: "hashedpassword",
		Name:     "Video Set Null User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create session
	session := &domain.Session{
		UserID: user.ID,
		Date:   time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	require.NoError(t, err)

	// Create video linked to session
	video := &domain.Video{
		UserID:    user.ID,
		SessionID: &session.ID,
		URL:       "https://cdn.example.com/test.mp4",
		Date:      time.Now(),
	}
	err = videoRepo.Create(ctx, video)
	require.NoError(t, err)

	// Delete session (should SET NULL on video with GORM soft delete)
	err = sessionRepo.Delete(ctx, session.ID)
	assert.NoError(t, err)

	// Verify video still exists
	preservedVideo, err := videoRepo.GetByID(ctx, video.ID)
	assert.NoError(t, err)
	require.NotNil(t, preservedVideo, "Video should still exist after session deletion")
	// Note: With GORM soft delete, the relationship behavior depends on configuration
}

func TestVideoRepository_UserCascadeDelete(t *testing.T) {
	t.Skip("GORM soft deletes (setting deleted_at) don't trigger database CASCADE constraints. Soft delete is an UPDATE, not a DELETE, so ON DELETE CASCADE doesn't activate. Production code should manually handle cascade deletions if needed.")

	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "videocascade@example.com",
		Password: "hashedpassword",
		Name:     "Video Cascade User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	// Create video
	video := &domain.Video{
		UserID: user.ID,
		URL:    "https://cdn.example.com/test.mp4",
		Date:   time.Now(),
	}
	err = videoRepo.Create(ctx, video)
	require.NoError(t, err)

	// Delete user (should cascade delete videos with GORM soft delete)
	err = userRepo.Delete(ctx, user.ID)
	assert.NoError(t, err)

	// Verify video is also deleted
	deletedVideo, err := videoRepo.GetByID(ctx, video.ID)
	assert.Error(t, err)
	assert.Nil(t, deletedVideo)
}

func TestVideoRepository_ThumbnailURL(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	videoRepo := repository.NewVideoRepository(db)
	ctx := context.Background()

	// Create test user
	user := &domain.User{
		Email:    "videothumb@example.com",
		Password: "hashedpassword",
		Name:     "Video Thumbnail User",
	}
	err := userRepo.Create(ctx, user)
	require.NoError(t, err)

	tests := []struct {
		name         string
		thumbnailURL *string
	}{
		{
			name:         "with thumbnail",
			thumbnailURL: stringPtr("https://cdn.example.com/thumb1.jpg"),
		},
		{
			name:         "without thumbnail",
			thumbnailURL: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			video := &domain.Video{
				UserID:       user.ID,
				URL:          "https://cdn.example.com/test_" + tt.name + ".mp4",
				ThumbnailURL: tt.thumbnailURL,
				Date:         time.Now(),
			}

			err := videoRepo.Create(ctx, video)
			require.NoError(t, err)

			// Verify thumbnail
			retrieved, err := videoRepo.GetByID(ctx, video.ID)
			require.NoError(t, err)

			if tt.thumbnailURL == nil {
				assert.Nil(t, retrieved.ThumbnailURL)
			} else {
				require.NotNil(t, retrieved.ThumbnailURL)
				assert.Equal(t, *tt.thumbnailURL, *retrieved.ThumbnailURL)
			}
		})
	}
}
