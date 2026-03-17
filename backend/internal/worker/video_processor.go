package worker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ascend/api/internal/repository"
	"github.com/ascend/api/pkg/queue"
	"github.com/ascend/api/pkg/storage"
	"github.com/ascend/api/pkg/video"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// VideoProcessorWorker handles video processing jobs
type VideoProcessorWorker struct {
	videoRepo      repository.VideoRepository
	s3Client       *storage.S3Client
	videoProcessor *video.Processor
	tempDir        string
}

// NewVideoProcessorWorker creates a new video processor worker
func NewVideoProcessorWorker(
	videoRepo repository.VideoRepository,
	s3Client *storage.S3Client,
	videoProcessor *video.Processor,
) *VideoProcessorWorker {
	return &VideoProcessorWorker{
		videoRepo:      videoRepo,
		s3Client:       s3Client,
		videoProcessor: videoProcessor,
		tempDir:        os.TempDir(),
	}
}

// ProcessVideo is the handler for video processing jobs
func (w *VideoProcessorWorker) ProcessVideo(ctx context.Context, job *queue.Job) error {
	// Extract video ID from job payload
	videoIDStr, ok := job.Payload["video_id"].(string)
	if !ok {
		return fmt.Errorf("video_id not found in job payload")
	}

	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		return fmt.Errorf("invalid video_id: %w", err)
	}

	log.Info().
		Str("job_id", job.ID.String()).
		Str("video_id", videoID.String()).
		Msg("Processing video")

	// Get video from database
	video, err := w.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		return fmt.Errorf("failed to get video: %w", err)
	}

	// Download video from S3 to temp file
	tempVideoPath, err := w.downloadVideo(ctx, video.URL)
	if err != nil {
		return fmt.Errorf("failed to download video: %w", err)
	}
	defer os.Remove(tempVideoPath)

	// Extract metadata
	metadata, err := w.videoProcessor.ExtractMetadata(ctx, tempVideoPath)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to extract metadata, continuing without it")
		// Don't fail the job if metadata extraction fails
	}

	// Update video with metadata
	if metadata != nil {
		video.Duration = &metadata.Duration
		log.Info().
			Str("video_id", videoID.String()).
			Int("duration", metadata.Duration).
			Int("width", metadata.Width).
			Int("height", metadata.Height).
			Msg("Extracted video metadata")
	}

	// Generate thumbnail at 10% of video duration
	thumbnailPath, err := w.videoProcessor.GenerateThumbnailAtPercentage(ctx, tempVideoPath, 10.0)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to generate thumbnail, continuing without it")
		// Don't fail the job if thumbnail generation fails
	} else {
		defer os.Remove(thumbnailPath)

		// Upload thumbnail to S3
		thumbnailURL, err := w.uploadThumbnail(ctx, thumbnailPath, videoID)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to upload thumbnail, continuing without it")
		} else {
			video.ThumbnailURL = &thumbnailURL
			log.Info().
				Str("video_id", videoID.String()).
				Str("thumbnail_url", thumbnailURL).
				Msg("Generated and uploaded thumbnail")
		}
	}

	// Update video record in database
	if err := w.videoRepo.Update(ctx, video); err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	log.Info().
		Str("job_id", job.ID.String()).
		Str("video_id", videoID.String()).
		Msg("Video processing completed")

	return nil
}

// downloadVideo downloads a video from URL to a temporary file
func (w *VideoProcessorWorker) downloadVideo(ctx context.Context, videoURL string) (string, error) {
	// Create temp file
	tempFile, err := os.CreateTemp(w.tempDir, "video_*.mp4")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	tempPath := tempFile.Name()

	// Download video
	req, err := http.NewRequestWithContext(ctx, "GET", videoURL, nil)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to download video: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Remove(tempPath)
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Write to temp file
	if _, err := io.Copy(tempFile, resp.Body); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to write video to temp file: %w", err)
	}

	return tempPath, nil
}

// uploadThumbnail uploads a thumbnail to S3
func (w *VideoProcessorWorker) uploadThumbnail(ctx context.Context, thumbnailPath string, videoID uuid.UUID) (string, error) {
	// Open thumbnail file
	file, err := os.Open(thumbnailPath)
	if err != nil {
		return "", fmt.Errorf("failed to open thumbnail: %w", err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat thumbnail: %w", err)
	}

	// Upload to S3 with a key based on video ID
	filename := fmt.Sprintf("thumbnail_%s%s", videoID.String(), filepath.Ext(thumbnailPath))
	result, err := w.s3Client.UploadFile(ctx, file, filename, "image/jpeg")
	if err != nil {
		return "", fmt.Errorf("failed to upload thumbnail to S3: %w", err)
	}

	log.Info().
		Str("thumbnail_url", result.URL).
		Int64("size", fileInfo.Size()).
		Msg("Thumbnail uploaded to S3")

	return result.URL, nil
}

// RegisterWithQueue registers this worker with a job queue
func (w *VideoProcessorWorker) RegisterWithQueue(q *queue.Queue) {
	q.RegisterHandler(queue.JobTypeVideoProcessing, w.ProcessVideo)
}
