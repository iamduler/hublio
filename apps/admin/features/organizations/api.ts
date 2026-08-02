import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type { Organization } from "./types";

/** Identity — organizations (JWT; list requires platform admin). */
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
};
