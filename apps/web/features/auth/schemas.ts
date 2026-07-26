import { z } from "zod";

export type ValidationTranslator = (key: string) => string;

export function makeLoginSchema(t: ValidationTranslator) {
  return z.object({
    email: z.string().min(1, t("required")).email(t("email")),
    password: z.string().min(1, t("required")),
  });
}

export function makeRegisterSchema(t: ValidationTranslator) {
  return z
    .object({
      organization_name: z.string().min(1, t("required")),
      full_name: z.string().min(1, t("required")),
      email: z.string().min(1, t("required")).email(t("email")),
      password: z.string().min(8, t("passwordMin")),
      confirm_password: z.string().min(1, t("required")),
      terms: z.boolean().refine((v) => v === true, {
        message: t("termsRequired"),
      }),
    })
    .refine((data) => data.password === data.confirm_password, {
      message: t("passwordsMatch"),
      path: ["confirm_password"],
    });
}

export type LoginValues = z.infer<ReturnType<typeof makeLoginSchema>>;
export type RegisterValues = z.infer<ReturnType<typeof makeRegisterSchema>>;
