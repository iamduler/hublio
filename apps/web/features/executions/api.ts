import { api, unwrapData, type SuccessEnvelope } from "@/lib/api/client";
import type { Execution, ExecutionTimeline } from "./types";

export type ExecutionsQuery = {
  status?: string;
  limit?: number;
};

/** Orchestration via JWT proxy (`/api/go` → Go Bearer + X-Workspace-ID). */
export const executionsApi = {
  list(query: ExecutionsQuery = {}) {
    return api
      .get<SuccessEnvelope<Execution[]>>("/executions", {
        status: query.status,
        limit: query.limit,
      })
      .then((res) => unwrapData<Execution[]>(res));
  },

  get(executionId: string) {
    return api
      .get<SuccessEnvelope<Execution>>(`/executions/${executionId}`)
      .then((res) => unwrapData<Execution>(res));
  },

  timeline(executionId: string) {
    return api
      .get<SuccessEnvelope<ExecutionTimeline>>(
        `/executions/${executionId}/timeline`,
      )
      .then((res) => unwrapData<ExecutionTimeline>(res));
  },

  cancel(executionId: string) {
    return api
      .post<SuccessEnvelope<Execution>>(
        `/executions/${executionId}/cancel`,
        {},
      )
      .then((res) => unwrapData<Execution>(res));
  },

  retry(executionId: string) {
    return api
      .post<SuccessEnvelope<Execution>>(
        `/executions/${executionId}/retry`,
        {},
      )
      .then((res) => unwrapData<Execution>(res));
  },
};
