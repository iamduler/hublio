"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { intentsApi } from "./api";
import type { CreateIntentPayload } from "./types";

export function useIntent(intentId: string | undefined) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.intent(workspaceId ?? "none", intentId ?? "none"),
    queryFn: () => intentsApi.get(intentId!),
    enabled: Boolean(workspaceId && intentId),
  });
}

export function useCreateIntent() {
  return useMutation({
    mutationFn: (payload: CreateIntentPayload) =>
      intentsApi.create(payload, crypto.randomUUID()),
  });
}
