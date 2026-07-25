"use client";

import { useMutation } from "@tanstack/react-query";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { teamApi } from "./api";
import type { AddMemberPayload } from "./types";

export function useAddMember() {
  const workspaceId = useActiveWorkspaceId();
  return useMutation({
    mutationFn: (payload: AddMemberPayload) =>
      teamApi.addMember(workspaceId!, payload),
  });
}
