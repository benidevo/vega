-- Drop quota tracking index
DROP INDEX IF EXISTS idx_user_quota_month;

-- Drop user quota usage table
DROP TABLE IF EXISTS user_quota_usage;

-- SQLite doesn't support dropping columns, so we need to recreate the table
-- This is the existing jobs table structure without first_analyzed_at
CREATE TABLE jobs_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    location TEXT,
    job_type INTEGER NOT NULL DEFAULT 0,
    source_url TEXT,
    required_skills TEXT, -- JSON array
    application_url TEXT,
    company_id INTEGER NOT NULL,
    status INTEGER NOT NULL DEFAULT 0,
    match_score INTEGER,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(id),
    CHECK (match_score IS NULL OR (match_score >= 0 AND match_score <= 100))
);

-- Copy data from current jobs table (excluding first_analyzed_at)
INSERT INTO jobs_temp SELECT id, user_id, title, description, location, job_type, source_url, required_skills, application_url, company_id, status, match_score, notes, created_at, updated_at FROM jobs;

-- Drop the old table
DROP TABLE jobs;

-- Rename temp table to jobs
ALTER TABLE jobs_temp RENAME TO jobs;

-- Recreate indexes
CREATE INDEX idx_jobs_user_id ON jobs(user_id);
CREATE INDEX idx_jobs_title ON jobs(title);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_match_score ON jobs(match_score);
CREATE INDEX idx_jobs_company_id ON jobs(company_id);
CREATE UNIQUE INDEX idx_jobs_user_id_source_url ON jobs(user_id, source_url);
CREATE INDEX idx_jobs_user_id_status ON jobs(user_id, status);
CREATE INDEX idx_jobs_user_id_created_at ON jobs(user_id, created_at DESC);