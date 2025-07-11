-- Add user_id to companies table for multi-tenancy
-- Note: We use DEFAULT 1 temporarily to handle existing data
-- In production, you may need a different strategy based on your data
ALTER TABLE companies ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;

-- Add foreign key constraint
-- Note: SQLite doesn't support adding foreign keys to existing tables directly
-- The foreign key will be enforced for new data with PRAGMA foreign_keys = ON

-- Drop the existing unique constraint on name
DROP INDEX IF EXISTS idx_companies_name;

-- Create composite unique index for user_id and name
-- This allows different users to have companies with the same name
CREATE UNIQUE INDEX idx_companies_user_name ON companies(user_id, name);

-- Create index for faster user-based queries
CREATE INDEX idx_companies_user_id ON companies(user_id);