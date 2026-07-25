import { z } from "zod";

export type ValidationTranslator = (key: string) => string;

export function makeInviteSchema(t: ValidationTranslator) {
  return z.object({
    email: z.string().min(1, t("required")).email(t("email")),
    role: z.enum(["owner", "admin", "member"]),
  });
}

export type InviteValues = z.infer<ReturnType<typeof makeInviteSchema>>;
