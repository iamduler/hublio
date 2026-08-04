import { z } from "zod";

type TFunc = (key: string) => string;

export function makeCreateOrganizationSchema(t: TFunc) {
  return z.object({
    name: z.string().min(1, t("required")),
  });
}

export function makeRenameOrganizationSchema(t: TFunc) {
  return z.object({
    name: z.string().min(1, t("required")),
  });
}

export function makeCreateWorkspaceSchema(t: TFunc) {
  return z.object({
    name: z.string().min(1, t("required")),
    environment: z.enum(["production", "staging", "development"]),
  });
}

export type CreateOrganizationValues = z.infer<
  ReturnType<typeof makeCreateOrganizationSchema>
>;
export type RenameOrganizationValues = z.infer<
  ReturnType<typeof makeRenameOrganizationSchema>
>;
export type CreateWorkspaceValues = z.infer<
  ReturnType<typeof makeCreateWorkspaceSchema>
>;
