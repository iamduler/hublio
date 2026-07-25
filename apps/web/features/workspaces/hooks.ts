"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { workspacesApi } from "./api";
import type { CreateWorkspacePayload } from "./types";

export function useWorkspaces(organizationId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.workspaces(organizationId ?? "none"),
    queryFn: () => workspacesApi.listByOrganization(organizationId!),
    enabled: Boolean(organizationId),
  });
}

export function useCreateWorkspace(organizationId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateWorkspacePayload) =>
      workspacesApi.create(organizationId!, payload),
    onSuccess: () => {
      if (organizationId) {
        void queryClient.invalidateQueries({
          queryKey: queryKeys.workspaces(organizationId),
        });
      }
    },
  });
}
