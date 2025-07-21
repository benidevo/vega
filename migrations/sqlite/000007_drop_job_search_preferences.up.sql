-- Drop indexes first
DROP INDEX IF EXISTS idx_job_search_preferences_active;
DROP INDEX IF EXISTS idx_job_search_preferences_user_id;

-- Drop the job_search_preferences table
DROP TABLE IF EXISTS job_search_preferences;