-- name: GetOrganizationByID :one
SELECT id, name, status, created_at, updated_at, deleted_at
FROM organizations
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetOrganizationByName :one
SELECT id, name, status, created_at, updated_at, deleted_at
FROM organizations
WHERE name = $1 AND deleted_at IS NULL;

-- name: InsertOrganization :exec
INSERT INTO organizations (id, name, status, created_at, updated_at, deleted_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateOrganization :exec
UPDATE organizations
SET name = $2,
    status = $3,
    updated_at = $4,
    deleted_at = $5
WHERE id = $1;

-- name: GetWorkspaceByID :one
SELECT id, organization_id, name, environment, status, created_at, updated_at, deleted_at
FROM workspaces
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListWorkspacesByOrganization :many
SELECT id, organization_id, name, environment, status, created_at, updated_at, deleted_at
FROM workspaces
WHERE organization_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: InsertWorkspace :exec
INSERT INTO workspaces (id, organization_id, name, environment, status, created_at, updated_at, deleted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpdateWorkspace :exec
UPDATE workspaces
SET name = $2,
    environment = $3,
    status = $4,
    updated_at = $5,
    deleted_at = $6
WHERE id = $1;

-- name: GetUserByID :one
SELECT id, organization_id, email, full_name, is_active, password_hash,
       email_verified_at, password_changed_at, last_login_at, status,
       created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, organization_id, email, full_name, is_active, password_hash,
       email_verified_at, password_changed_at, last_login_at, status,
       created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: InsertUser :exec
INSERT INTO users (
  id, organization_id, email, full_name, is_active, password_hash,
  email_verified_at, password_changed_at, last_login_at, status,
  created_at, updated_at, deleted_at
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10,
  $11, $12, $13
);

-- name: UpdateUser :exec
UPDATE users
SET full_name = $2,
    is_active = $3,
    password_hash = $4,
    last_login_at = $5,
    status = $6,
    updated_at = $7,
    deleted_at = $8
WHERE id = $1;

-- name: InsertWorkspaceUser :exec
INSERT INTO workspace_users (workspace_id, user_id, role, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetWorkspaceUser :one
SELECT workspace_id, user_id, role, created_at
FROM workspace_users
WHERE workspace_id = $1 AND user_id = $2;

-- name: ListWorkspaceUsersByWorkspace :many
SELECT workspace_id, user_id, role, created_at
FROM workspace_users
WHERE workspace_id = $1
ORDER BY created_at ASC;

-- name: ListWorkspaceUsersByUser :many
SELECT workspace_id, user_id, role, created_at
FROM workspace_users
WHERE user_id = $1
ORDER BY created_at ASC;

-- name: GetAPIKeyByID :one
SELECT id, workspace_id, name, key_hash, last_used_at, expires_at, status, prefix,
       last_used_ip, last_used_user_agent, created_at, updated_at, deleted_at
FROM api_keys
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetAPIKeyByPrefix :one
SELECT id, workspace_id, name, key_hash, last_used_at, expires_at, status, prefix,
       last_used_ip, last_used_user_agent, created_at, updated_at, deleted_at
FROM api_keys
WHERE prefix = $1 AND deleted_at IS NULL;

-- name: ListAPIKeysByWorkspace :many
SELECT id, workspace_id, name, key_hash, last_used_at, expires_at, status, prefix,
       last_used_ip, last_used_user_agent, created_at, updated_at, deleted_at
FROM api_keys
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: InsertAPIKey :exec
INSERT INTO api_keys (
  id, workspace_id, name, key_hash, last_used_at, expires_at, status, prefix,
  last_used_ip, last_used_user_agent, created_at, updated_at, deleted_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8,
  $9, $10, $11, $12, $13
);

-- name: UpdateAPIKey :exec
UPDATE api_keys
SET name = $2,
    key_hash = $3,
    expires_at = $4,
    status = $5,
    prefix = $6,
    updated_at = $7,
    deleted_at = $8
WHERE id = $1;

-- name: TouchAPIKeyLastUsed :exec
UPDATE api_keys
SET last_used_at = $2,
    updated_at = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: Ping :one
SELECT 1::int AS ok;
