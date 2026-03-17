# Database Migrations Guide

## Overview

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema management. Migrations provide version control for your database schema, allowing you to track changes over time and safely deploy schema updates.

## Why Migrations?

### Problems with GORM AutoMigrate (Previous Approach):
- ❌ No rollback capability
- ❌ No version control
- ❌ Can't track what changed when
- ❌ Limited control over indexes and constraints
- ❌ Doesn't handle column renames or type changes safely
- ❌ Not production-ready

### Benefits of golang-migrate:
- ✅ Version-controlled schema changes
- ✅ Rollback support (up and down migrations)
- ✅ Transaction safety
- ✅ Explicit control over DDL statements
- ✅ Production-grade migration management
- ✅ Multi-database support
- ✅ Works in CI/CD pipelines

## Migration File Structure

```
migrations/
├── 000001_create_users_table.up.sql       # Create users table
├── 000001_create_users_table.down.sql     # Rollback users table
├── 000002_create_sessions_table.up.sql    # Create sessions table
├── 000002_create_sessions_table.down.sql  # Rollback sessions table
├── 000003_create_sets_table.up.sql
├── 000003_create_sets_table.down.sql
├── 000004_create_one_rep_maxes_table.up.sql
├── 000004_create_one_rep_maxes_table.down.sql
├── 000005_create_videos_table.up.sql
└── 000005_create_videos_table.down.sql
```

### Naming Convention:
```
{version}_{description}.{up|down}.sql

Examples:
000001_create_users_table.up.sql
000006_add_email_verification.up.sql
000007_add_user_avatar_column.up.sql
```

- **Version**: Sequential number (6 digits, zero-padded)
- **Description**: Snake_case description of the change
- **Direction**: `up` for applying, `down` for rolling back

## Installation

### Install golang-migrate CLI:

```bash
# Automatically via Makefile (recommended)
make migrate-install

# Or manually
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Verify installation:

```bash
migrate -version
```

## Common Commands

### Apply Migrations

```bash
# Run all pending migrations
make migrate-up

# Or with explicit DATABASE_URL
DATABASE_URL="postgresql://user:pass@localhost:5432/dbname?sslmode=disable" make migrate-up
```

### Check Migration Status

```bash
# Show current migration version
make migrate-status

# Output: 5 (if all 5 migrations are applied)
```

### Rollback Migrations

```bash
# Rollback last migration
make migrate-down

# Rollback all migrations (WARNING: destructive)
make migrate-down-all
```

### Create New Migration

```bash
# Create new migration files
make migrate-create name=add_user_avatar

# Creates:
# migrations/000006_add_user_avatar.up.sql
# migrations/000006_add_user_avatar.down.sql
```

### Validate Migrations

```bash
# Check migration files for basic issues
make migrate-validate
```

### Force Migration Version

If migrations get out of sync (dirty state), force a specific version:

```bash
# Force version to 5
make migrate-force version=5
```

## Writing Migrations

### Good Migration Practices

#### 1. **Always Write Down Migrations**

Every `up` migration must have a corresponding `down` migration:

```sql
-- 000006_add_user_avatar.up.sql
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(500);
CREATE INDEX idx_users_avatar ON users(avatar_url);

-- 000006_add_user_avatar.down.sql
DROP INDEX IF EXISTS idx_users_avatar;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
```

#### 2. **Use Idempotent Operations**

Always use `IF EXISTS` / `IF NOT EXISTS`:

```sql
-- Good ✅
CREATE TABLE IF NOT EXISTS users (...);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);

