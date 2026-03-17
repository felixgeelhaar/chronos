-- Rollback videos table creation
-- WARNING: This will delete all video metadata (but not S3 files - manual cleanup required)

DROP TABLE IF EXISTS videos CASCADE;

-- NOTE: S3 objects must be deleted separately
-- Use aws s3 rm or configure S3 lifecycle policies
