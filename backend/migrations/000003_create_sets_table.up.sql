-- Create sets table
-- Individual exercise sets within a training session

CREATE TABLE IF NOT EXISTS sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL,
    exercise_name VARCHAR(255) NOT NULL,
    set_order INTEGER NOT NULL DEFAULT 1,
    weight DECIMAL(6,2) NOT NULL DEFAULT 0,
    reps INTEGER NOT NULL DEFAULT 0,
    rpe DECIMAL(3,1), -- Rating of Perceived Exertion (0-10 scale)
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraint
    CONSTRAINT fk_sets_session FOREIGN KEY (session_id)
        REFERENCES sessions(id)
        ON DELETE CASCADE,

    -- Data validation constraints
    CONSTRAINT chk_sets_weight_positive CHECK (weight >= 0),
    CONSTRAINT chk_sets_reps_positive CHECK (reps >= 0),
    CONSTRAINT chk_sets_rpe_range CHECK (rpe IS NULL OR (rpe >= 0 AND rpe <= 10)),
    CONSTRAINT chk_sets_order_positive CHECK (set_order > 0)
);

-- Create indexes for sets table
CREATE INDEX IF NOT EXISTS idx_sets_session_id ON sets(session_id);
CREATE INDEX IF NOT EXISTS idx_sets_session_id_order ON sets(session_id, set_order);
CREATE INDEX IF NOT EXISTS idx_sets_exercise_name ON sets(exercise_name);
CREATE INDEX IF NOT EXISTS idx_sets_exercise_weight ON sets(exercise_name, weight DESC);

-- Add comments
COMMENT ON TABLE sets IS 'Individual exercise sets within training sessions';
COMMENT ON COLUMN sets.session_id IS 'Parent session this set belongs to';
COMMENT ON COLUMN sets.exercise_name IS 'Name of the exercise (e.g., "Bench Press", "Squat")';
COMMENT ON COLUMN sets.set_order IS 'Order of this set within the session (1st set, 2nd set, etc.)';
COMMENT ON COLUMN sets.weight IS 'Weight lifted in kg';
COMMENT ON COLUMN sets.reps IS 'Number of repetitions completed';
COMMENT ON COLUMN sets.rpe IS 'Rating of Perceived Exertion on 0-10 scale (optional)';
COMMENT ON CONSTRAINT fk_sets_session ON sets IS 'Cascade delete when session is deleted';
