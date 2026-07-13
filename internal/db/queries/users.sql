-- name: CreateUser :one
INSERT INTO users(
	email, password, full_name, age, status, level
) VALUES (
	$1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET
	password = COALESCE(sqlc.narg(password), password),
	full_name = COALESCE(sqlc.narg(full_name), full_name),
	age = COALESCE(sqlc.narg(age), age),
	status = COALESCE(sqlc.narg(status), status),
	level = COALESCE(sqlc.narg(level), level)
WHERE uuid = sqlc.arg(uuid) AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteUser :one
UPDATE users SET
	deleted_at = NOW()
WHERE uuid = sqlc.arg(uuid)::uuid AND deleted_at IS NULL
RETURNING *;

-- name: RestoreUser :one
UPDATE users SET
	deleted_at = NULL
WHERE uuid = sqlc.arg(uuid)::uuid AND deleted_at IS NOT NULL
RETURNING *;

-- name: TrashUser :one
DELETE FROM users
WHERE uuid = sqlc.arg(uuid)::uuid AND deleted_at IS NOT NULL
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE uuid = $1 AND deleted_at IS NULL;

-- name: ListUsersWUserIdAsc :many
SELECT * FROM users
WHERE deleted_at IS NULL
AND (
	sqlc.narg(search)::TEXT IS NULL
	OR sqlc.narg(search)::TEXT = ''
	OR email ILIKE '%' || sqlc.narg(search) || '%'
	OR full_name ILIKE '%' || sqlc.narg(search) || '%'
)
ORDER BY id ASC
LIMIT $1 OFFSET $2;

-- name: ListUsersWUserIdDesc :many
SELECT * FROM users
WHERE deleted_at IS NULL
AND (
	sqlc.narg(search)::TEXT IS NULL
	OR sqlc.narg(search)::TEXT = ''
	OR email ILIKE '%' || sqlc.narg(search) || '%'
	OR full_name ILIKE '%' || sqlc.narg(search) || '%'
)
ORDER BY id DESC
LIMIT $1 OFFSET $2;

-- name: ListUsersWCreatedAtAsc :many
SELECT * FROM users
WHERE deleted_at IS NULL
AND (
	sqlc.narg(search)::TEXT IS NULL
	OR sqlc.narg(search)::TEXT = ''
	OR email ILIKE '%' || sqlc.narg(search) || '%'
	OR full_name ILIKE '%' || sqlc.narg(search) || '%'
)
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: ListUsersWCreatedAtDesc :many
SELECT * FROM users
WHERE deleted_at IS NULL
AND (
	sqlc.narg(search)::TEXT IS NULL
	OR sqlc.narg(search)::TEXT = ''
	OR email ILIKE '%' || sqlc.narg(search) || '%'
	OR full_name ILIKE '%' || sqlc.narg(search) || '%'
)
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(1) FROM users
WHERE (
	sqlc.narg(deleted)::bool IS NULL
	OR (sqlc.narg(deleted)::bool = TRUE AND deleted_at IS NOT NULL)
	OR (sqlc.narg(deleted)::bool = FALSE AND deleted_at IS NULL)
)
AND (
	sqlc.narg(search)::TEXT IS NULL
	OR sqlc.narg(search)::TEXT = ''
	OR email ILIKE '%' || sqlc.narg(search) || '%'
	OR full_name ILIKE '%' || sqlc.narg(search) || '%'
);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: UpdatePassword :one
UPDATE users SET
	password = sqlc.arg(password)
WHERE uuid = sqlc.arg(uuid) AND deleted_at IS NULL
RETURNING *;