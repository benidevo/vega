-- Revert match_results table to single-tenant structure
-- WARNING: This will lose user association data

-- Drop the new indexes
DROP INDEX IF EXISTS idx_match_results_user_id;
DROP INDEX IF EXISTS idx_match_results_user_created;

-- SQLite doesn't support dropping columns directly, so we need to recreate the table
CREATE TABLE match_results_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id INTEGER NOT NULL,
    match_score INTEGER NOT NULL CHECK (match_score >= 0 AND match_score <= 100),
    strengths TEXT,
    weaknesses TEXT,
    highlights TEXT,
    feedback TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

-- Copy data (losing user_id information)
INSERT INTO match_results_new (id, job_id, match_score, strengths, weaknesses, 
                              highlights, feedback, created_at)
SELECT id, job_id, match_score, strengths, weaknesses, 
       highlights, feedback, created_at FROM match_results;

-- Replace the old table
DROP TABLE match_results;
ALTER TABLE match_results_new RENAME TO match_results;

-- Recreate the original indexes
CREATE INDEX idx_match_results_job_created ON match_results(job_id, created_at DESC);
CREATE INDEX idx_match_results_created ON match_results(created_at DESC);