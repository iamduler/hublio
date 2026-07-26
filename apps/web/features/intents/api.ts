import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type { CreateIntentPayload, CreateIntentResult, Intent } from "./types";

/** Orchestration via JWT proxy (`/api/go` → Go Bearer + X-Workspace-ID). */
export const intentsApi = {
  create(payload: CreateIntentPayload, idempotencyKey?: string) {
    return api
      .post<SuccessEnvelope<CreateIntentResult>>("/intents", payload, {
        headers: idempotencyKey
          ? { "Idempotency-Key": idempotencyKey }
          : undefined,
      })
      .then((res) => unwrapData<CreateIntentResult>(res));
  },

  get(intentId: string) {
    return api
      .get<SuccessEnvelope<Intent>>(`/intents/${intentId}`)
      .then((res) => unwrapData<Intent>(res));
  },
};
