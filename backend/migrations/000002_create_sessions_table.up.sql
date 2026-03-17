-- Create sessions table
-- Training sessions are the core workout tracking entity

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Foreign key constraint
    CONSTRAINT fk_sessions_user FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Create indexes for sessions table
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id_date ON sessions(user_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_date ON sessions(date DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_deleted_at ON sessions(deleted_at);

-- Create unique constraint to prevent duplicate sessions per user per day
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_user_date_unique
    ON sessions(user_id, date)
    WHERE deleted_at IS NULL;

-- Add comments
COMMENT ON TABLE sessions IS 'Training sessions containing sets of exercises';
COMMENT ON COLUMN sessions.user_id IS 'Owner of the session';
COMMENT ON COLUMN sessions.date IS 'Date when the workout was performed';
COMMENT ON COLUMN sessions.notes IS 'Optional workout notes or comments';
COMMENT ON CONSTRAINT fk_sessions_user ON sessions IS 'Cascade delete when user is deleted';
