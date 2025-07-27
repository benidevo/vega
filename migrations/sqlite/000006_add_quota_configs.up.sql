-- Add quota configuration table for dynamic quota limits
CREATE TABLE quota_configs (
    quota_type TEXT PRIMARY KEY,
    free_limit INTEGER NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed with default quota values
INSERT INTO quota_configs (quota_type, free_limit, description) VALUES
('ai_analysis_monthly', 10, 'AI job analysis per month');