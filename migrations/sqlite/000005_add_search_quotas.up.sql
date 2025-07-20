-- Create table for tracking daily quotas
CREATE TABLE user_daily_quotas (
    user_id INTEGER NOT NULL,
    date TEXT NOT NULL,          -- Format: "2006-01-02"
    quota_key TEXT NOT NULL,     -- 'jobs_found', 'searches_run'
    value INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, date, quota_key),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Index for efficient lookups
CREATE INDEX idx_user_daily_quotas_lookup ON user_daily_quotas(user_id, date);