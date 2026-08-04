export type Organization = {
  id: string;
  name: string;
  status: "active" | "suspended" | "archived" | string;
  created_at?: string;
  updated_at?: string;
};

export type Workspace = {
  id: string;
  organization_id: string;
  name: string;
  environment: string;
  status: "active" | "disabled" | string;
  created_at?: string;
  updated_at?: string;
};

export type CreateOrganizationPayload = {
  name: string;
  workspace_name?: string;
  environment?: string;
};

export type CreateOrganizationResult = {
  organization: Organization;
  workspace: Workspace;
};

export type CreateWorkspacePayload = {
  name: string;
  environment: string;
};

export type OrganizationUser = {
  id: string;
  organization_id: string;
  email: string;
  full_name: string;
  status: string;
  is_platform_admin?: boolean;
  created_at?: string;
  updated_at?: string;
  last_login_at?: string | null;
};
