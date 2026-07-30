"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { intentsApi, type IntentsQuery } from "./api";
import type { CreateIntentPayload } from "./types";

export function useIntents(query: IntentsQuery = {}) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.intents(
      workspaceId ?? "none",
      query.status,
      query.limit,
    ),
    queryFn: () => intentsApi.list(query),
    enabled: Boolean(workspaceId),
  });
}

export function useIntent(intentId: string | undefined) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.intent(workspaceId ?? "none", intentId ?? "none"),
    queryFn: () => intentsApi.get(intentId!),
    enabled: Boolean(workspaceId && intentId),
  });
}

export function useCreateIntent() {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateIntentPayload) =>
      intentsApi.create(payload, crypto.randomUUID()),
    onSuccess: () => {
      if (!workspaceId) return;
      void queryClient.invalidateQueries({
        queryKey: ["intents", workspaceId],
      });
      void queryClient.invalidateQueries({
        queryKey: ["executions", workspaceId],
      });
    },
  });
}
