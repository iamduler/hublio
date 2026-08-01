"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { CheckCircle2, RefreshCw } from "lucide-react";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Label } from "@hublio/ui/ui/label";
import { AuthCard, AuthFormHeader } from "@hublio/ui/common/auth-card";
import { Logo } from "@hublio/ui/common/logo";
import { authApi } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";
import {
  makeForgotPasswordSchema,
  type ForgotPasswordValues,
} from "@/features/auth/schemas";
import { Link } from "@/i18n/navigation";

export function ForgotPasswordForm() {
  const t = useTranslations("auth.forgot");
  const tv = useTranslations("validation");
  const [sentEmail, setSentEmail] = useState<string | null>(null);
  const [formError, setFormError] = useState<string | null>(null);

  const schema = useMemo(() => makeForgotPasswordSchema(tv), [tv]);
  const form = useForm<ForgotPasswordValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: "" },
  });

  async function onSubmit(values: ForgotPasswordValues) {
    setFormError(null);
    try {
      const email = values.email.trim().toLowerCase();
      await authApi.forgotPassword(email);
      setSentEmail(email);
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

  if (sentEmail) {
    return (
      <AuthCard contentClassName="space-y-4 py-8 text-center">
        <div className="flex justify-center">
          <Logo size="md" />
        </div>
        <div className="mx-auto flex size-12 items-center justify-center rounded-full bg-green-50 text-green-600">
          <CheckCircle2 className="size-6" />
        </div>
        <h2 className="text-[17px] font-semibold tracking-tight text-(--ink)">
          {t("sentTitle")}
        </h2>
        <p className="text-[13px] text-muted-foreground">
          {t("sentBody", { email: sentEmail })}
        </p>
        <p className="text-xs text-muted-foreground">{t("successGeneric")}</p>
        <Button
          type="button"
          variant="outline"
          className="w-full"
          onClick={() => {
            setSentEmail(null);
            form.reset({ email: sentEmail });
          }}
        >
          <RefreshCw size={14} />
          {t("resend")}
        </Button>
        <Link
          href="/login"
          className="block text-sm text-muted-foreground no-underline hover:underline"
        >
          {t("backToLogin")}
        </Link>
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

      <form
        className="space-y-4"
        onSubmit={form.handleSubmit((values) => void onSubmit(values))}
        noValidate
      >
        <div className="space-y-2">
          <Label htmlFor="forgot-email">
            {t("email")}
            <span className="ml-0.5 text-destructive">*</span>
          </Label>
          <Input
            id="forgot-email"
            type="email"
            autoComplete="email"
            autoFocus
            aria-invalid={Boolean(form.formState.errors.email) || undefined}
            {...form.register("email")}
          />
          {form.formState.errors.email ? (
            <p className="text-sm text-destructive" role="alert">
              {form.formState.errors.email.message}
            </p>
          ) : null}
        </div>

        {formError ? (
          <p className="text-sm text-destructive" role="alert">
            {formError}
          </p>
        ) : null}

        <Button
          type="submit"
          className="w-full"
          disabled={form.formState.isSubmitting}
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
