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

export interface CursorPagination {
  next_cursor: string;
  has_next: boolean;
  limit: number;
}

export interface EventsPage {
  items: DomainEvent[];
  pagination: CursorPagination;
}

export interface EventsQuery {
  execution_id?: string;
  category?: EventCategory | "all";
  cursor?: string;
  limit?: number;
}
