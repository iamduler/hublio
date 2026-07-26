-- Platform admin flag, OAuth-only users (nullable password), OAuth identity links, local seed accounts.

DO $$ BEGIN
  CREATE TYPE oauth_provider AS ENUM ('google', 'microsoft', 'github');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS is_platform_admin BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE users
  ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE users
  DROP CONSTRAINT IF EXISTS users_platform_admin_password_check;

ALTER TABLE users
  ADD CONSTRAINT users_platform_admin_password_check
  CHECK (
    (is_platform_admin = TRUE AND password_hash IS NOT NULL)
    OR (is_platform_admin = FALSE)
  );

CREATE TABLE IF NOT EXISTS user_oauth_identities (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users (id),
  provider oauth_provider NOT NULL,
  provider_subject VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_login_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT user_oauth_identities_provider_subject_unique UNIQUE (provider, provider_subject),
  CONSTRAINT user_oauth_identities_user_provider_unique UNIQUE (user_id, provider)
);

CREATE INDEX IF NOT EXISTS user_oauth_identities_user_id_idx
  ON user_oauth_identities (user_id);

-- Seed platform admin + demo tenant (local / development quick login).
-- Passwords: Admin123! / Demo123! (bcrypt cost 12).

INSERT INTO organizations (id, name, status, created_at, updated_at, deleted_at)
VALUES
  ('01900000-0000-7000-8000-000000000001', 'Hublio Platform', 'active', NOW(), NOW(), NULL),
  ('01900000-0000-7000-8000-000000000002', 'Demo Organization', 'active', NOW(), NOW(), NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO workspaces (id, organization_id, name, environment, status, created_at, updated_at, deleted_at)
VALUES
  ('01900000-0000-7000-8000-000000000021', '01900000-0000-7000-8000-000000000001', 'default', 'production', 'active', NOW(), NOW(), NULL),
  ('01900000-0000-7000-8000-000000000022', '01900000-0000-7000-8000-000000000002', 'default', 'production', 'active', NOW(), NOW(), NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (
  id, organization_id, email, full_name, is_active, is_platform_admin, password_hash,
  email_verified_at, password_changed_at, last_login_at, status,
  created_at, updated_at, deleted_at
) VALUES
  (
    '01900000-0000-7000-8000-000000000011',
    '01900000-0000-7000-8000-000000000001',
    'admin@hublio.local',
    'Platform Admin',
    TRUE,
    TRUE,
    '$2b$12$ygusn8wYEo6UTTipp5clsOKZe9KquoOsPFBy0Z3Tx5PT1.E/MCUg2',
    NOW(),
    NOW(),
    NULL,
    'active',
    NOW(),
    NOW(),
    NULL
  ),
  (
    '01900000-0000-7000-8000-000000000012',
    '01900000-0000-7000-8000-000000000002',
    'demo@hublio.local',
    'Demo User',
    TRUE,
    FALSE,
    '$2b$12$EW8rWpK2FfUmIGRLHZtf0eoUilIt35IfEY3qnjl4d5aT.mm2ilaMS',
    NOW(),
    NOW(),
    NULL,
    'active',
    NOW(),
    NOW(),
    NULL
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO workspace_users (workspace_id, user_id, role, created_at)
VALUES
  ('01900000-0000-7000-8000-000000000021', '01900000-0000-7000-8000-000000000011', 'owner', NOW()),
  ('01900000-0000-7000-8000-000000000022', '01900000-0000-7000-8000-000000000012', 'owner', NOW())
ON CONFLICT (workspace_id, user_id) DO NOTHING;
