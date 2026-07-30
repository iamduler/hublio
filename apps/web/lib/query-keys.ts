/**
 * Central TanStack Query key factory.
 *
 * Convention: every workspace-scoped list/detail is keyed by the active
 * workspace id so switching workspaces transparently swaps cached data.
 */
export const queryKeys = {
  workspaces: (organizationId: string) =>
    ["workspaces", organizationId] as const,

  connectors: () => ["connectors"] as const,
  connector: (connectorId: string) => ["connectors", connectorId] as const,

  connections: (workspaceId: string) =>
    ["connections", workspaceId] as const,
  connection: (workspaceId: string, connectionId: string) =>
    ["connections", workspaceId, connectionId] as const,

  syncRoutes: (workspaceId: string) => ["sync-routes", workspaceId] as const,
  syncRoute: (workspaceId: string, syncRouteId: string) =>
    ["sync-routes", workspaceId, syncRouteId] as const,

  apiKeys: (workspaceId: string) => ["api-keys", workspaceId] as const,

  members: (workspaceId: string) => ["members", workspaceId] as const,

  intent: (workspaceId: string, intentId: string) =>
    ["intents", workspaceId, intentId] as const,
  execution: (workspaceId: string, executionId: string) =>
    ["executions", workspaceId, executionId] as const,
  executionTimeline: (workspaceId: string, executionId: string) =>
    ["executions", workspaceId, executionId, "timeline"] as const,

  events: (
    workspaceId: string,
    executionId?: string,
    category?: string,
    limit?: number,
  ) =>
    [
      "events",
      workspaceId,
      executionId ?? "all",
      category ?? "all",
      limit ?? "default",
    ] as const,

  mfaStatus: () => ["mfa", "status"] as const,
} as const;
