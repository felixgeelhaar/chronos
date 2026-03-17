-- Rollback users table creation
-- WARNING: This will delete all user data

DROP TABLE IF EXISTS users CASCADE;

-- Note: CASCADE will also drop any foreign key constraints referencing this table
-- Ensure dependent tables are migrated down first
