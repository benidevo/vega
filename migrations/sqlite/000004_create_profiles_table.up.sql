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
