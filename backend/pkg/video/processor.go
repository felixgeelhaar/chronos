package video

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProcessorConfig contains configuration for video processor
type ProcessorConfig struct {
	FFmpegPath      string
	FFprobePath     string
	TempDir         string
	ThumbnailWidth  int
	ThumbnailHeight int
}

// Processor handles video processing operations
type Processor struct {
	config ProcessorConfig
}

// NewProcessor creates a new video processor
func NewProcessor(config ProcessorConfig) *Processor {
	// Set defaults
	if config.FFmpegPath == "" {
		config.FFmpegPath = "ffmpeg"
	}
	if config.FFprobePath == "" {
		config.FFprobePath = "ffprobe"
	}
	if config.TempDir == "" {
		config.TempDir = os.TempDir()
	}
	if config.ThumbnailWidth == 0 {
		config.ThumbnailWidth = 320
	}
	if config.ThumbnailHeight == 0 {
		config.ThumbnailHeight = 180
	}

	return &Processor{
		config: config,
	}
}

// VideoMetadata contains extracted video metadata
type VideoMetadata struct {
	Duration int // Duration in seconds
	Width    int
	Height   int
	Codec    string
	Bitrate  int64
}

// ExtractMetadata extracts metadata from a video file
func (p *Processor) ExtractMetadata(ctx context.Context, videoPath string) (*VideoMetadata, error) {
	// Use ffprobe to extract metadata
	// ffprobe -v error -select_streams v:0 -show_entries stream=width,height,codec_name,bit_rate,duration -of csv=p=0 video.mp4
	cmd := exec.CommandContext(ctx, p.config.FFprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,codec_name,bit_rate,duration",
		"-of", "csv=p=0",
		videoPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w, stderr: %s", err, stderr.String())
	}

	// Parse output: width,height,codec,bitrate,duration
	output := strings.TrimSpace(stdout.String())
	parts := strings.Split(output, ",")
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected ffprobe output: %s", output)
	}

	metadata := &VideoMetadata{}

	// Parse width
	if len(parts) > 0 && parts[0] != "" {
		if width, err := strconv.Atoi(parts[0]); err == nil {
			metadata.Width = width
		}
	}

	// Parse height
	if len(parts) > 1 && parts[1] != "" {
		if height, err := strconv.Atoi(parts[1]); err == nil {
			metadata.Height = height
		}
	}

	// Parse codec
	if len(parts) > 2 && parts[2] != "" {
		metadata.Codec = parts[2]
	}

	// Parse bitrate
	if len(parts) > 3 && parts[3] != "" {
		if bitrate, err := strconv.ParseInt(parts[3], 10, 64); err == nil {
			metadata.Bitrate = bitrate
		}
	}

	// Parse duration
	if len(parts) > 4 && parts[4] != "" {
		if duration, err := strconv.ParseFloat(parts[4], 64); err == nil {
			metadata.Duration = int(duration)
		}
	}

	// If duration not in stream, get from format
	if metadata.Duration == 0 {
		duration, err := p.extractDurationFromFormat(ctx, videoPath)
		if err == nil {
			metadata.Duration = duration
		}
	}

	return metadata, nil
}

// extractDurationFromFormat extracts duration from video format
func (p *Processor) extractDurationFromFormat(ctx context.Context, videoPath string) (int, error) {
	cmd := exec.CommandContext(ctx, p.config.FFprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		videoPath,
	)

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return 0, err
	}

	output := strings.TrimSpace(stdout.String())
	duration, err := strconv.ParseFloat(output, 64)
	if err != nil {
		return 0, err
	}

	return int(duration), nil
}

// GenerateThumbnail generates a thumbnail from video at specified timestamp
func (p *Processor) GenerateThumbnail(ctx context.Context, videoPath string, timestamp time.Duration) (string, error) {
	// Generate unique filename for thumbnail
	thumbnailFilename := fmt.Sprintf("thumb_%s.jpg", uuid.New().String())
	thumbnailPath := filepath.Join(p.config.TempDir, thumbnailFilename)

	// Use ffmpeg to extract frame at timestamp
	// ffmpeg -i video.mp4 -ss 00:00:05 -vframes 1 -vf scale=320:180 thumbnail.jpg
	timestampStr := fmt.Sprintf("%02d:%02d:%02d",
		int(timestamp.Hours()),
		int(timestamp.Minutes())%60,
		int(timestamp.Seconds())%60,
	)

	scaleFilter := fmt.Sprintf("scale=%d:%d", p.config.ThumbnailWidth, p.config.ThumbnailHeight)

	cmd := exec.CommandContext(ctx, p.config.FFmpegPath,
		"-i", videoPath,
		"-ss", timestampStr,
		"-vframes", "1",
		"-vf", scaleFilter,
		"-y", // Overwrite output file
		thumbnailPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg thumbnail generation failed: %w, stderr: %s", err, stderr.String())
	}

	return thumbnailPath, nil
}

// GenerateThumbnailAtPercentage generates a thumbnail at a percentage of video duration
func (p *Processor) GenerateThumbnailAtPercentage(ctx context.Context, videoPath string, percentage float64) (string, error) {
	// Get video duration first
	metadata, err := p.ExtractMetadata(ctx, videoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get video metadata: %w", err)
	}

	if metadata.Duration == 0 {
		return "", fmt.Errorf("video duration is 0")
	}

	// Calculate timestamp
	timestamp := time.Duration(float64(metadata.Duration)*percentage/100.0) * time.Second

	return p.GenerateThumbnail(ctx, videoPath, timestamp)
}

// CleanupTempFile removes a temporary file
func (p *Processor) CleanupTempFile(filePath string) error {
	return os.Remove(filePath)
}
