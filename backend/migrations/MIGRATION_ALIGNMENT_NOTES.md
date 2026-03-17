# Database Migration Alignment Notes

## Overview
This document explains the migrations created to align the database schema with the actual Go domain models.

## Migration 000006: Align Videos Table with Domain Model

**Date**: 2025-10-26
**Purpose**: Fix critical mismatch between videos table schema and domain.Video model

### Changes Made

#### Columns Added:
- `url` VARCHAR(500) NOT NULL - Stores video URL (CDN, S3 presigned URL, etc.)
- `thumbnail_url` VARCHAR(500) - Optional thumbnail URL
- `set_id` UUID - Links video to specific set (with FK to sets table)

#### Columns Removed:
- `s3_key` - Replaced by generic `url` field
- `s3_bucket` - No longer needed with URL-based storage
- `file_name` - Not tracked in domain model
- `content_type` - Not tracked in domain model
- `thumbnail_s3_key` - Replaced by `thumbnail_url`
- `processing_status` - No video processing in current implementation
- `processing_error` - No video processing in current implementation

#### Columns Modified:
- `duration_seconds` → `duration` (renamed for consistency)
- `file_size` - Changed from NOT NULL to nullable (*int64 in Go)

#### Indexes Changed:
- Added: `idx_videos_set` for set_id lookups
- Removed: `idx_videos_processing_status`, `idx_videos_pending_processing`

### Rationale

The original migration assumed S3-based storage with server-side video processing. The actual implementation uses:
- Generic URL-based storage (flexible for any CDN/storage provider)
- No server-side video processing
- Direct linking to specific sets (not just sessions)

This migration brings the database schema in line with the implemented functionality.

### Data Migration Notes

**Important**: If you have existing data in the videos table:
1. You'll need to migrate `s3_key` → `url` (convert S3 keys to full URLs)
2. Ensure all videos have a `url` before running this migration
3. The migration will fail if `url` cannot be set to NOT NULL

**For new installations**: This migration can be run immediately as part of the normal migration sequence.

## Migration 000007: Add deleted_at to Sets Table

**Date**: 2025-10-26
**Purpose**: Enable GORM soft deletes for sets table

### Changes Made

#### Columns Added:
- `deleted_at` TIMESTAMP WITH TIME ZONE - Soft delete timestamp (NULL = active)

#### Indexes Added:
- `idx_sets_deleted_at` - Improves query performance for soft delete filtering

### Rationale

All other tables (users, sessions, one_rep_maxes, videos) have `deleted_at` columns for soft delete support. The sets table was missing this column, causing inconsistency in delete behavior.

GORM's soft delete feature requires a `DeletedAt` field (type `gorm.DeletedAt`) in the model and corresponding `deleted_at` column in the database.

### Impact

- Sets can now be soft-deleted (marked as deleted without physical removal)
- Maintains referential integrity for historical data
- Allows undelete functionality if needed in the future
- Consistent delete behavior across all entities

## Testing

All repository tests have been updated to account for these schema changes. Tests use SQLite in-memory databases with manually created schemas that match these migrations.

## Rollback

Both migrations include down files for safe rollback:
- `000006_align_videos_table_with_domain_model.down.sql`
- `000007_add_deleted_at_to_sets_table.down.sql`

**Warning**: Rolling back migration 000006 requires that data can satisfy the original NOT NULL constraints for `s3_key`, `s3_bucket`, `file_name`, etc.

## Next Steps

1. **Review** these migrations with the team
2. **Test** on development database before production
3. **Backup** production data before running migrations
4. **Monitor** application behavior after migration
5. **Update** API documentation if video handling has changed

## Questions or Issues

If you encounter issues with these migrations:
1. Check that your domain models match the expected schema
2. Verify foreign key relationships are intact
3. Ensure soft delete indexes are being used efficiently
4. Review query performance after adding new indexes
