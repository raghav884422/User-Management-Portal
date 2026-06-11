-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id  SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    dob  DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on name for faster queries
CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);
