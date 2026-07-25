"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { apiKeysApi } from "./api";

export function useApiKeys() {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.apiKeys(workspaceId ?? "none"),
    queryFn: () => apiKeysApi.list(workspaceId!),
    enabled: Boolean(workspaceId),
  });
}

function useInvalidate() {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return () => {
    if (workspaceId) {
      void queryClient.invalidateQueries({
        queryKey: queryKeys.apiKeys(workspaceId),
      });
    }
  };
}

export function useCreateApiKey() {
  const workspaceId = useActiveWorkspaceId();
  const invalidate = useInvalidate();
  return useMutation({
    mutationFn: (name: string) => apiKeysApi.create(workspaceId!, name),
    onSuccess: invalidate,
  });
}

export function useDisableApiKey() {
  const workspaceId = useActiveWorkspaceId();
  const invalidate = useInvalidate();
  return useMutation({
    mutationFn: (id: string) => apiKeysApi.disable(workspaceId!, id),
    onSuccess: invalidate,
  });
}

export function useRotateApiKey() {
  const workspaceId = useActiveWorkspaceId();
  const invalidate = useInvalidate();
  return useMutation({
    mutationFn: (id: string) => apiKeysApi.rotate(workspaceId!, id),
    onSuccess: invalidate,
  });
}
