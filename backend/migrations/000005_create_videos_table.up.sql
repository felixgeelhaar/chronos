-- Create videos table
-- Video recordings of exercises for form checking and progress tracking

CREATE TABLE IF NOT EXISTS videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    session_id UUID, -- Optional: video may not be linked to a session
    exercise_name VARCHAR(255),
    date DATE NOT NULL DEFAULT CURRENT_DATE,

    -- S3 storage information
    s3_key VARCHAR(500) NOT NULL, -- S3 object key
    s3_bucket VARCHAR(255) NOT NULL,

    -- Video metadata
    file_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    duration_seconds INTEGER,
    thumbnail_s3_key VARCHAR(500),

    -- Processing status
    processing_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    processing_error TEXT,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Foreign key constraints
    CONSTRAINT fk_videos_user FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_videos_session FOREIGN KEY (session_id)
        REFERENCES sessions(id)
        ON DELETE SET NULL, -- Preserve video even if session is deleted

    -- Data validation constraints
    CONSTRAINT chk_videos_file_size_positive CHECK (file_size > 0),
    CONSTRAINT chk_videos_duration_positive CHECK (duration_seconds IS NULL OR duration_seconds > 0),
    CONSTRAINT chk_videos_processing_status CHECK (
        processing_status IN ('pending', 'processing', 'completed', 'failed')
    )
);

-- Create indexes for videos table
CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos(user_id);
CREATE INDEX IF NOT EXISTS idx_videos_user_id_date ON videos(user_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_videos_session_id ON videos(session_id);
CREATE INDEX IF NOT EXISTS idx_videos_exercise ON videos(exercise_name, date DESC);
CREATE INDEX IF NOT EXISTS idx_videos_processing_status ON videos(processing_status);
CREATE INDEX IF NOT EXISTS idx_videos_deleted_at ON videos(deleted_at);

-- Create index for finding videos needing processing
CREATE INDEX IF NOT EXISTS idx_videos_pending_processing
    ON videos(processing_status, created_at)
    WHERE processing_status IN ('pending', 'processing');

-- Add comments
COMMENT ON TABLE videos IS 'Video recordings of exercises stored in S3';
COMMENT ON COLUMN videos.user_id IS 'User who uploaded the video';
COMMENT ON COLUMN videos.session_id IS 'Optional session this video belongs to';
COMMENT ON COLUMN videos.s3_key IS 'S3 object key for the video file';
COMMENT ON COLUMN videos.processing_status IS 'Video processing status: pending, processing, completed, failed';
COMMENT ON COLUMN videos.duration_seconds IS 'Video duration in seconds (populated after processing)';
COMMENT ON COLUMN videos.thumbnail_s3_key IS 'S3 key for generated thumbnail';
COMMENT ON CONSTRAINT fk_videos_user ON videos IS 'Cascade delete when user is deleted';
COMMENT ON CONSTRAINT fk_videos_session ON videos IS 'Set to NULL when session is deleted (preserve video)';
