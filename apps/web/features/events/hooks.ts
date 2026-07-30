"use client";

import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { eventsApi } from "./api";
import type { EventsQuery } from "./types";

export function useEvents(query: EventsQuery = {}) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.events(
      workspaceId ?? "none",
      query.execution_id,
      query.category,
      query.limit,
    ),
    queryFn: async () => {
      const events = await eventsApi.list({
        execution_id: query.execution_id,
        limit: query.limit,
      });
      if (!query.category || query.category === "all") return events;
      return events.filter((e) => e.category === query.category);
    },
    enabled: Boolean(workspaceId),
  });
}
