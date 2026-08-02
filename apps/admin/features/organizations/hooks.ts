"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { organizationsApi } from "./api";

export function useOrganizations() {
  return useQuery({
    queryKey: queryKeys.organizations.all,
    queryFn: () => organizationsApi.list(),
  });
}

export function useOrganization(organizationId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.organizations.detail(organizationId ?? "none"),
    queryFn: () => organizationsApi.get(organizationId!),
    enabled: Boolean(organizationId),
  });
}

export function useSuspendOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (organizationId: string) =>
      organizationsApi.suspend(organizationId),
    onSuccess: (data) => {
      void queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.all,
      });
      void queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.detail(data.id),
      });
    },
  });
}

export function useActivateOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (organizationId: string) =>
      organizationsApi.activate(organizationId),
    onSuccess: (data) => {
      void queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.all,
      });
      void queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.detail(data.id),
      });
    },
  });
}