-- Bad ❌
CREATE TABLE users (...);  -- Fails if table exists
CREATE INDEX idx_users_email ON users(email);  -- Fails if index exists
```

#### 3. **Add Comments**

Document your schema:

```sql
COMMENT ON TABLE users IS 'Core user accounts with authentication credentials';
COMMENT ON COLUMN users.email IS 'Unique email address for authentication';
COMMENT ON CONSTRAINT fk_sessions_user ON sessions IS 'Cascade delete when user is deleted';
```

#### 4. **Use Constraints**

Enforce data integrity at the database level:

```sql
-- Data validation constraints
CONSTRAINT chk_weight_positive CHECK (weight >= 0),
CONSTRAINT chk_rpe_range CHECK (rpe IS NULL OR (rpe >= 0 AND rpe <= 10)),

-- Foreign key constraints
CONSTRAINT fk_sets_session FOREIGN KEY (session_id)
    REFERENCES sessions(id)
    ON DELETE CASCADE
```

#### 5. **Create Indexes Strategically**

```sql
-- Single-column index for lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_sessions_user_date ON sessions(user_id, date DESC);

-- Partial index for filtering
CREATE INDEX IF NOT EXISTS idx_videos_pending
    ON videos(created_at)
    WHERE processing_status = 'pending';
```

### Migration Templates

#### Adding a Column:

```sql
-- up
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone_number);

-- down
DROP INDEX IF EXISTS idx_users_phone;
ALTER TABLE users DROP COLUMN IF EXISTS phone_number;
```

#### Adding a Table:

```sql
-- up
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    message TEXT NOT NULL,
    read BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_notifications_user FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, read, created_at DESC);

-- down
DROP TABLE IF EXISTS notifications CASCADE;
```

#### Modifying a Column:

```sql
-- up: Change column type
ALTER TABLE users ALTER COLUMN name TYPE VARCHAR(500);

-- down: Revert to original type
ALTER TABLE users ALTER COLUMN name TYPE VARCHAR(255);
```

#### Adding a Foreign Key:

```sql
-- up
ALTER TABLE videos
    ADD CONSTRAINT fk_videos_exercise
    FOREIGN KEY (exercise_id)
    REFERENCES exercises(id)
    ON DELETE SET NULL;

-- down
ALTER TABLE videos DROP CONSTRAINT IF EXISTS fk_videos_exercise;
```

## Migration Workflow

### Development Workflow

1. **Create migration**:
   ```bash
   make migrate-create name=add_feature
   ```

2. **Write SQL** in both `.up.sql` and `.down.sql` files

3. **Test migration locally**:
   ```bash
   make migrate-up
   ```

4. **Verify database schema**:
   ```bash
   psql $DATABASE_URL -c "\dt"  # List tables
   psql $DATABASE_URL -c "\d users"  # Describe users table
   ```

5. **Test rollback**:
   ```bash
   make migrate-down  # Test down migration
   make migrate-up    # Apply again
   ```

6. **Commit migration files**:
   ```bash
   git add migrations/
   git commit -m "feat: add user avatar column"
   ```

### Production Deployment Workflow

1. **Backup database** (always!):
   ```bash
   pg_dump $DATABASE_URL > backup_$(date +%Y%m%d).sql
   ```

2. **Run migrations** (automated in CI/CD):
   ```bash
   make migrate-up
   ```

3. **Verify schema version**:
   ```bash
   make migrate-status
   ```

4. **If issues occur**, rollback:
   ```bash
   make migrate-down
   ```

5. **Monitor application logs** for database errors

## Troubleshooting

### Dirty Migration State

**Problem**: Migration failed mid-execution, database is in "dirty" state.

**Solution**:
```bash
# Check current version
make migrate-status

# Output: 3/d (dirty)

# Fix manually in database, then force clean state
make migrate-force version=3
```

### Migration Out of Sync

**Problem**: Local migrations don't match production.

**Solution**:
```bash
# Check production version
make migrate-status

# If behind, apply missing migrations
make migrate-up

