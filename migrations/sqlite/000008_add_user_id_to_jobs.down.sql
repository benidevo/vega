-- Revert jobs table to single-tenant structure
-- WARNING: This will lose user association data

-- Drop the new indexes
DROP INDEX IF EXISTS idx_jobs_user_id;
DROP INDEX IF EXISTS idx_jobs_user_source_url;

-- SQLite doesn't support dropping columns directly, so we need to recreate the table
CREATE TABLE jobs_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    location TEXT,
    match_score INTEGER DEFAULT NULL,
    job_type INTEGER NOT NULL DEFAULT 0,
    source_url TEXT UNIQUE,
    required_skills TEXT,
    application_url TEXT,
    company_id INTEGER NOT NULL,
    status INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (company_id) REFERENCES companies(id)
);

-- Copy data (losing user_id information)
INSERT INTO jobs_new (id, title, description, location, match_score, job_type, 
                     source_url, required_skills, application_url, company_id, 
                     status, notes, created_at, updated_at)
SELECT id, title, description, location, match_score, job_type, 
       source_url, required_skills, application_url, company_id, 
       status, notes, created_at, updated_at FROM jobs;

-- Replace the old table
DROP TABLE jobs;
ALTER TABLE jobs_new RENAME TO jobs;

-- Recreate the original indexes
CREATE INDEX idx_jobs_title ON jobs(title);
CREATE INDEX idx_jobs_source_url ON jobs(source_url);
CREATE INDEX idx_jobs_company_id ON jobs(company_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_updated_at ON jobs(updated_at);
CREATE INDEX idx_jobs_match_score ON jobs(match_score);