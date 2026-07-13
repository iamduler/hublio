CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
	id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
	email VARCHAR(100) NOT NULL UNIQUE,
	password VARCHAR(90) NOT NULL,
	full_name VARCHAR(100) NOT NULL,
	age INT CHECK (age >= 1 AND age <= 150),
	status INT NOT NULL DEFAULT 1 CHECK (status IN (1, 2, 3)),
	level INT NOT NULL DEFAULT 1 CHECK (level IN (1, 2, 3)),
	deleted_at TIMESTAMPTZ NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN users.status IS '1: Active, 2: Inactive, 3: Banned';
COMMENT ON COLUMN users.age IS 'Age must be between 1 and 150';
COMMENT ON COLUMN users.level IS '1: Admin, 2: Moderator, 3: Member';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp';

-- Index for the email column
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Trigger to update the updated_at column automatically
CREATE OR REPLACE FUNCTION update_user_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
	NEW.updated_at = NOW();
	RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update the updated_at column automatically
CREATE TRIGGER update_user_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_user_updated_at_column();
