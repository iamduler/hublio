/**
 * Central TanStack Query key factory.
 *
 * Convention: every workspace-scoped list is keyed by the active workspace id
 * so switching workspaces transparently swaps cached data.
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

  intent: (intentId: string) => ["intents", intentId] as const,
  execution: (executionId: string) => ["executions", executionId] as const,
  executionTimeline: (executionId: string) =>
    ["executions", executionId, "timeline"] as const,

  events: (workspaceId: string, executionId?: string) =>
    ["events", workspaceId, executionId ?? "all"] as const,
} as const;
