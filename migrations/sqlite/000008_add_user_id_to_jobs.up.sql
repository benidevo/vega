-- Add user_id to jobs table for multi-tenancy
-- Note: We use DEFAULT 1 temporarily to handle existing data
ALTER TABLE jobs ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;

-- Add foreign key constraint (enforced with PRAGMA foreign_keys = ON)
-- SQLite doesn't support adding foreign keys to existing tables directly

-- Create index for faster user-based queries
CREATE INDEX idx_jobs_user_id ON jobs(user_id);

-- Drop the unique constraint on source_url to make it unique per user
DROP INDEX IF EXISTS idx_jobs_source_url;

-- Create composite unique index for user_id and source_url
-- This allows different users to save the same job posting
CREATE UNIQUE INDEX idx_jobs_user_source_url ON jobs(user_id, source_url);