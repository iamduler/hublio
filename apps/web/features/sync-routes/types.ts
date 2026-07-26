import type { Schemas } from "@/lib/api/sdk";

export type SyncRouteStatus = "draft" | "enabled" | "disabled";
export type SyncRouteTrigger = Schemas["CreateSyncRouteRequest"]["trigger_type"];
export type GroupMode = Schemas["SyncRouteActivityGroup"]["group_mode"];

export type SyncRouteStep = Schemas["SyncRouteActivityGroup"]["steps"][number];
export type SyncRouteActivity = Schemas["SyncRouteActivityGroup"];

/** Response entity — not yet a named OpenAPI schema. */
export interface SyncRoute {
  id: string;
  workspace_id: string;
  source_connection_id: string;
  name: string;
  status: SyncRouteStatus;
  trigger_type: SyncRouteTrigger;
  resource_types?: string[];
  activities?: SyncRouteActivity[];
  schedule?: Record<string, unknown>;
  created_at?: string;
  updated_at?: string;
  /** Returned once on create / rotate. */
  webhook_secret?: string;
}

export type CreateSyncRoutePayload = Omit<
  Schemas["CreateSyncRouteRequest"],
  "schedule" | "filter" | "idempotency_rule" | "retry_policy"
> & {
  schedule?: Record<string, unknown>;
  filter?: Record<string, unknown>;
  idempotency_rule?: Record<string, unknown>;
  retry_policy?: Record<string, unknown>;
};

export type UpdateSyncRoutePayload = Omit<
  Schemas["UpdateSyncRouteRequest"],
  "schedule" | "filter" | "idempotency_rule" | "retry_policy" | "activities"
> & {
  schedule?: Record<string, unknown>;
  filter?: Record<string, unknown>;
  idempotency_rule?: Record<string, unknown>;
  retry_policy?: Record<string, unknown>;
  activities?: SyncRouteActivity[];
};

export interface SyncRouteWatermark {
  resource_type: string;
  cursor?: Record<string, unknown>;
  updated_at?: string;
}
