-- Add quota tracking to jobs table
ALTER TABLE jobs ADD COLUMN first_analyzed_at TIMESTAMP;

-- Create user quota usage tracking table
CREATE TABLE IF NOT EXISTS user_quota_usage (
    user_id INTEGER PRIMARY KEY,
    month_year TEXT NOT NULL,  -- Format: "2024-01"
    jobs_analyzed INTEGER DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, month_year),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index for efficient quota lookups
CREATE INDEX idx_user_quota_month ON user_quota_usage(user_id, month_year);