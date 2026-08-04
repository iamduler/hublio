export const queryKeys = {
  organizations: {
    all: ["organizations"] as const,
    detail: (id: string) => ["organizations", id] as const,
    workspaces: (id: string) => ["organizations", id, "workspaces"] as const,
    users: (id: string) => ["organizations", id, "users"] as const,
  },
};
