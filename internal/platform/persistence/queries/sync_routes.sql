-- name: InsertSyncRoute :exec
INSERT INTO sync_routes (
  id, workspace_id, source_connection_id, name, status, trigger_type,
  resource_types, schedule, filter, idempotency_rule, activities, reverse, retry_policy,
  webhook_secret, created_at, updated_at, deleted_at
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10, $11, $12, $13,
  $14, $15, $16, $17
);

-- name: UpdateSyncRoute :exec
UPDATE sync_routes
SET source_connection_id = $2,
    name = $3,
    status = $4,
    trigger_type = $5,
    resource_types = $6,
    schedule = $7,
    filter = $8,
    idempotency_rule = $9,
    activities = $10,
    reverse = $11,
    retry_policy = $12,
    webhook_secret = $13,
    updated_at = $14,
    deleted_at = $15
WHERE id = $1;

-- name: GetSyncRouteByID :one
SELECT id, workspace_id, source_connection_id, name, status, trigger_type,
       resource_types, schedule, filter, idempotency_rule, activities, reverse, retry_policy,
       webhook_secret, created_at, updated_at, deleted_at
FROM sync_routes
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListSyncRoutesByWorkspace :many
SELECT id, workspace_id, source_connection_id, name, status, trigger_type,
       resource_types, schedule, filter, idempotency_rule, activities, reverse, retry_policy,
       webhook_secret, created_at, updated_at, deleted_at
FROM sync_routes
WHERE workspace_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: UpsertSyncRouteWatermark :exec
INSERT INTO sync_route_watermarks (sync_route_id, resource_type, cursor, updated_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (sync_route_id, resource_type) DO UPDATE
SET cursor = EXCLUDED.cursor,
    updated_at = EXCLUDED.updated_at;

-- name: GetSyncRouteWatermark :one
SELECT sync_route_id, resource_type, cursor, updated_at
FROM sync_route_watermarks
WHERE sync_route_id = $1 AND resource_type = $2;

-- name: ListSyncRouteWatermarks :many
SELECT sync_route_id, resource_type, cursor, updated_at
FROM sync_route_watermarks
WHERE sync_route_id = $1
ORDER BY resource_type ASC;
