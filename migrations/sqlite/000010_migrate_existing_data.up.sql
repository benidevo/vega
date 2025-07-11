-- Data migration for existing installations
-- This migration assigns existing data to the appropriate user

-- Strategy 1: Single User Installation (most common)
-- If there's only one user, assign all data to that user

-- First, remove the default value so future inserts require explicit user_id
-- Note: In SQLite, we can't modify column constraints directly, 
-- but the application will enforce this through proper queries

-- Update companies to belong to the first user if not already set
UPDATE companies 
SET user_id = (SELECT id FROM users ORDER BY id LIMIT 1)
WHERE user_id = 1;

-- Update jobs to belong to the same user as their company
UPDATE jobs 
SET user_id = (SELECT c.user_id FROM companies c WHERE c.id = jobs.company_id)
WHERE user_id = 1;

-- Update match_results to belong to the same user as their job
UPDATE match_results 
SET user_id = (SELECT j.user_id FROM jobs j WHERE j.id = match_results.job_id)
WHERE user_id = 1;

-- For multi-user installations, admin intervention may be required
-- Consider creating a separate migration tool or admin interface