# If ahead, DO NOT ROLLBACK PRODUCTION
# Instead, apply your local migrations to production
```

### Migration Failed to Apply

**Problem**: SQL error during migration.

**Solution**:
1. Check error message carefully
2. Fix SQL in migration file
3. Rollback to previous version: `make migrate-down`
4. Apply fixed migration: `make migrate-up`

### Can't Rollback Migration

**Problem**: Down migration fails.

**Solution**:
1. Manually fix database schema
2. Force to correct version: `make migrate-force version=X`
3. Update down migration for future use

## Database Schema Versioning

### Current Schema (Version 5):

| Table | Purpose | Foreign Keys |
|-------|---------|-------------|
| `users` | User accounts | - |
| `sessions` | Training sessions | → users |
| `sets` | Exercise sets | → sessions |
| `one_rep_maxes` | Personal records | → users |
| `videos` | Exercise videos | → users, sessions |

### Migration History:

| Version | Description | Applied |
|---------|-------------|---------|
| 000001 | Create users table | ✅ |
| 000002 | Create sessions table | ✅ |
| 000003 | Create sets table | ✅ |
| 000004 | Create one_rep_maxes table | ✅ |
| 000005 | Create videos table | ✅ |

## CI/CD Integration

### GitHub Actions Example:

```yaml
name: Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install golang-migrate
        run: |
          go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

      - name: Run migrations
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
        run: |
          migrate -path migrations -database "$DATABASE_URL" up
```

### Docker Deployment:

```dockerfile
# In Dockerfile
FROM golang:1.21 as builder
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# In entrypoint script
#!/bin/sh
migrate -path /app/migrations -database "$DATABASE_URL" up
exec /app/ascend-api
```

## Best Practices

### DO:
✅ Always write both up and down migrations
✅ Test migrations locally before deploying
✅ Backup production database before migrating
✅ Use idempotent operations (IF EXISTS, IF NOT EXISTS)
✅ Add comments to complex schema changes
✅ Create indexes for common queries
✅ Use constraints for data integrity
✅ Version control all migration files

### DON'T:
❌ Modify existing migration files (create new ones instead)
❌ Skip writing down migrations
❌ Run migrations without backups
❌ Use AutoMigrate in production
❌ Commit dirty migration state
❌ Force version without understanding why
❌ Delete old migration files (break history)

## Advanced Topics

### Zero-Downtime Migrations

For production systems requiring zero downtime:

1. **Add new column** (nullable):
   ```sql
   ALTER TABLE users ADD COLUMN new_field VARCHAR(255);
   ```

2. **Deploy code** that writes to both old and new fields

3. **Backfill data**:
   ```sql
   UPDATE users SET new_field = old_field WHERE new_field IS NULL;
   ```

4. **Deploy code** that only uses new field

5. **Remove old column**:
   ```sql
   ALTER TABLE users DROP COLUMN old_field;
   ```

### Large Data Migrations

For migrations affecting millions of rows:

```sql
-- Process in batches
DO $$
DECLARE
    batch_size INT := 1000;
    processed INT := 0;
BEGIN
    LOOP
        UPDATE users
        SET normalized_email = LOWER(email)
        WHERE id IN (
            SELECT id FROM users
            WHERE normalized_email IS NULL
            LIMIT batch_size
        );

        processed := processed + batch_size;

        EXIT WHEN NOT FOUND;

        -- Log progress
        RAISE NOTICE 'Processed % rows', processed;

        -- Commit and sleep to avoid blocking
        COMMIT;
        PERFORM pg_sleep(0.1);
    END LOOP;
END $$;
```

## Resources

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Database Migration Best Practices](https://martinfowler.com/articles/evodb.html)
- [Zero-Downtime Deployments](https://blog.codeship.com/zero-downtime-deployments/)

## Quick Reference

```bash
# Installation
make migrate-install

# Create migration
make migrate-create name=your_migration_name

# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check status
make migrate-status

# Force version (emergency)
make migrate-force version=5

# Validate migrations
make migrate-validate
```

---

**Last Updated**: Sprint 1, Week 2
**Current Schema Version**: 5
**Migration System**: golang-migrate v4
