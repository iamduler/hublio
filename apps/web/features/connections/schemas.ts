import { z } from "zod";
import type { CredentialType } from "./types";

export type ValidationTranslator = (key: string) => string;

/** Secret fields required per credential type. */
export const SECRET_FIELDS: Record<CredentialType, string[]> = {
  api_key: ["api_key"],
  bearer_token: ["token"],
  jwt: ["token"],
  basic_auth: ["username", "password"],
  oauth2: ["client_id", "client_secret"],
};

export function makeConnectionSchema(t: ValidationTranslator) {
  return z.object({
    connector_id: z.string().min(1, t("required")),
    name: z.string().min(1, t("required")),
    environment: z.string().min(1, t("required")),
    credential_type: z.enum([
      "api_key",
      "oauth2",
      "bearer_token",
      "basic_auth",
      "jwt",
    ]),
    description: z.string().optional(),
    secret: z.record(z.string(), z.string()),
  });
}

export type ConnectionFormValues = z.infer<
  ReturnType<typeof makeConnectionSchema>
>;
