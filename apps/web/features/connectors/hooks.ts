"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { connectorsApi } from "./api";

export function useConnectors() {
  return useQuery({
    queryKey: queryKeys.connectors(),
    queryFn: () => connectorsApi.list(),
  });
}

export function useConnector(connectorId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.connector(connectorId ?? "none"),
    queryFn: () => connectorsApi.get(connectorId!),
    enabled: Boolean(connectorId),
  });
}

export function useToggleConnector() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, enable }: { id: string; enable: boolean }) =>
      enable ? connectorsApi.enable(id) : connectorsApi.disable(id),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.connectors() });
      void queryClient.invalidateQueries({
        queryKey: queryKeys.connector(variables.id),
      });
    },
  });
}
