export type WorkspaceEnvironment = "production" | "staging" | "development";

export type WorkspaceStatus = "active" | "disabled";

export interface Organization {
  id: string;
  name: string;
  status: "active" | "suspended" | "archived";
  created_at?: string;
  updated_at?: string;
}

export interface Workspace {
  id: string;
  organization_id: string;
  name: string;
  environment: string;
  status: WorkspaceStatus;
  created_at?: string;
  updated_at?: string;
}

export interface CreateWorkspacePayload {
  name: string;
  environment: string;
}
