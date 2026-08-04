import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type {
  CreateOrganizationPayload,
  CreateOrganizationResult,
  CreateWorkspacePayload,
  Organization,
  OrganizationUser,
  Workspace,
} from "./types";

/** Identity — organizations (JWT; list/create require platform admin). */
export const organizationsApi = {
  list() {
    return api
      .get<SuccessEnvelope<Organization[]>>("/identity/organizations")
      .then((res) => unwrapData<Organization[]>(res));
  },

  get(organizationId: string) {
    return api
      .get<SuccessEnvelope<Organization>>(
        `/identity/organizations/${organizationId}`,
      )
      .then((res) => unwrapData<Organization>(res));
  },

  create(payload: CreateOrganizationPayload) {
    return api
      .post<SuccessEnvelope<CreateOrganizationResult>>(
        "/identity/organizations",
        payload,
      )
      .then((res) => unwrapData<CreateOrganizationResult>(res));
  },

  update(organizationId: string, name: string) {
    return api
      .patch<SuccessEnvelope<Organization>>(
        `/identity/organizations/${organizationId}`,
        { name },
      )
      .then((res) => unwrapData<Organization>(res));
  },

  suspend(organizationId: string) {
    return api
      .post<SuccessEnvelope<Organization>>(
        `/identity/organizations/${organizationId}/suspend`,
        {},
      )
      .then((res) => unwrapData<Organization>(res));
  },

  activate(organizationId: string) {
    return api
      .post<SuccessEnvelope<Organization>>(
        `/identity/organizations/${organizationId}/activate`,
        {},
      )
      .then((res) => unwrapData<Organization>(res));
  },

  archive(organizationId: string) {
    return api
      .post<SuccessEnvelope<Organization>>(
        `/identity/organizations/${organizationId}/archive`,
        {},
      )
      .then((res) => unwrapData<Organization>(res));
  },

  listWorkspaces(organizationId: string) {
    return api
      .get<SuccessEnvelope<Workspace[]>>(
        `/identity/organizations/${organizationId}/workspaces`,
      )
      .then((res) => unwrapData<Workspace[]>(res));
  },

  createWorkspace(organizationId: string, payload: CreateWorkspacePayload) {
    return api
      .post<SuccessEnvelope<Workspace>>(
        `/identity/organizations/${organizationId}/workspaces`,
        payload,
      )
      .then((res) => unwrapData<Workspace>(res));
  },

  enableWorkspace(workspaceId: string) {
    return api
      .post<SuccessEnvelope<Workspace>>(
        `/identity/workspaces/${workspaceId}/enable`,
        {},
      )
      .then((res) => unwrapData<Workspace>(res));
  },

  disableWorkspace(workspaceId: string) {
    return api
      .post<SuccessEnvelope<Workspace>>(
        `/identity/workspaces/${workspaceId}/disable`,
        {},
      )
      .then((res) => unwrapData<Workspace>(res));
  },

  listUsers(organizationId: string) {
    return api
      .get<SuccessEnvelope<OrganizationUser[]>>(
        `/identity/organizations/${organizationId}/users`,
      )
      .then((res) => unwrapData<OrganizationUser[]>(res));
  },
};
