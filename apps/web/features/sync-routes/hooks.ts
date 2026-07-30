"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { syncRoutesApi } from "./api";
import type {
  CreateSyncRoutePayload,
  UpdateSyncRoutePayload,
} from "./types";

export function useSyncRoutes() {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.syncRoutes(workspaceId ?? "none"),
    queryFn: () => syncRoutesApi.list(workspaceId!),
    enabled: Boolean(workspaceId),
  });
}

export function useSyncRoute(syncRouteId: string | undefined) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.syncRoute(workspaceId ?? "none", syncRouteId ?? "none"),
    queryFn: () => syncRoutesApi.get(workspaceId!, syncRouteId!),
    enabled: Boolean(workspaceId && syncRouteId),
  });
}

export function useCreateSyncRoute() {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateSyncRoutePayload) =>
      syncRoutesApi.create(workspaceId!, payload),
    onSuccess: () => {
      if (workspaceId) {
        void queryClient.invalidateQueries({
          queryKey: queryKeys.syncRoutes(workspaceId),
        });
      }
    },
  });
}

export function useUpdateSyncRoute(syncRouteId: string) {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: UpdateSyncRoutePayload) =>
      syncRoutesApi.update(workspaceId!, syncRouteId, payload),
    onSuccess: () => {
      if (!workspaceId) return;
      void queryClient.invalidateQueries({
        queryKey: queryKeys.syncRoutes(workspaceId),
      });
      void queryClient.invalidateQueries({
        queryKey: queryKeys.syncRoute(workspaceId, syncRouteId),
      });
    },
  });
}

type Action = "enable" | "disable" | "remove" | "rotate";

export function useSyncRouteAction() {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, action }: { id: string; action: Action }) => {
      switch (action) {
        case "enable":
          return syncRoutesApi.enable(workspaceId!, id);
        case "disable":
          return syncRoutesApi.disable(workspaceId!, id);
        case "remove":
          return syncRoutesApi.remove(workspaceId!, id);
        case "rotate":
          return syncRoutesApi.rotateWebhookSecret(workspaceId!, id);
      }
    },
    onSuccess: (_data, variables) => {
      if (!workspaceId) return;
      void queryClient.invalidateQueries({
        queryKey: queryKeys.syncRoutes(workspaceId),
      });
      void queryClient.invalidateQueries({
        queryKey: queryKeys.syncRoute(workspaceId, variables.id),
      });
    },
  });
}

export function usePollSyncRoute(syncRouteId: string) {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (resourceType: string) =>
      syncRoutesApi.poll(syncRouteId, resourceType),
    onSuccess: () => {
      if (!workspaceId) return;
      void queryClient.invalidateQueries({
        queryKey: [
          ...queryKeys.syncRoute(workspaceId, syncRouteId),
          "watermarks",
        ],
      });
      void queryClient.invalidateQueries({
        queryKey: queryKeys.events(workspaceId),
      });
    },
  });
}

export function useUpsertSyncRouteWatermark(syncRouteId: string) {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      resourceType,
      cursor,
    }: {
      resourceType: string;
      cursor: Record<string, unknown>;
    }) =>
      syncRoutesApi.upsertWatermark(
        workspaceId!,
        syncRouteId,
        resourceType,
        cursor,
      ),
    onSuccess: () => {
      if (!workspaceId) return;
      void queryClient.invalidateQueries({
        queryKey: [
          ...queryKeys.syncRoute(workspaceId, syncRouteId),
          "watermarks",
        ],
      });
    },
  });
}
