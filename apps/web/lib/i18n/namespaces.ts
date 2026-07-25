export const FEATURE_NAMESPACES = [
  "common",
  "auth",
  "validation",
  "errors",
  "dashboard",
  "workspaces",
  "connectors",
  "connections",
  "syncRoutes",
  "apiKeys",
  "team",
  "intents",
  "executions",
  "events",
] as const;

export type FeatureNamespace = (typeof FEATURE_NAMESPACES)[number];

export type MessageNamespace = FeatureNamespace;
