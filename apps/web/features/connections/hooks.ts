"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { connectionsApi } from "./api";
import type { CreateConnectionPayload } from "./types";

export function useConnections() {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.connections(workspaceId ?? "none"),
    queryFn: () => connectionsApi.list(workspaceId!),
    enabled: Boolean(workspaceId),
  });
}

export function useConnection(connectionId: string | undefined) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.connection(workspaceId ?? "none", connectionId ?? "none"),
    queryFn: () => connectionsApi.get(workspaceId!, connectionId!),
    enabled: Boolean(workspaceId && connectionId),
  });
}

export function useCreateConnection() {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateConnectionPayload) =>
      connectionsApi.create(workspaceId!, payload),
    onSuccess: () => {
      if (workspaceId) {
        void queryClient.invalidateQueries({
          queryKey: queryKeys.connections(workspaceId),
        });
      }
    },
  });
}

type Action = "verify" | "enable" | "disable" | "rotate";

export function useConnectionAction() {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, action }: { id: string; action: Action }) => {
      switch (action) {
        case "verify":
          return connectionsApi.verify(workspaceId!, id);
        case "enable":
          return connectionsApi.enable(workspaceId!, id);
        case "disable":
          return connectionsApi.disable(workspaceId!, id);
        case "rotate":
          return connectionsApi.rotateCredentials(workspaceId!, id);
      }
    },
    onSuccess: (_data, variables) => {
      if (!workspaceId) return;
      void queryClient.invalidateQueries({
        queryKey: queryKeys.connections(workspaceId),
      });
      void queryClient.invalidateQueries({
        queryKey: queryKeys.connection(workspaceId, variables.id),
      });
    },
  });
}
