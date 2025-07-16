-- Drop quota tracking index
DROP INDEX IF EXISTS idx_user_quota_month;

-- Drop user quota usage table
DROP TABLE IF EXISTS user_quota_usage;

-- SQLite doesn't support dropping columns, so we need to recreate the table
-- This is the existing jobs table structure without first_analyzed_at
CREATE TABLE jobs_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    company TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT,
    description TEXT,
    location TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    analyzed_at TIMESTAMP,
    match_score REAL,
    analysis_result TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy data from current jobs table
INSERT INTO jobs_temp SELECT id, user_id, company, title, url, description, location, created_at, updated_at, analyzed_at, match_score, analysis_result FROM jobs;

-- Drop the old table
DROP TABLE jobs;

-- Rename temp table to jobs
ALTER TABLE jobs_temp RENAME TO jobs;

-- Recreate indexes
CREATE INDEX idx_jobs_user_id ON jobs(user_id);
CREATE INDEX idx_jobs_created_at ON jobs(created_at);