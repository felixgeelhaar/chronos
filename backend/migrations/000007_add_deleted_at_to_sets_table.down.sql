-- Remove deleted_at column from sets table

DROP INDEX IF EXISTS idx_sets_deleted_at;
ALTER TABLE sets DROP COLUMN IF EXISTS deleted_at;
