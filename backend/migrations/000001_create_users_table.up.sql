-- Create users table
-- This is the core user entity for authentication and profile management

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    body_weight DECIMAL(5,2),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at); -- For soft delete queries

-- Add comment to table
COMMENT ON TABLE users IS 'Core user accounts with authentication credentials and profile information';
COMMENT ON COLUMN users.id IS 'Primary key - UUID generated automatically';
COMMENT ON COLUMN users.email IS 'Unique email address for authentication';
COMMENT ON COLUMN users.password IS 'Bcrypt hashed password (cost factor 12)';
COMMENT ON COLUMN users.body_weight IS 'Optional body weight in kg for analytics';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp - NULL means active user';
