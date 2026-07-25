import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type { Connector } from "./types";

/** Integration context — connector catalog (JWT). */
export const connectorsApi = {
  list() {
    return api
      .get<SuccessEnvelope<Connector[]>>("/integration/connectors")
      .then((res) => unwrapData<Connector[]>(res));
  },

  get(connectorId: string) {
    return api
      .get<SuccessEnvelope<Connector>>(
        `/integration/connectors/${connectorId}`,
      )
      .then((res) => unwrapData<Connector>(res));
  },

  enable(connectorId: string) {
    return api
      .post<SuccessEnvelope<Connector>>(
        `/integration/connectors/${connectorId}/enable`,
        {},
      )
      .then((res) => unwrapData<Connector>(res));
  },

  disable(connectorId: string) {
    return api
      .post<SuccessEnvelope<Connector>>(
        `/integration/connectors/${connectorId}/disable`,
        {},
      )
      .then((res) => unwrapData<Connector>(res));
  },
};
