"use client";

import { useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Label } from "@hublio/ui/ui/label";
import { Card, CardContent } from "@hublio/ui/ui/card";
import { authApi, isMFAChallenge } from "@/lib/api/auth";
import { ApiError } from "@/lib/api/client";
import { useAuth } from "@/providers/auth-provider";
import { PasswordField } from "@/features/auth/components/password-field";
import {
  makeLoginSchema,
  type LoginValues,
} from "@/features/auth/schemas";
import { Link, useRouter } from "@/i18n/navigation";
import { Logo } from "@/components/logo";

const DEMO_ADMIN =
  process.env.NODE_ENV === "development" ||
    process.env.NEXT_PUBLIC_DEMO_LOGIN === "true"
    ? { email: "admin@hublio.local", password: "Admin123!" }
    : null;

export function LoginForm() {
  const t = useTranslations("auth.login");
  const tv = useTranslations("validation");
  const { establishSession, logout } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [formError, setFormError] = useState<string | null>(null);

  const schema = useMemo(() => makeLoginSchema(tv), [tv]);
  const form = useForm<LoginValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", password: "" },
  });

  async function onSubmit(values: LoginValues) {
    setFormError(null);
    try {
      const data = await authApi.login(values.email.trim(), values.password);
      if (isMFAChallenge(data)) {
        setFormError(t("mfaUnsupported"));
        return;
      }
      const user = await establishSession(data);
      if (!user.is_platform_admin) {
        await logout();
        setFormError(t("notPlatformAdmin"));
        return;
      }
      const callback = searchParams.get("callbackUrl") || "/";
      router.replace(callback.startsWith("/") ? callback : "/");
    } catch (err) {
      const message =
        err instanceof ApiError
          ? err.message
          : err instanceof Error
            ? err.message
            : t("failed");
      setFormError(message);
    }
  }

  return (
    <Card className="w-full max-w-md">
      <CardContent className="pt-8">
        <div className="mb-7 flex flex-col items-center text-center">
          <Logo size="lg" priority />
          <h1 className="mt-5 text-[17px] font-semibold tracking-tight text-(--ink)">
            {t("title")}
          </h1>
          <p className="mt-1.5 text-[13px] text-muted-foreground">
            {t("subtitle")}
          </p>
        </div>

        <form
          className="space-y-4"
          onSubmit={form.handleSubmit((values) => void onSubmit(values))}
          noValidate
        >
          <div className="space-y-2">
            <Label htmlFor="admin-email">
              {t("email")}
              <span className="ml-0.5 text-destructive">*</span>
            </Label>
            <Input
              id="admin-email"
              type="email"
              autoComplete="username"
              aria-invalid={Boolean(form.formState.errors.email) || undefined}
              {...form.register("email")}
            />
            {form.formState.errors.email ? (
              <p className="text-sm text-destructive" role="alert">
                {form.formState.errors.email.message}
              </p>
            ) : null}
          </div>

          <PasswordField
            id="admin-password"
            label={t("password")}
            autoComplete="current-password"
            error={form.formState.errors.password?.message}
            {...form.register("password")}
          />

          <div className="flex justify-end">
            <Link
              href="/forgot-password"
              className="text-sm text-primary no-underline hover:underline"
            >
              {t("forgotPassword")}
            </Link>
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

          {DEMO_ADMIN ? (
            <Button
              type="button"
              variant="outline"
              className="w-full"
              onClick={() => {
                form.setValue("email", DEMO_ADMIN.email, {
                  shouldValidate: true,
                });
                form.setValue("password", DEMO_ADMIN.password, {
                  shouldValidate: true,
                });
              }}
            >
              {t("fillDemo")}
            </Button>
          ) : null}
        </form>
      </CardContent>
    </Card>
  );
}
