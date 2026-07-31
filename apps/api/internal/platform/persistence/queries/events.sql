-- name: InsertEvent :exec
INSERT INTO events (
  id, organization_id, workspace_id, aggregate_type, aggregate_id, execution_id,
  category, event_name, correlation_id, payload, metadata, published_by, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10, $11, $12, $13
);

-- name: ListEventsFiltered :many
SELECT id, organization_id, workspace_id, aggregate_type, aggregate_id, execution_id,
       category, event_name, correlation_id, payload, metadata, published_by, created_at
FROM events
WHERE workspace_id = sqlc.arg(workspace_id)
  AND (sqlc.narg(execution_id)::uuid IS NULL OR execution_id = sqlc.narg(execution_id))
  AND (sqlc.narg(category)::text IS NULL OR category = sqlc.narg(category)::event_category)
  AND (
    sqlc.narg(cursor_created_at)::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg(cursor_created_at)::timestamptz, sqlc.narg(cursor_id)::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(fetch_limit);

-- name: InsertAuditLog :exec
INSERT INTO audit_logs (
  id, organization_id, workspace_id, actor_type, actor_id, action, resource_type,
  resource_id, request_id, correlation_id, ip, user_agent, metadata, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11, $12, $13, $14
);
