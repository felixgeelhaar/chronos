-- Add deleted_at column to sets table for GORM soft delete support

ALTER TABLE sets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Create index for soft delete queries
CREATE INDEX IF NOT EXISTS idx_sets_deleted_at ON sets(deleted_at);

-- Add comment
COMMENT ON COLUMN sets.deleted_at IS 'Soft delete timestamp - NULL means active set';
