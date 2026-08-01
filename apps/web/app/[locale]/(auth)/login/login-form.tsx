"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useSearchParams } from "next/navigation";
import { Link, useRouter } from "@/i18n/navigation";
import { authApi, isMFAChallenge } from "@/lib/api/auth";
import { useAuth } from "@/providers/auth-provider";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { makeLoginSchema, type LoginValues } from "@/features/auth/schemas";
import {
  AuthCard,
  AuthFormHeader,
  AuthDivider,
  AuthFieldLabel,
  Logo,
  PasswordField,
  SSOButton,
  cx,
  inputBase,
} from "@/features/auth/auth-ui";
import { getOrCreateDeviceId, storeMFAToken } from "@/lib/mfa-device";
import { toast } from "@/lib/toast";

const DEMO_ACCOUNTS = [
  {
    id: "demo",
    email: "demo@hublio.local",
    password: "Demo123!",
  },
  {
    id: "admin",
    email: "admin@hublio.local",
    password: "Admin123!",
  },
] as const;

const showDemoLogin =
  process.env.NODE_ENV === "development" ||
  process.env.NEXT_PUBLIC_DEMO_LOGIN === "true";

type OAuthProvider = "google" | "microsoft" | "github";

export function LoginForm() {
  const t = useTranslations("auth.login");
  const tv = useTranslations("validation");
  const { establishSession } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const getError = useApiErrorMessage();
  const [providers, setProviders] = useState<OAuthProvider[]>([]);

  const schema = useMemo(() => makeLoginSchema(tv), [tv]);
  const form = useForm<LoginValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", password: "" },
  });

  useEffect(() => {
    const oauthError = searchParams.get("oauth_error");
    if (oauthError) {
      toast.error(oauthError);
    }
  }, [searchParams]);

  useEffect(() => {
    let cancelled = false;
    authApi
      .oauthProviders()
      .then((list) => {
        if (!cancelled) setProviders(list);
      })
      .catch(() => {
        if (!cancelled) setProviders([]);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  async function onSubmit(values: LoginValues) {
    try {
      const data = await authApi.login(
        values.email,
        values.password,
        getOrCreateDeviceId(),
      );
      if (isMFAChallenge(data)) {
        storeMFAToken(data.mfa_token);
        const callback = searchParams.get("callbackUrl") || "/dashboard";
        router.replace(
          `/mfa?callbackUrl=${encodeURIComponent(callback)}`,
        );
        return;
      }
      await establishSession(data);
      const callback = searchParams.get("callbackUrl") || "/dashboard";
      router.replace(callback);
    } catch (err) {
      toast.error(getError(err));
    }
  }

  function fillDemo(account: (typeof DEMO_ACCOUNTS)[number]) {
    form.setValue("email", account.email, { shouldValidate: true });
    form.setValue("password", account.password, { shouldValidate: true });
  }

  return (
    <AuthCard>
      <AuthFormHeader
        logo={<Logo size="lg" priority />}
        title={t("title")}
        subtitle={t("subtitle")}
      />

      {showDemoLogin && (
        <div className="mb-4 space-y-2">
          <p className="text-[11px] font-semibold uppercase tracking-wider text-(--faint)">
            {t("quickLogin")}
          </p>
          <div className="flex flex-wrap gap-2">
            {DEMO_ACCOUNTS.map((account) => (
              <button
                key={account.id}
                type="button"
                onClick={() => fillDemo(account)}
                className="rounded-md border border-(--line) bg-(--line-2) px-2.5 py-1 text-[12px] font-medium text-(--ink-2) hover:border-(--line-3) hover:bg-(--white)"
              >
                {t(`demo.${account.id}`)}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="mb-4 space-y-2.5">
        {(["google", "microsoft", "github"] as const).map((provider) => {
          const enabled = providers.includes(provider);
          return (
            <SSOButton
              key={provider}
              provider={provider}
              label={t(`oauth.${provider}`)}
              disabled={!enabled}
              onClick={() => {
                if (!enabled) {
                  toast.error(t("oauthUnavailable"));
                  return;
                }
                window.location.href = `/api/auth/oauth/start?provider=${provider}`;
              }}
            />
          );
        })}
      </div>

      <AuthDivider label={t("or")} />

      <form
        className="mt-1 space-y-4"
        onSubmit={form.handleSubmit(onSubmit)}
        noValidate
      >
        <div>
          <AuthFieldLabel htmlFor="email" required>
            {t("email")}
          </AuthFieldLabel>
          <input
            id="email"
            type="email"
            autoComplete="email"
            autoFocus
            placeholder="you@company.com"
            className={cx(
              inputBase,
              form.formState.errors.email
                ? "border-red-300 bg-red-50/40"
                : "border-(--line) hover:border-(--line-3)",
            )}
            {...form.register("email")}
          />
          {form.formState.errors.email && (
            <p className="mt-1 text-[12px] text-red-600">
              {form.formState.errors.email.message}
            </p>
          )}
        </div>

        <div>
          <div className="mb-1.5 flex items-center justify-between">
            <label
              htmlFor="password"
              className="text-[13px] font-medium text-(--ink-2)"
            >
              {t("password")} <span className="text-red-500">*</span>
            </label>
            <Link
              href="/forgot-password"
              className="text-[13px] font-medium text-blue-600 hover:text-blue-700 hover:underline"
            >
              {t("forgotPassword")}
            </Link>
          </div>
          <PasswordField
            id="password"
            autoComplete="current-password"
            error={form.formState.errors.password?.message}
            {...form.register("password")}
          />
        </div>

        <button
          type="submit"
          disabled={form.formState.isSubmitting}
          className="flex h-9 w-full items-center justify-center gap-2 rounded-lg bg-blue-600 text-[13px] font-medium text-white transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {form.formState.isSubmitting ? t("submitting") : t("submit")}
        </button>
      </form>

      <p className="mt-5 text-center text-[13px] text-(--muted-clr)">
        {t("noAccount")}{" "}
        <Link
          href="/register"
          className="font-medium text-blue-600 hover:text-blue-700 hover:underline"
        >
          {t("registerLink")}
        </Link>
      </p>
    </AuthCard>
  );
}
