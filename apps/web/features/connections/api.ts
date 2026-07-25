import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type { Connection, CreateConnectionPayload } from "./types";

const base = (workspaceId: string) =>
  `/integration/workspaces/${workspaceId}/connections`;

/** Integration context — connections (JWT, workspace-scoped). */
export const connectionsApi = {
  list(workspaceId: string) {
    return api
      .get<SuccessEnvelope<Connection[]>>(base(workspaceId))
      .then((res) => unwrapData<Connection[]>(res));
  },

  get(workspaceId: string, connectionId: string) {
    return api
      .get<SuccessEnvelope<Connection>>(`${base(workspaceId)}/${connectionId}`)
      .then((res) => unwrapData<Connection>(res));
  },

  create(workspaceId: string, payload: CreateConnectionPayload) {
    return api
      .post<SuccessEnvelope<Connection>>(base(workspaceId), payload)
      .then((res) => unwrapData<Connection>(res));
  },

  verify(workspaceId: string, connectionId: string) {
    return api
      .post<SuccessEnvelope<Connection>>(
        `${base(workspaceId)}/${connectionId}/verify`,
        {},
      )
      .then((res) => unwrapData<Connection>(res));
  },

  enable(workspaceId: string, connectionId: string) {
    return api
      .post<SuccessEnvelope<Connection>>(
        `${base(workspaceId)}/${connectionId}/enable`,
        {},
      )
      .then((res) => unwrapData<Connection>(res));
  },

  disable(workspaceId: string, connectionId: string) {
    return api
      .post<SuccessEnvelope<Connection>>(
        `${base(workspaceId)}/${connectionId}/disable`,
        {},
      )
      .then((res) => unwrapData<Connection>(res));
  },

  rotateCredentials(workspaceId: string, connectionId: string) {
    return api
      .post<SuccessEnvelope<Connection>>(
        `${base(workspaceId)}/${connectionId}/credentials/rotate`,
        {},
      )
      .then((res) => unwrapData<Connection>(res));
  },
};
