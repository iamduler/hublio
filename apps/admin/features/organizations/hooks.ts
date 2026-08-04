"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { organizationsApi } from "./api";
import type { CreateOrganizationPayload, CreateWorkspacePayload } from "./types";

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

export function useOrganizationWorkspaces(organizationId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.organizations.workspaces(organizationId ?? "none"),
    queryFn: () => organizationsApi.listWorkspaces(organizationId!),
    enabled: Boolean(organizationId),
  });
}

export function useOrganizationUsers(organizationId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.organizations.users(organizationId ?? "none"),
    queryFn: () => organizationsApi.listUsers(organizationId!),
    enabled: Boolean(organizationId),
  });
}

function invalidateOrg(queryClient: ReturnType<typeof useQueryClient>, id: string) {
  void queryClient.invalidateQueries({
    queryKey: queryKeys.organizations.all,
  });
  void queryClient.invalidateQueries({
    queryKey: queryKeys.organizations.detail(id),
  });
}

export function useCreateOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateOrganizationPayload) =>
      organizationsApi.create(payload),
    onSuccess: (data) => {
      invalidateOrg(queryClient, data.organization.id);
    },
  });
}

export function useUpdateOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      organizationId,
      name,
    }: {
      organizationId: string;
      name: string;
    }) => organizationsApi.update(organizationId, name),
    onSuccess: (data) => {
      invalidateOrg(queryClient, data.id);
    },
  });
}

export function useSuspendOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (organizationId: string) =>
      organizationsApi.suspend(organizationId),
    onSuccess: (data) => {
      invalidateOrg(queryClient, data.id);
    },
  });
}

export function useActivateOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (organizationId: string) =>
      organizationsApi.activate(organizationId),
    onSuccess: (data) => {
      invalidateOrg(queryClient, data.id);
    },
  });
}

export function useArchiveOrganization() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (organizationId: string) =>
      organizationsApi.archive(organizationId),
    onSuccess: (data) => {
      invalidateOrg(queryClient, data.id);
      void queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.workspaces(data.id),
      });
    },
  });
}

export function useCreateWorkspace(organizationId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: CreateWorkspacePayload) =>
      organizationsApi.createWorkspace(organizationId!, payload),
    onSuccess: () => {
      if (!organizationId) return;
      void queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.workspaces(organizationId),
      });
    },
  });
}

export function useEnableWorkspace(organizationId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (workspaceId: string) =>
      organizationsApi.enableWorkspace(workspaceId),
    onSuccess: () => {
      if (!organizationId) return;
      void queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.workspaces(organizationId),
      });
    },
  });
}

export function useDisableWorkspace(organizationId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (workspaceId: string) =>
      organizationsApi.disableWorkspace(workspaceId),
    onSuccess: () => {
      if (!organizationId) return;
      void queryClient.invalidateQueries({
        queryKey: queryKeys.organizations.workspaces(organizationId),
      });
    },
  });
}
