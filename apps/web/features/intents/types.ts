import type { Schemas } from "@/lib/api/sdk";

/**
 * Request DTO from OpenAPI. JSONB `payload` is typed loosely for form use
 * (OpenAPI generates `Record<string, never>` for free-form objects).
 */
export type CreateIntentPayload = Omit<
  Schemas["SubmitIntentRequest"],
  "payload"
> & {
  payload?: Record<string, unknown>;
};

/** Response entities are not yet named in OpenAPI — keep local until promoted. */
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

export interface CreateIntentResult {
  intent: Intent;
  execution?: ExecutionSummary;
  executions?: ExecutionSummary[];
}
