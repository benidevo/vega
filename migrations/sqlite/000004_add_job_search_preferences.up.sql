CREATE TABLE job_search_preferences (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    job_title TEXT NOT NULL,
    location TEXT NOT NULL,
    skills JSON,
    max_age INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_job_search_preferences_user_id ON job_search_preferences(user_id);
CREATE INDEX idx_job_search_preferences_active ON job_search_preferences(user_id, is_active);
