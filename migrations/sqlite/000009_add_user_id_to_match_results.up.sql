-- Add user_id to match_results table for multi-tenancy
-- Note: We use DEFAULT 1 temporarily to handle existing data
ALTER TABLE match_results ADD COLUMN user_id INTEGER NOT NULL DEFAULT 1;

-- Add foreign key constraint (enforced with PRAGMA foreign_keys = ON)
-- SQLite doesn't support adding foreign keys to existing tables directly

-- Create index for faster user-based queries
CREATE INDEX idx_match_results_user_id ON match_results(user_id);

-- Update the existing index to include user_id for better performance
DROP INDEX IF EXISTS idx_match_results_created;
CREATE INDEX idx_match_results_user_created ON match_results(user_id, created_at DESC);