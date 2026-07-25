"use client";

import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { eventsApi } from "./api";
import type { EventsQuery } from "./types";

export function useEvents(query: EventsQuery = {}) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.events(workspaceId ?? "none", query.execution_id),
    queryFn: () => eventsApi.list(query),
    enabled: Boolean(workspaceId),
  });
}
