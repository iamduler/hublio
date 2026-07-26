import type { Schemas } from "@/lib/api/sdk";

export type ConnectorCategory =
  Schemas["RegisterConnectorRequest"]["category"];

export type ConnectorStatus =
  | "registered"
  | "enabled"
  | "disabled"
  | "removed";

export interface ConnectorCapability {
  id?: string;
  capability_code?: string;
  code?: string;
  display_name: string;
  status?: string;
  is_async?: boolean;
}

/** Response entity — not yet a named OpenAPI schema. */
export interface Connector {
  id: string;
  code: string;
  name: string;
  vendor: string;
  category: ConnectorCategory;
  version: string;
  status: ConnectorStatus;
  description?: string;
  homepage?: string;
  documentation_url?: string;
  capabilities?: ConnectorCapability[];
  created_at?: string;
  updated_at?: string;
}

export type RegisterConnectorPayload = Schemas["RegisterConnectorRequest"];
