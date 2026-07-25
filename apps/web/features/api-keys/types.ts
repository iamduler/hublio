export type ApiKeyStatus = "active" | "disabled";

export interface ApiKey {
  id: string;
  workspace_id: string;
  name: string;
  prefix?: string;
  status: ApiKeyStatus;
  last_used_at?: string;
  expires_at?: string;
  created_at?: string;
}

/** Returned once on create/rotate — plaintext is never persisted. */
export interface ApiKeySecret {
  api_key: ApiKey;
  plaintext: string;
  warning?: string;
}
