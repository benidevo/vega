-- Revert companies table to single-tenant structure
-- WARNING: This will lose user association data

-- Drop the new indexes
DROP INDEX IF EXISTS idx_companies_user_name;
DROP INDEX IF EXISTS idx_companies_user_id;

-- SQLite doesn't support dropping columns directly, so we need to recreate the table
CREATE TABLE companies_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Copy data (losing user_id information)
INSERT INTO companies_new (id, name, created_at, updated_at)
SELECT id, name, created_at, updated_at FROM companies;

-- Replace the old table
DROP TABLE companies;
ALTER TABLE companies_new RENAME TO companies;

-- Recreate the original index
CREATE INDEX idx_companies_name ON companies(name);