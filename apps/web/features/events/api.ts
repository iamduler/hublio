import { bff, unwrapData } from "@/lib/api/bff-client";
import type { SuccessEnvelope } from "@/lib/api/types";
import type { DomainEvent, EventsQuery } from "./types";

/** Events context via BFF (X-API-KEY held server-side, workspace-scoped). */
export const eventsApi = {
  list(query: EventsQuery = {}) {
    return bff
      .get<SuccessEnvelope<DomainEvent[]>>("/events", {
        execution_id: query.execution_id,
        limit: query.limit,
      })
      .then((res) => unwrapData<DomainEvent[]>(res));
  },
};
