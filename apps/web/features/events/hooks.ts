"use client";

import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { eventsApi } from "./api";
import type { EventsQuery } from "./types";

const DEFAULT_PAGE_LIMIT = 50;

function listFilters(query: EventsQuery) {
  return {
    execution_id: query.execution_id,
    category: query.category,
    limit: query.limit ?? DEFAULT_PAGE_LIMIT,
  };
}

/** First-page list (dashboard overview KPIs / recent activity). */
export function useEvents(query: EventsQuery = {}) {
  const workspaceId = useActiveWorkspaceId();
  const filters = listFilters(query);
  return useQuery({
    queryKey: queryKeys.events(
      workspaceId ?? "none",
      filters.execution_id,
      filters.category,
      filters.limit,
    ),
    queryFn: async () => {
      const page = await eventsApi.list(filters);
      return page.items;
    },
    enabled: Boolean(workspaceId),
  });
}

/** Keyset-paginated explorer (`Load more`). */
export function useInfiniteEvents(query: EventsQuery = {}) {
  const workspaceId = useActiveWorkspaceId();
  const filters = listFilters(query);
  return useInfiniteQuery({
    queryKey: [
      ...queryKeys.events(
        workspaceId ?? "none",
        filters.execution_id,
        filters.category,
        filters.limit,
      ),
      "infinite",
    ],
    queryFn: ({ pageParam }) =>
      eventsApi.list({
        ...filters,
        cursor: pageParam,
      }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (last) =>
      last.pagination.has_next ? last.pagination.next_cursor || undefined : undefined,
    enabled: Boolean(workspaceId),
  });
}
