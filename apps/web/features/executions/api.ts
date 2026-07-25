import { bff, unwrapData } from "@/lib/api/bff-client";
import type { SuccessEnvelope } from "@/lib/api/types";
import type { Execution, ExecutionTimeline } from "./types";

/** Orchestration context via BFF (X-API-KEY held server-side). */
export const executionsApi = {
  get(executionId: string) {
    return bff
      .get<SuccessEnvelope<Execution>>(`/executions/${executionId}`)
      .then((res) => unwrapData<Execution>(res));
  },

  timeline(executionId: string) {
    return bff
      .get<SuccessEnvelope<ExecutionTimeline>>(
        `/executions/${executionId}/timeline`,
      )
      .then((res) => unwrapData<ExecutionTimeline>(res));
  },

  cancel(executionId: string) {
    return bff
      .post<SuccessEnvelope<Execution>>(
        `/executions/${executionId}/cancel`,
        {},
      )
      .then((res) => unwrapData<Execution>(res));
  },

  retry(executionId: string) {
    return bff
      .post<SuccessEnvelope<Execution>>(
        `/executions/${executionId}/retry`,
        {},
      )
      .then((res) => unwrapData<Execution>(res));
  },
};
