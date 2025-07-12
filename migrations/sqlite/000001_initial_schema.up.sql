-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    password TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    last_login TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_users_username ON users(username);

-- Create companies table (shared across all users)
CREATE TABLE IF NOT EXISTS companies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_companies_name ON companies(name);

-- Create jobs table with multi-tenancy
CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    location TEXT,
    job_type INTEGER NOT NULL DEFAULT 0,
    source_url TEXT,
    required_skills TEXT, -- JSON array
    application_url TEXT,
    company_id INTEGER NOT NULL,
    status INTEGER NOT NULL DEFAULT 0,
    match_score INTEGER,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (company_id) REFERENCES companies(id),
    CHECK (match_score IS NULL OR (match_score >= 0 AND match_score <= 100))
);
CREATE INDEX idx_jobs_user_id ON jobs(user_id);
CREATE INDEX idx_jobs_title ON jobs(title);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_match_score ON jobs(match_score);
CREATE INDEX idx_jobs_company_id ON jobs(company_id);
CREATE UNIQUE INDEX idx_jobs_user_id_source_url ON jobs(user_id, source_url);
CREATE INDEX idx_jobs_user_id_status ON jobs(user_id, status);
CREATE INDEX idx_jobs_user_id_created_at ON jobs(user_id, created_at DESC);

-- Create profiles table
CREATE TABLE IF NOT EXISTS profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    first_name TEXT DEFAULT '',
    last_name TEXT DEFAULT '',
    title TEXT DEFAULT '',
    industry INTEGER DEFAULT 64, -- Default to IndustryUnspecified (64)
    career_summary TEXT DEFAULT '',
    skills TEXT DEFAULT '', -- Stored as JSON string
    phone_number TEXT DEFAULT '',
    email TEXT DEFAULT '',
    location TEXT DEFAULT '',
    linkedin_profile TEXT DEFAULT '',
    github_profile TEXT DEFAULT '',
    website TEXT DEFAULT '',
    context TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create work_experiences table
CREATE TABLE IF NOT EXISTS work_experiences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id INTEGER NOT NULL,
    company TEXT NOT NULL,
    title TEXT NOT NULL,
    location TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    description TEXT,
    current BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
);

-- Create education table
CREATE TABLE IF NOT EXISTS education (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id INTEGER NOT NULL,
    institution TEXT NOT NULL,
    degree TEXT NOT NULL,
    field_of_study TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
);

-- Create certifications table
CREATE TABLE IF NOT EXISTS certifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    issuing_org TEXT NOT NULL,
    issue_date TIMESTAMP NOT NULL,
    expiry_date TIMESTAMP,
    credential_id TEXT,
    credential_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_work_experiences_profile_id ON work_experiences(profile_id);
CREATE INDEX IF NOT EXISTS idx_work_experiences_start_date ON work_experiences(start_date);
CREATE INDEX IF NOT EXISTS idx_education_profile_id ON education(profile_id);
CREATE INDEX IF NOT EXISTS idx_education_start_date ON education(start_date);
CREATE INDEX IF NOT EXISTS idx_certifications_profile_id ON certifications(profile_id);
CREATE INDEX IF NOT EXISTS idx_certifications_issue_date ON certifications(issue_date);

-- Create trigger to automatically create a profile when a user is created
CREATE TRIGGER create_profile_after_user_insert
AFTER INSERT ON users
FOR EACH ROW
BEGIN
  INSERT INTO profiles (user_id, created_at) VALUES (NEW.id, CURRENT_TIMESTAMP);
END;

-- Create match_results table with multi-tenancy
CREATE TABLE IF NOT EXISTS match_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    job_id INTEGER NOT NULL,
    match_score INTEGER NOT NULL,
    strengths TEXT, -- JSON array
    weaknesses TEXT, -- JSON array
    highlights TEXT, -- JSON array
    feedback TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE,
    CHECK (match_score >= 0 AND match_score <= 100)
);
CREATE INDEX idx_match_results_user_id ON match_results(user_id);
CREATE INDEX idx_match_results_job_id ON match_results(job_id);
CREATE INDEX idx_match_results_user_id_created_at ON match_results(user_id, created_at DESC);