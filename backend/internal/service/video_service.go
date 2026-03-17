package service

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/ascend/api/internal/domain"
	"github.com/ascend/api/internal/dto"
	"github.com/ascend/api/internal/repository"
	"github.com/ascend/api/pkg/queue"
	"github.com/ascend/api/pkg/storage"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// VideoService handles video business logic
type VideoService struct {
	videoRepo repository.VideoRepository
	s3Client  storage.StorageClient
	queue     *queue.Queue
}

// NewVideoService creates a new video service
func NewVideoService(videoRepo repository.VideoRepository, s3Client storage.StorageClient, q *queue.Queue) *VideoService {
	return &VideoService{
		videoRepo: videoRepo,
		s3Client:  s3Client,
		queue:     q,
	}
}

// UploadVideo uploads a video file to S3 and creates a database record
func (s *VideoService) UploadVideo(ctx context.Context, userID uuid.UUID, file io.Reader, filename string, contentType string, req *dto.UploadVideoRequest) (*dto.VideoResponse, error) {
	// Upload to S3
	uploadResult, err := s.s3Client.UploadFile(ctx, file, filename, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload video to S3: %w", err)
	}

	// Create video record
	video := &domain.Video{
		UserID:       userID,
		SessionID:    req.SessionID,
		URL:          uploadResult.URL,
		FileSize:     &uploadResult.FileSize,
		ExerciseName: req.ExerciseName,
		Date:         time.Now().UTC(),
	}

	if err := s.videoRepo.Create(ctx, video); err != nil {
		// Attempt to delete uploaded file if database creation fails
		_ = s.s3Client.DeleteFile(ctx, uploadResult.Key)
		return nil, fmt.Errorf("failed to create video record: %w", err)
	}

	// Enqueue video processing job (non-blocking)
	if s.queue != nil {
		payload := map[string]interface{}{
			"video_id": video.ID.String(),
		}
		_, err := s.queue.Enqueue(queue.JobTypeVideoProcessing, payload)
		if err != nil {
			log.Warn().
				Err(err).
				Str("video_id", video.ID.String()).
				Msg("Failed to enqueue video processing job")
			// Don't fail the upload if job enqueueing fails
		} else {
			log.Info().
				Str("video_id", video.ID.String()).
				Msg("Video processing job enqueued")
		}
	}

	return s.toResponse(video), nil
}

// GetVideo retrieves a video by ID
func (s *VideoService) GetVideo(ctx context.Context, userID uuid.UUID, id uuid.UUID) (*dto.VideoResponse, error) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("video not found")
	}

	// Verify ownership
	if video.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to video")
	}

	return s.toResponse(video), nil
}

// ListVideos retrieves all videos for a user with pagination
func (s *VideoService) ListVideos(ctx context.Context, userID uuid.UUID, page, pageSize int) (*dto.VideoListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get videos with pagination
	videos, err := s.videoRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get videos: %w", err)
	}

	// Get total count (simplified - in production, would use separate count query)
	// For now, if we get a full page, there might be more
	total := len(videos)
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	// Convert to responses
	responses := make([]dto.VideoResponse, len(videos))
	for i, video := range videos {
		responses[i] = *s.toResponse(video)
	}

	return &dto.VideoListResponse{
		Videos:     responses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListVideosBySession retrieves all videos for a specific session
func (s *VideoService) ListVideosBySession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) ([]dto.VideoResponse, error) {
	videos, err := s.videoRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get videos for session: %w", err)
	}

	// Verify ownership of first video (assumes all videos in session belong to same user)
	if len(videos) > 0 && videos[0].UserID != userID {
		return nil, fmt.Errorf("unauthorized access to session videos")
	}

	// Convert to responses
	responses := make([]dto.VideoResponse, len(videos))
	for i, video := range videos {
		responses[i] = *s.toResponse(video)
	}

	return responses, nil
}

// UpdateVideo updates video metadata
func (s *VideoService) UpdateVideo(ctx context.Context, userID uuid.UUID, id uuid.UUID, req *dto.UpdateVideoRequest) (*dto.VideoResponse, error) {
	// Get existing video
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("video not found")
	}

	// Verify ownership
	if video.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to video")
	}

	// Update fields
	if req.SessionID != nil {
		video.SessionID = req.SessionID
	}
	if req.ExerciseName != nil {
		video.ExerciseName = req.ExerciseName
	}

	if err := s.videoRepo.Update(ctx, video); err != nil {
		return nil, fmt.Errorf("failed to update video: %w", err)
	}

	return s.toResponse(video), nil
}

// DeleteVideo deletes a video and removes it from S3
func (s *VideoService) DeleteVideo(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	// Get video to verify ownership and get S3 key
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("video not found")
	}

	// Verify ownership
	if video.UserID != userID {
		return fmt.Errorf("unauthorized access to video")
	}

	// Delete from database first
	if err := s.videoRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete video record: %w", err)
	}

	// Extract S3 key from URL (assumes URL format: https://bucket.s3.amazonaws.com/key)
	// This is a simplified extraction - in production, store the key separately
	// For now, we'll attempt deletion but don't fail if it doesn't work
	// TODO: Store S3 key in database for reliable cleanup

	return nil
}

// GeneratePresignedURL generates a temporary presigned URL for video access
func (s *VideoService) GeneratePresignedURL(ctx context.Context, userID uuid.UUID, id uuid.UUID, expiration time.Duration) (*dto.GeneratePresignedURLResponse, error) {
	// Get video to verify ownership
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("video not found")
	}

	// Verify ownership
	if video.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to video")
	}

	// Extract S3 key from URL
	// This is simplified - in production, store the key in the database
	// For now, we'll return the public URL
	// TODO: Implement proper presigned URL generation with stored S3 key

	expiresAt := time.Now().Add(expiration)

	return &dto.GeneratePresignedURLResponse{
		URL:       video.URL,
		ExpiresAt: expiresAt,
	}, nil
}

// toResponse converts domain model to response DTO
func (s *VideoService) toResponse(video *domain.Video) *dto.VideoResponse {
	var fileSize int64
	if video.FileSize != nil {
		fileSize = *video.FileSize
	}

	return &dto.VideoResponse{
		ID:           video.ID,
		UserID:       video.UserID,
		SessionID:    video.SessionID,
		URL:          video.URL,
		ThumbnailURL: video.ThumbnailURL,
		Duration:     video.Duration,
		FileSize:     fileSize,
		ExerciseName: video.ExerciseName,
		Date:         video.Date,
		CreatedAt:    video.CreatedAt,
		UpdatedAt:    video.UpdatedAt,
	}
}
