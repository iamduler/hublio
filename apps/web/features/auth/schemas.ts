import { z } from "zod";

export type ValidationTranslator = (key: string) => string;

export function makeLoginSchema(t: ValidationTranslator) {
  return z.object({
    email: z.string().min(1, t("required")).email(t("email")),
    password: z.string().min(1, t("required")),
  });
}

export function makeRegisterSchema(t: ValidationTranslator) {
  return z.object({
    organization_name: z.string().min(1, t("required")),
    full_name: z.string().min(1, t("required")),
    email: z.string().min(1, t("required")).email(t("email")),
    password: z.string().min(8, t("passwordMin")),
  });
}

export type LoginValues = z.infer<ReturnType<typeof makeLoginSchema>>;
export type RegisterValues = z.infer<ReturnType<typeof makeRegisterSchema>>;
