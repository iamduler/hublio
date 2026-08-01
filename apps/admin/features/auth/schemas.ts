import { z } from "zod";

export type ValidationTranslator = (key: string) => string;

export function makeLoginSchema(t: ValidationTranslator) {
  return z.object({
    email: z.string().min(1, t("required")).email(t("email")),
    password: z.string().min(1, t("required")),
  });
}

export function makeForgotPasswordSchema(t: ValidationTranslator) {
  return z.object({
    email: z.string().min(1, t("required")).email(t("email")),
  });
}

export function makeResetPasswordSchema(t: ValidationTranslator) {
  return z
    .object({
      password: z.string().min(8, t("passwordMin")),
      confirm_password: z.string().min(1, t("required")),
    })
    .refine((data) => data.password === data.confirm_password, {
      message: t("passwordsMatch"),
      path: ["confirm_password"],
    });
}

export type LoginValues = z.infer<ReturnType<typeof makeLoginSchema>>;
export type ForgotPasswordValues = z.infer<
  ReturnType<typeof makeForgotPasswordSchema>
>;
export type ResetPasswordValues = z.infer<
  ReturnType<typeof makeResetPasswordSchema>
>;
