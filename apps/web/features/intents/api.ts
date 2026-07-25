import { bff, unwrapData } from "@/lib/api/bff-client";
import type { SuccessEnvelope } from "@/lib/api/types";
import type { CreateIntentPayload, CreateIntentResult, Intent } from "./types";

/** Orchestration context via BFF (X-API-KEY held server-side). */
export const intentsApi = {
  create(payload: CreateIntentPayload, idempotencyKey?: string) {
    return bff
      .post<SuccessEnvelope<CreateIntentResult>>("/intents", payload, {
        ...(idempotencyKey ? { "Idempotency-Key": idempotencyKey } : {}),
      })
      .then((res) => unwrapData<CreateIntentResult>(res));
  },

  get(intentId: string) {
    return bff
      .get<SuccessEnvelope<Intent>>(`/intents/${intentId}`)
      .then((res) => unwrapData<Intent>(res));
  },
};
