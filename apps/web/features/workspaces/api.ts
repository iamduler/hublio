import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type {
  CreateWorkspacePayload,
  Organization,
  Workspace,
} from "./types";

/** Identity context endpoints (JWT-authenticated, org-scoped). */
export const workspacesApi = {
  listByOrganization(organizationId: string) {
    return api
      .get<SuccessEnvelope<Workspace[]>>(
        `/identity/organizations/${organizationId}/workspaces`,
      )
      .then((res) => unwrapData<Workspace[]>(res));
  },

  create(organizationId: string, payload: CreateWorkspacePayload) {
    return api
      .post<SuccessEnvelope<Workspace>>(
        `/identity/organizations/${organizationId}/workspaces`,
        payload,
      )
      .then((res) => unwrapData<Workspace>(res));
  },

  getOrganization(organizationId: string) {
    return api
      .get<SuccessEnvelope<Organization>>(
        `/identity/organizations/${organizationId}`,
      )
      .then((res) => unwrapData<Organization>(res));
  },

  enable(workspaceId: string) {
    return api
      .post<SuccessEnvelope<Workspace>>(
        `/identity/workspaces/${workspaceId}/enable`,
        {},
      )
      .then((res) => unwrapData<Workspace>(res));
  },

  disable(workspaceId: string) {
    return api
      .post<SuccessEnvelope<Workspace>>(
        `/identity/workspaces/${workspaceId}/disable`,
        {},
      )
      .then((res) => unwrapData<Workspace>(res));
  },
};
