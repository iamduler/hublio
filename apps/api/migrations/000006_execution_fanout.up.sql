-- Fan-out: one Intent may own multiple Executions (SyncRoute activities).
-- Steps inside each Execution remain sequential.

ALTER TABLE executions DROP CONSTRAINT IF EXISTS executions_intent_id_unique;

CREATE INDEX IF NOT EXISTS idx_executions_intent_id ON executions (intent_id);
