export type CredentialType =
  | "api_key"
  | "oauth2"
  | "bearer_token"
  | "basic_auth"
  | "jwt";

export type ConnectionStatus =
  | "draft"
  | "verifying"
  | "active"
  | "verification_failed"
  | "disabled";

export interface Connection {
  id: string;
  workspace_id: string;
  connector_id: string;
  name: string;
  is_default?: boolean;
  description?: string;
  environment: string;
  status: ConnectionStatus;
  config?: Record<string, unknown>;
  timeout_seconds?: number;
  active_credential_id?: string;
  created_at?: string;
  updated_at?: string;
}

export interface CreateConnectionPayload {
  connector_id: string;
  name: string;
  environment: string;
  credential_type: CredentialType;
  secret: Record<string, string>;
  description?: string;
  is_default?: boolean;
  config?: Record<string, unknown>;
  timeout_seconds?: number;
}
