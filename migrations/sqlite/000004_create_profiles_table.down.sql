-- Drop indexes
DROP INDEX IF EXISTS idx_certifications_issue_date;
DROP INDEX IF EXISTS idx_certifications_profile_id;
DROP INDEX IF EXISTS idx_education_start_date;
DROP INDEX IF EXISTS idx_education_profile_id;
DROP INDEX IF EXISTS idx_work_experiences_start_date;
DROP INDEX IF EXISTS idx_work_experiences_profile_id;
DROP INDEX IF EXISTS idx_profiles_user_id;

-- Drop tables in reverse order (cascade constraints)
DROP TABLE IF EXISTS certifications;
DROP TABLE IF EXISTS education;
DROP TABLE IF EXISTS work_experiences;
DROP TABLE IF EXISTS profiles;