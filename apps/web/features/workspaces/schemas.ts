import { z } from "zod";

type TFunc = (key: string) => string;

export function makeWorkspaceSchema(t: TFunc) {
  return z.object({
    name: z.string().min(1, t("required")),
    environment: z.enum(["production", "staging", "development"]),
  });
}

export type WorkspaceFormValues = z.infer<
  ReturnType<typeof makeWorkspaceSchema>
>;
