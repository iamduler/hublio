-- name: InsertEvent :exec
INSERT INTO events (
  id, organization_id, workspace_id, aggregate_type, aggregate_id, execution_id,
  category, event_name, correlation_id, payload, metadata, published_by, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10, $11, $12, $13
);

-- name: ListEventsByWorkspace :many
SELECT id, organization_id, workspace_id, aggregate_type, aggregate_id, execution_id,
       category, event_name, correlation_id, payload, metadata, published_by, created_at
FROM events
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListEventsByWorkspaceAndExecution :many
SELECT id, organization_id, workspace_id, aggregate_type, aggregate_id, execution_id,
       category, event_name, correlation_id, payload, metadata, published_by, created_at
FROM events
WHERE workspace_id = $1 AND execution_id = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: InsertAuditLog :exec
INSERT INTO audit_logs (
  id, organization_id, workspace_id, actor_type, actor_id, action, resource_type,
  resource_id, request_id, correlation_id, ip, user_agent, metadata, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11, $12, $13, $14
);
