import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type { ApiKey, ApiKeySecret } from "./types";

const base = (workspaceId: string) =>
  `/identity/workspaces/${workspaceId}/api-keys`;

/** Identity context — workspace API keys (JWT). */
export const apiKeysApi = {
  list(workspaceId: string) {
    return api
      .get<SuccessEnvelope<ApiKey[]>>(base(workspaceId))
      .then((res) => unwrapData<ApiKey[]>(res));
  },

  create(workspaceId: string, name: string) {
    return api
      .post<SuccessEnvelope<ApiKeySecret>>(base(workspaceId), { name })
      .then((res) => unwrapData<ApiKeySecret>(res));
  },

  disable(workspaceId: string, apiKeyId: string) {
    return api
      .post<SuccessEnvelope<ApiKey>>(
        `${base(workspaceId)}/${apiKeyId}/disable`,
        {},
      )
      .then((res) => unwrapData<ApiKey>(res));
  },

  rotate(workspaceId: string, apiKeyId: string) {
    return api
      .post<SuccessEnvelope<ApiKeySecret>>(
        `${base(workspaceId)}/${apiKeyId}/rotate`,
        {},
      )
      .then((res) => unwrapData<ApiKeySecret>(res));
  },
};
