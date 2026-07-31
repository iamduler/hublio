-- Keyset list indexes for GET /api/v1/events (workspace + optional category).

CREATE INDEX IF NOT EXISTS idx_events_workspace_created_id
  ON events (workspace_id, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_events_workspace_category_created_id
  ON events (workspace_id, category, created_at DESC, id DESC);
