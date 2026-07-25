export type WorkspaceRole = "owner" | "admin" | "member";

export interface AddMemberPayload {
  email: string;
  role: WorkspaceRole;
}

export interface WorkspaceMember {
  user_id?: string;
  email: string;
  role: WorkspaceRole;
  created_at?: string;
}
