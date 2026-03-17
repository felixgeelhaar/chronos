-- Migration to align videos table with actual domain.Video model
-- This migration updates the videos table to match the Go domain model implementation

-- Add new columns that exist in domain model
ALTER TABLE videos ADD COLUMN IF NOT EXISTS url VARCHAR(500);
ALTER TABLE videos ADD COLUMN IF NOT EXISTS thumbnail_url VARCHAR(500);
ALTER TABLE videos ADD COLUMN IF NOT EXISTS set_id UUID;

-- Add foreign key constraint for set_id
ALTER TABLE videos ADD CONSTRAINT fk_videos_set
    FOREIGN KEY (set_id) REFERENCES sets(id) ON DELETE SET NULL;

-- Create index for set_id
CREATE INDEX IF NOT EXISTS idx_videos_set ON videos(set_id);

-- Rename duration_seconds to duration to match domain model
ALTER TABLE videos RENAME COLUMN duration_seconds TO duration;

-- Make file_size nullable to match domain model (*int64)
ALTER TABLE videos ALTER COLUMN file_size DROP NOT NULL;

-- Drop columns that don't exist in domain model
ALTER TABLE videos DROP COLUMN IF EXISTS s3_key;
ALTER TABLE videos DROP COLUMN IF EXISTS s3_bucket;
ALTER TABLE videos DROP COLUMN IF EXISTS file_name;
ALTER TABLE videos DROP COLUMN IF EXISTS content_type;
ALTER TABLE videos DROP COLUMN IF EXISTS thumbnail_s3_key;
ALTER TABLE videos DROP COLUMN IF EXISTS processing_status;
ALTER TABLE videos DROP COLUMN IF EXISTS processing_error;

-- Drop indexes that reference dropped columns
DROP INDEX IF EXISTS idx_videos_processing_status;
DROP INDEX IF EXISTS idx_videos_pending_processing;

-- Make url NOT NULL after adding it
-- Note: This migration assumes either no existing data or that url has been populated
ALTER TABLE videos ALTER COLUMN url SET NOT NULL;

-- Update comments to reflect actual implementation
COMMENT ON TABLE videos IS 'Video recordings of exercises with URL-based storage';
COMMENT ON COLUMN videos.url IS 'URL where the video is stored (e.g., CDN, S3 presigned URL)';
COMMENT ON COLUMN videos.thumbnail_url IS 'Optional URL for video thumbnail';
COMMENT ON COLUMN videos.set_id IS 'Optional specific set this video is linked to';
COMMENT ON COLUMN videos.session_id IS 'Optional session this video belongs to';
COMMENT ON COLUMN videos.duration IS 'Video duration in seconds';
COMMENT ON COLUMN videos.file_size IS 'Video file size in bytes';
COMMENT ON COLUMN videos.exercise_name IS 'Optional exercise name for categorization';
COMMENT ON COLUMN videos.date IS 'Date the video was recorded';
COMMENT ON CONSTRAINT fk_videos_set ON videos IS 'Set to NULL when set is deleted';
