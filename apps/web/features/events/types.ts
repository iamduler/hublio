export type EventCategory = "runtime" | "business" | "system";

export interface DomainEvent {
  id: string;
  organization_id?: string;
  workspace_id?: string;
  aggregate_type: string;
  aggregate_id: string;
  execution_id?: string;
  category: EventCategory;
  event_name: string;
  correlation_id?: string;
  payload?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  published_by?: string;
  created_at?: string;
}

export interface EventsQuery {
  execution_id?: string;
  /** Client-side filter — backend list API does not yet accept category. */
  category?: EventCategory | "all";
  limit?: number;
}
