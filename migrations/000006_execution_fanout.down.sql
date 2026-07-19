DROP INDEX IF EXISTS idx_executions_intent_id;

-- Restore v1 uniqueness (only safe if each Intent still has at most one Execution).
ALTER TABLE executions DROP CONSTRAINT IF EXISTS executions_intent_id_unique;
ALTER TABLE executions ADD CONSTRAINT executions_intent_id_unique UNIQUE (intent_id);
