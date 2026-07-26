-- Per-user MFA/2FA configuration (TOTP secret + recovery code hashes).
-- Identity configuration entity, not a Runtime Aggregate (like user_oauth_identities).

CREATE TABLE IF NOT EXISTS user_mfa (
  user_id UUID PRIMARY KEY REFERENCES users (id),
  totp_secret_encrypted TEXT NOT NULL,
  enabled_at TIMESTAMPTZ NULL,
  recovery_codes_hash JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS user_mfa_enabled_at_idx
  ON user_mfa (enabled_at)
  WHERE enabled_at IS NOT NULL;
