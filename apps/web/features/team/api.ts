import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type { AddMemberPayload, WorkspaceMember } from "./types";

/** Identity context — workspace membership (JWT). */
export const teamApi = {
  addMember(workspaceId: string, payload: AddMemberPayload) {
    return api
      .post<SuccessEnvelope<WorkspaceMember>>(
        `/identity/workspaces/${workspaceId}/members`,
        payload,
      )
      .then((res) => unwrapData<WorkspaceMember>(res));
  },
};
