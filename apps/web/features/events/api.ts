import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type { DomainEvent, EventsQuery } from "./types";

/** Events via JWT proxy (`/api/go` → Go Bearer + X-Workspace-ID). */
export const eventsApi = {
  list(query: EventsQuery = {}) {
    return api
      .get<SuccessEnvelope<DomainEvent[]>>("/events", {
        execution_id: query.execution_id,
        limit: query.limit,
      })
      .then((res) => unwrapData<DomainEvent[]>(res));
  },
};
