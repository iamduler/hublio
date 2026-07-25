"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { executionsApi } from "./api";

export function useExecution(executionId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.execution(executionId ?? "none"),
    queryFn: () => executionsApi.get(executionId!),
    enabled: Boolean(executionId),
  });
}

export function useExecutionAction(executionId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (action: "cancel" | "retry") =>
      action === "cancel"
        ? executionsApi.cancel(executionId)
        : executionsApi.retry(executionId),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: queryKeys.execution(executionId),
      });
    },
  });
}
