"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { executionsApi, type ExecutionsQuery } from "./api";

export function useExecutions(query: ExecutionsQuery = {}) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.executions(
      workspaceId ?? "none",
      query.status,
      query.limit,
    ),
    queryFn: () => executionsApi.list(query),
    enabled: Boolean(workspaceId),
  });
}

export function useExecution(executionId: string | undefined) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.execution(workspaceId ?? "none", executionId ?? "none"),
    queryFn: () => executionsApi.get(executionId!),
    enabled: Boolean(workspaceId && executionId),
  });
}

export function useExecutionTimeline(executionId: string | undefined) {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.executionTimeline(
      workspaceId ?? "none",
      executionId ?? "none",
    ),
    queryFn: () => executionsApi.timeline(executionId!),
    enabled: Boolean(workspaceId && executionId),
  });
}

export function useExecutionAction(executionId: string) {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (action: "cancel" | "retry") =>
      action === "cancel"
        ? executionsApi.cancel(executionId)
        : executionsApi.retry(executionId),
    onSuccess: () => {
      if (!workspaceId) return;
      void queryClient.invalidateQueries({
        queryKey: queryKeys.execution(workspaceId, executionId),
      });
      void queryClient.invalidateQueries({
        queryKey: queryKeys.executionTimeline(workspaceId, executionId),
      });
      void queryClient.invalidateQueries({
        queryKey: ["executions", workspaceId],
      });
    },
  });
}
