export type SyncRouteStatus = "draft" | "enabled" | "disabled";
export type SyncRouteTrigger = "webhook" | "schedule" | "both";
export type GroupMode = "sequential" | "parallel";

export interface SyncRouteStep {
  destination_connection_id: string;
  capability: string;
  mapping_key?: string;
}

export interface SyncRouteActivity {
  group_mode: GroupMode;
  steps: SyncRouteStep[];
}

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

export interface CreateSyncRoutePayload {
  source_connection_id: string;
  name: string;
  trigger_type: SyncRouteTrigger;
  resource_types: string[];
  activities: SyncRouteActivity[];
}

export interface SyncRouteWatermark {
  resource_type: string;
  cursor?: Record<string, unknown>;
  updated_at?: string;
}
