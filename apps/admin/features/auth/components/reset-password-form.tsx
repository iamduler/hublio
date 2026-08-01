"use client";

import { useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { CheckCircle2 } from "lucide-react";
import { Button } from "@hublio/ui/ui/button";
import { AuthCard, AuthFormHeader } from "@hublio/ui/common/auth-card";
import { Logo } from "@hublio/ui/common/logo";
import { authApi } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";
import { PasswordField } from "@/features/auth/components/password-field";
import {
  makeResetPasswordSchema,
  type ResetPasswordValues,
} from "@/features/auth/schemas";
import { Link, useRouter } from "@/i18n/navigation";

export function ResetPasswordForm() {
  const t = useTranslations("auth.reset");
  const tv = useTranslations("validation");
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";
  const missingToken = !token.trim();

  const [done, setDone] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const schema = useMemo(() => makeResetPasswordSchema(tv), [tv]);
  const form = useForm<ResetPasswordValues>({
    resolver: zodResolver(schema),
    defaultValues: { password: "", confirm_password: "" },
  });

  async function onSubmit(values: ResetPasswordValues) {
    setFormError(null);
    if (missingToken) {
      setFormError(t("invalidToken"));
      return;
    }
    try {
      await authApi.resetPassword(token, values.password);
      setDone(true);
    } catch (err) {
      setFormError(
        err instanceof ApiError
          ? err.message
          : err instanceof Error
            ? err.message
            : t("failed"),
      );
    }
  }

  if (done) {
    return (
      <AuthCard contentClassName="space-y-4 py-8 text-center">
        <div className="flex justify-center">
          <Logo size="md" />
        </div>
        <div className="mx-auto flex size-12 items-center justify-center rounded-full bg-green-50 text-green-600">
          <CheckCircle2 className="size-6" />
        </div>
        <h2 className="text-[17px] font-semibold tracking-tight text-(--ink)">
          {t("doneTitle")}
        </h2>
        <p className="text-[13px] text-muted-foreground">{t("doneBody")}</p>
        <Button
          type="button"
          className="w-full"
          onClick={() => router.replace("/login")}
        >
          {t("continueLogin")}
        </Button>
      </AuthCard>
    );
  }

  return (
    <AuthCard>
      <AuthFormHeader
        logo={<Logo size="lg" priority />}
        title={t("title")}
        subtitle={t("subtitle")}
      />
      {missingToken ? (
        <p className="mb-4 text-center text-sm text-destructive" role="alert">
          {t("invalidToken")}
        </p>
      ) : null}

      <form
        className="space-y-4"
        onSubmit={form.handleSubmit((values) => void onSubmit(values))}
        noValidate
      >
        <PasswordField
          id="new-password"
          label={t("password")}
          autoComplete="new-password"
          error={form.formState.errors.password?.message}
          disabled={missingToken}
          {...form.register("password")}
        />
        <PasswordField
          id="confirm-password"
          label={t("confirmPassword")}
          autoComplete="new-password"
          error={form.formState.errors.confirm_password?.message}
          disabled={missingToken}
          {...form.register("confirm_password")}
        />

        {formError ? (
          <p className="text-sm text-destructive" role="alert">
            {formError}
          </p>
        ) : null}

        <Button
          type="submit"
          className="w-full"
          disabled={form.formState.isSubmitting || missingToken}
        >
          {form.formState.isSubmitting ? t("submitting") : t("submit")}
        </Button>

        <Link
          href="/login"
          className="block text-center text-sm text-muted-foreground no-underline hover:underline"
        >
          {t("backToLogin")}
        </Link>
      </form>
    </AuthCard>
  );
}
