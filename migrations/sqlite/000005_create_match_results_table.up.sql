-- Create match_results table to store job match analysis history
CREATE TABLE IF NOT EXISTS match_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id INTEGER NOT NULL,
    match_score INTEGER NOT NULL CHECK (match_score >= 0 AND match_score <= 100),
    strengths TEXT, -- JSON array stored as text
    weaknesses TEXT, -- JSON array stored as text
    highlights TEXT, -- JSON array stored as text
    feedback TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_match_results_job_created ON match_results(job_id, created_at DESC);
CREATE INDEX idx_match_results_created ON match_results(created_at DESC);
