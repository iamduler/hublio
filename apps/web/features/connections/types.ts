import type { Schemas } from "@/lib/api/sdk";

export type CredentialType = Schemas["CreateConnectionRequest"]["credential_type"];

export type ConnectionStatus =
  | "draft"
  | "verifying"
  | "active"
  | "verification_failed"
  | "disabled";

/** Response entity — not yet a named OpenAPI schema. */
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

export type CreateConnectionPayload = Omit<
  Schemas["CreateConnectionRequest"],
  "config" | "retry_policy" | "secret" | "is_default" | "timeout_seconds"
> & {
  is_default?: boolean;
  timeout_seconds?: number;
  config?: Record<string, unknown>;
  secret: Record<string, string>;
};

export type RotateCredentialPayload = Omit<
  Schemas["RotateCredentialRequest"],
  "secret"
> & {
  secret: Record<string, string>;
};
