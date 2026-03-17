-- Rollback migration to restore original videos table schema

-- Drop new columns added in up migration
DROP INDEX IF EXISTS idx_videos_set;
ALTER TABLE videos DROP CONSTRAINT IF EXISTS fk_videos_set;
ALTER TABLE videos DROP COLUMN IF EXISTS set_id;
ALTER TABLE videos DROP COLUMN IF EXISTS thumbnail_url;
ALTER TABLE videos DROP COLUMN IF EXISTS url;

-- Restore original columns
ALTER TABLE videos ADD COLUMN s3_key VARCHAR(500);
ALTER TABLE videos ADD COLUMN s3_bucket VARCHAR(255);
ALTER TABLE videos ADD COLUMN file_name VARCHAR(255);
ALTER TABLE videos ADD COLUMN content_type VARCHAR(100);
ALTER TABLE videos ADD COLUMN thumbnail_s3_key VARCHAR(500);
ALTER TABLE videos ADD COLUMN processing_status VARCHAR(50) DEFAULT 'pending';
ALTER TABLE videos ADD COLUMN processing_error TEXT;

-- Rename duration back to duration_seconds
ALTER TABLE videos RENAME COLUMN duration TO duration_seconds;

-- Make file_size NOT NULL again (original schema)
-- Note: This may fail if there are NULL values in file_size
ALTER TABLE videos ALTER COLUMN file_size SET NOT NULL;

-- Restore NOT NULL constraints on restored columns
-- Note: These may fail if existing data doesn't satisfy the constraints
ALTER TABLE videos ALTER COLUMN s3_key SET NOT NULL;
ALTER TABLE videos ALTER COLUMN s3_bucket SET NOT NULL;
ALTER TABLE videos ALTER COLUMN file_name SET NOT NULL;
ALTER TABLE videos ALTER COLUMN content_type SET NOT NULL;
ALTER TABLE videos ALTER COLUMN processing_status SET NOT NULL;

-- Recreate dropped indexes
CREATE INDEX IF NOT EXISTS idx_videos_processing_status ON videos(processing_status);
CREATE INDEX IF NOT EXISTS idx_videos_pending_processing
    ON videos(processing_status, created_at)
    WHERE processing_status IN ('pending', 'processing');

-- Add back validation constraint
ALTER TABLE videos ADD CONSTRAINT chk_videos_processing_status CHECK (
    processing_status IN ('pending', 'processing', 'completed', 'failed')
);

-- Restore original comments
COMMENT ON TABLE videos IS 'Video recordings of exercises stored in S3';
COMMENT ON COLUMN videos.s3_key IS 'S3 object key for the video file';
COMMENT ON COLUMN videos.processing_status IS 'Video processing status: pending, processing, completed, failed';
COMMENT ON COLUMN videos.duration_seconds IS 'Video duration in seconds (populated after processing)';
COMMENT ON COLUMN videos.thumbnail_s3_key IS 'S3 key for generated thumbnail';
