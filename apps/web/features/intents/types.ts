export type IntentStatus = "submitted" | "accepted" | "rejected" | "expired";

export interface Intent {
  id: string;
  organization_id?: string;
  workspace_id?: string;
  connection_id: string;
  capability: string;
  payload?: Record<string, unknown>;
  status: IntentStatus;
  correlation_id?: string;
  submitted_at?: string;
  created_at?: string;
}

export interface ExecutionSummary {
  id: string;
  status: string;
}

export interface CreateIntentPayload {
  connection_id: string;
  capability: string;
  payload: Record<string, unknown>;
  correlation_id?: string;
}

export interface CreateIntentResult {
  intent: Intent;
  execution?: ExecutionSummary;
  executions?: ExecutionSummary[];
}
