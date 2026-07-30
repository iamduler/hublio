"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { teamApi } from "./api";
import type { AddMemberPayload } from "./types";

export function useMembers() {
  const workspaceId = useActiveWorkspaceId();
  return useQuery({
    queryKey: queryKeys.members(workspaceId ?? "none"),
    queryFn: () => teamApi.list(workspaceId!),
    enabled: Boolean(workspaceId),
  });
}

export function useAddMember() {
  const workspaceId = useActiveWorkspaceId();
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: AddMemberPayload) =>
      teamApi.addMember(workspaceId!, payload),
    onSuccess: () => {
      if (!workspaceId) return;
      void queryClient.invalidateQueries({
        queryKey: queryKeys.members(workspaceId),
      });
    },
  });
}
