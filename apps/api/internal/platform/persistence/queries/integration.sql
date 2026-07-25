-- name: InsertConnector :exec
INSERT INTO connectors (
  id, code, name, vendor, category, version, status,
  description, homepage, documentation_url, created_at, updated_at, deleted_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11, $12, $13
);

-- name: UpdateConnector :exec
UPDATE connectors
SET status = $2,
    updated_at = $3,
    deleted_at = $4
WHERE id = $1;

-- name: GetConnectorByID :one
SELECT id, code, name, vendor, category, version, status,
       description, homepage, documentation_url, created_at, updated_at, deleted_at
FROM connectors
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetConnectorByCode :one
SELECT id, code, name, vendor, category, version, status,
       description, homepage, documentation_url, created_at, updated_at, deleted_at
FROM connectors
WHERE code = $1 AND deleted_at IS NULL;

-- name: ListConnectors :many
SELECT id, code, name, vendor, category, version, status,
       description, homepage, documentation_url, created_at, updated_at, deleted_at
FROM connectors
WHERE deleted_at IS NULL
ORDER BY created_at ASC;

-- name: InsertConnectorCapability :exec
INSERT INTO connector_capabilities (
  id, connector_id, capability_code, display_name, status, is_async, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: ListCapabilitiesByConnector :many
SELECT id, connector_id, capability_code, display_name, status, is_async, created_at, updated_at
FROM connector_capabilities
WHERE connector_id = $1
ORDER BY created_at ASC;

-- name: InsertConnection :exec
INSERT INTO connections (
  id, active_credential_id, workspace_id, connector_id, name, is_default,
  description, environment, status, config, retry_policy, timeout_seconds,
  created_at, updated_at, deleted_at
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10, $11, $12,
  $13, $14, $15
);

-- name: UpdateConnection :exec
UPDATE connections
SET active_credential_id = $2,
    name = $3,
    description = $4,
    status = $5,
    config = $6,
    retry_policy = $7,
    timeout_seconds = $8,
    updated_at = $9,
    deleted_at = $10
WHERE id = $1;

-- name: GetConnectionByID :one
SELECT id, active_credential_id, workspace_id, connector_id, name, is_default,
       description, environment, status, config, retry_policy, timeout_seconds,
       created_at, updated_at, deleted_at
FROM connections
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListConnectionsByWorkspace :many
SELECT id, active_credential_id, workspace_id, connector_id, name, is_default,
       description, environment, status, config, retry_policy, timeout_seconds,
       created_at, updated_at, deleted_at
FROM connections
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: InsertCredential :exec
INSERT INTO credentials (
  id, connection_id, type, status, version, encrypted_secret,
  expires_at, rotated_at, created_at, updated_at, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10, $11
);

-- name: UpdateCredential :exec
UPDATE credentials
SET status = $2,
    rotated_at = $3,
    updated_at = $4
WHERE id = $1;

-- name: GetCredentialByID :one
SELECT id, connection_id, type, status, version, encrypted_secret,
       expires_at, rotated_at, created_at, updated_at, created_by
FROM credentials
WHERE id = $1;

-- name: GetActiveCredentialByConnection :one
SELECT id, connection_id, type, status, version, encrypted_secret,
       expires_at, rotated_at, created_at, updated_at, created_by
FROM credentials
WHERE connection_id = $1 AND status = 'active'
ORDER BY version DESC
LIMIT 1;

-- name: ListCredentialsByConnection :many
SELECT id, connection_id, type, status, version, encrypted_secret,
       expires_at, rotated_at, created_at, updated_at, created_by
FROM credentials
WHERE connection_id = $1
ORDER BY version DESC;
