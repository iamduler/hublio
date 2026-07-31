import { api, type SuccessEnvelope } from "@/lib/api/client";
import type { DomainEvent, EventsPage, EventsQuery } from "./types";

export type { EventsPage };

/** Events via JWT proxy (`/api/go` → Go Bearer + X-Workspace-ID). */
export const eventsApi = {
  list(query: EventsQuery = {}): Promise<EventsPage> {
    const category =
      query.category && query.category !== "all" ? query.category : undefined;
    return api
      .get<
        SuccessEnvelope<DomainEvent[]> & {
          pagination?: EventsPage["pagination"];
        }
      >("/events", {
        execution_id: query.execution_id,
        category,
        cursor: query.cursor,
        limit: query.limit,
      })
      .then((res) => {
        const items = Array.isArray(res.data) ? res.data : [];
        const pagination = res.pagination ?? {
          next_cursor: "",
          has_next: false,
          limit: query.limit ?? 50,
        };
        return { items, pagination };
      });
  },
};
