-- Add user profiles cache table for storing Infor user profile data
-- This avoids session storage size limits and enables horizontal scaling
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id VARCHAR(100) PRIMARY KEY,
    profile_data JSONB NOT NULL,
    fetched_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for efficient expiration cleanup
CREATE INDEX idx_user_profiles_expires_at ON user_profiles(expires_at);

-- GIN index for fast role/group queries (for authorization)
CREATE INDEX idx_user_profiles_groups ON user_profiles USING GIN ((profile_data->'groups'));

-- Add comments
COMMENT ON TABLE user_profiles IS 'Cache for Infor user management API profiles with TTL';
COMMENT ON COLUMN user_profiles.user_id IS 'Infor user ID from profile.id';
COMMENT ON COLUMN user_profiles.profile_data IS 'Full JSONB profile including roles/groups';
COMMENT ON COLUMN user_profiles.expires_at IS 'Cache expiration timestamp (TTL: 15 minutes)';
