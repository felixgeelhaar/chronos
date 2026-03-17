-- Create one_rep_maxes table
-- Track personal records (1RM) for each exercise

CREATE TABLE IF NOT EXISTS one_rep_maxes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    exercise_name VARCHAR(255) NOT NULL,
    weight DECIMAL(6,2) NOT NULL,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Foreign key constraint
    CONSTRAINT fk_one_rep_maxes_user FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    -- Data validation constraints
    CONSTRAINT chk_one_rep_maxes_weight_positive CHECK (weight > 0)
);

-- Create indexes for one_rep_maxes table
CREATE INDEX IF NOT EXISTS idx_one_rep_maxes_user_id ON one_rep_maxes(user_id);
CREATE INDEX IF NOT EXISTS idx_one_rep_maxes_user_exercise ON one_rep_maxes(user_id, exercise_name, date DESC);
CREATE INDEX IF NOT EXISTS idx_one_rep_maxes_exercise ON one_rep_maxes(exercise_name, date DESC);
CREATE INDEX IF NOT EXISTS idx_one_rep_maxes_deleted_at ON one_rep_maxes(deleted_at);

-- Add comments
COMMENT ON TABLE one_rep_maxes IS 'One-rep max records for tracking personal records';
COMMENT ON COLUMN one_rep_maxes.user_id IS 'User who achieved this 1RM';
COMMENT ON COLUMN one_rep_maxes.exercise_name IS 'Exercise for which the 1RM was achieved';
COMMENT ON COLUMN one_rep_maxes.weight IS 'Maximum weight lifted for one repetition (in kg)';
COMMENT ON COLUMN one_rep_maxes.date IS 'Date when the 1RM was achieved';
COMMENT ON CONSTRAINT fk_one_rep_maxes_user ON one_rep_maxes IS 'Cascade delete when user is deleted';
