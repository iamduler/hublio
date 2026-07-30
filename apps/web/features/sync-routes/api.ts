import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type {
  CreateSyncRoutePayload,
  SyncRoute,
  SyncRouteWatermark,
  UpdateSyncRoutePayload,
} from "./types";

const base = (workspaceId: string) =>
  `/integration/workspaces/${workspaceId}/sync-routes`;

/** Integration context — sync routes (JWT, workspace-scoped). */
export const syncRoutesApi = {
  list(workspaceId: string) {
    return api
      .get<SuccessEnvelope<SyncRoute[]>>(base(workspaceId))
      .then((res) => unwrapData<SyncRoute[]>(res));
  },

  get(workspaceId: string, syncRouteId: string) {
    return api
      .get<SuccessEnvelope<SyncRoute>>(`${base(workspaceId)}/${syncRouteId}`)
      .then((res) => unwrapData<SyncRoute>(res));
  },

  create(workspaceId: string, payload: CreateSyncRoutePayload) {
    return api
      .post<SuccessEnvelope<SyncRoute>>(base(workspaceId), payload)
      .then((res) => unwrapData<SyncRoute>(res));
  },

  update(
    workspaceId: string,
    syncRouteId: string,
    payload: UpdateSyncRoutePayload,
  ) {
    return api
      .patch<SuccessEnvelope<SyncRoute>>(
        `${base(workspaceId)}/${syncRouteId}`,
        payload,
      )
      .then((res) => unwrapData<SyncRoute>(res));
  },

  enable(workspaceId: string, syncRouteId: string) {
    return api
      .post<SuccessEnvelope<SyncRoute>>(
        `${base(workspaceId)}/${syncRouteId}/enable`,
        {},
      )
      .then((res) => unwrapData<SyncRoute>(res));
  },

  disable(workspaceId: string, syncRouteId: string) {
    return api
      .post<SuccessEnvelope<SyncRoute>>(
        `${base(workspaceId)}/${syncRouteId}/disable`,
        {},
      )
      .then((res) => unwrapData<SyncRoute>(res));
  },

  remove(workspaceId: string, syncRouteId: string): Promise<null> {
    return api
      .del<SuccessEnvelope<null>>(`${base(workspaceId)}/${syncRouteId}`)
      .then(() => null);
  },

  rotateWebhookSecret(workspaceId: string, syncRouteId: string) {
    return api
      .post<SuccessEnvelope<SyncRoute>>(
        `${base(workspaceId)}/${syncRouteId}/webhook-secret/rotate`,
        {},
      )
      .then((res) => unwrapData<SyncRoute>(res));
  },

  watermarks(workspaceId: string, syncRouteId: string) {
    return api
      .get<SuccessEnvelope<SyncRouteWatermark[]>>(
        `${base(workspaceId)}/${syncRouteId}/watermarks`,
      )
      .then((res) => unwrapData<SyncRouteWatermark[]>(res));
  },

  upsertWatermark(
    workspaceId: string,
    syncRouteId: string,
    resourceType: string,
    cursor: Record<string, unknown>,
  ) {
    return api
      .put<SuccessEnvelope<SyncRouteWatermark>>(
        `${base(workspaceId)}/${syncRouteId}/watermarks/${encodeURIComponent(resourceType)}`,
        { cursor },
      )
      .then((res) => unwrapData<SyncRouteWatermark>(res));
  },

  /** Orchestration poll tick — JWT + X-Workspace-ID via /api/go. */
  poll(syncRouteId: string, resourceType: string) {
    return api
      .post<SuccessEnvelope<unknown>>(`/sync-routes/${syncRouteId}/poll`, {
        resource_type: resourceType,
      })
      .then((res) => unwrapData(res));
  },
};
