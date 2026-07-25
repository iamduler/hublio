-- name: InsertIntent :exec
INSERT INTO intents (
  id, organization_id, workspace_id, connection_id, capability, payload,
  status, correlation_id, idempotency_key, submitted_at, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6,
  $7, $8, $9, $10, $11
);

-- name: UpdateIntent :exec
UPDATE intents
SET status = $2
WHERE id = $1;

-- name: GetIntentByID :one
SELECT id, organization_id, workspace_id, connection_id, capability, payload,
       status, correlation_id, idempotency_key, submitted_at, created_at
FROM intents
WHERE id = $1;

-- name: InsertExecution :exec
INSERT INTO executions (
  id, intent_id, status, result, retry_attempt, current_step_no, context,
  failure_reason, started_at, completed_at, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11
);

-- name: UpdateExecution :exec
UPDATE executions
SET status = $2,
    result = $3,
    retry_attempt = $4,
    current_step_no = $5,
    context = $6,
    failure_reason = $7,
    started_at = $8,
    completed_at = $9
WHERE id = $1;

-- name: GetExecutionByID :one
SELECT id, intent_id, status, result, retry_attempt, current_step_no, context,
       failure_reason, started_at, completed_at, created_at
FROM executions
WHERE id = $1;

-- name: GetExecutionByIntentID :one
SELECT id, intent_id, status, result, retry_attempt, current_step_no, context,
       failure_reason, started_at, completed_at, created_at
FROM executions
WHERE intent_id = $1
ORDER BY created_at ASC
LIMIT 1;

-- name: ListExecutionsByIntentID :many
SELECT id, intent_id, status, result, retry_attempt, current_step_no, context,
       failure_reason, started_at, completed_at, created_at
FROM executions
WHERE intent_id = $1
ORDER BY created_at ASC;

-- name: InsertExecutionStep :exec
INSERT INTO execution_steps (
  id, execution_id, step_no, step_type, status, retry_attempt, duration_ms,
  error_message, error_code, started_at, completed_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11
);

-- name: UpdateExecutionStep :exec
UPDATE execution_steps
SET status = $2,
    retry_attempt = $3,
    duration_ms = $4,
    error_message = $5,
    error_code = $6,
    started_at = $7,
    completed_at = $8
WHERE id = $1;

-- name: ListExecutionStepsByExecution :many
SELECT id, execution_id, step_no, step_type, status, retry_attempt, duration_ms,
       error_message, error_code, started_at, completed_at
FROM execution_steps
WHERE execution_id = $1
ORDER BY step_no ASC;

-- name: InsertExecutionSnapshot :exec
INSERT INTO execution_snapshots (
  id, execution_id, step_id, snapshot_type, snapshot, content_type, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (id) DO NOTHING;

-- name: ListExecutionSnapshotsByExecution :many
SELECT id, execution_id, step_id, snapshot_type, snapshot, content_type, created_at
FROM execution_snapshots
WHERE execution_id = $1
ORDER BY created_at ASC;

-- name: InsertExecutionTimeline :exec
INSERT INTO execution_timelines (
  id, execution_id, event, message, metadata, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6
)
ON CONFLICT (id) DO NOTHING;

-- name: ListExecutionTimelinesByExecution :many
SELECT id, execution_id, event, message, metadata, created_at
FROM execution_timelines
WHERE execution_id = $1
ORDER BY created_at ASC;

-- name: InsertIdempotencyKey :exec
INSERT INTO idempotency_keys (
  id, organization_id, workspace_id, idempotency_key, intent_id, expires_at, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
);

-- name: GetIdempotencyKeyByOrgWorkspaceKey :one
SELECT id, organization_id, workspace_id, idempotency_key, intent_id, expires_at, created_at
FROM idempotency_keys
WHERE organization_id = $1 AND workspace_id = $2 AND idempotency_key = $3;
