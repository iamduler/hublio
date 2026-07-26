"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Link, useRouter } from "@/i18n/navigation";
import { authApi } from "@/lib/api/auth";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  makeRegisterSchema,
  type RegisterValues,
} from "@/features/auth/schemas";
import {
  AuthCard,
  AuthDivider,
  AuthFieldLabel,
  CheckboxField,
  Logo,
  PasswordField,
  PasswordRequirements,
  PasswordStrengthMeter,
  SSOButton,
  cx,
  inputBase,
} from "@/features/auth/auth-ui";
import { toast } from "@/lib/toast";

type OAuthProvider = "google" | "microsoft" | "github";

export default function RegisterPage() {
  const t = useTranslations("auth.register");
  const tv = useTranslations("validation");
  const router = useRouter();
  const getError = useApiErrorMessage();
  const [providers, setProviders] = useState<OAuthProvider[]>([]);

  const schema = useMemo(() => makeRegisterSchema(tv), [tv]);
  const form = useForm<RegisterValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      organization_name: "",
      full_name: "",
      email: "",
      password: "",
      confirm_password: "",
      terms: false,
    },
  });

  const password = form.watch("password");

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

  async function onSubmit(values: RegisterValues) {
    try {
      await authApi.register({
        organization_name: values.organization_name,
        full_name: values.full_name,
        email: values.email,
        password: values.password,
      });
      toast.success(t("success"));
      router.replace(
        `/verify-email?email=${encodeURIComponent(values.email.trim().toLowerCase())}`,
      );
    } catch (err) {
      toast.error(getError(err));
    }
  }

  return (
    <AuthCard>
      <div className="mb-6 flex flex-col items-center">
        <Logo size="lg" />
        <div className="mt-5 text-center">
          <h1 className="text-[17px] font-semibold tracking-tight text-[var(--ink)]">
            {t("title")}
          </h1>
          <p className="mt-1.5 text-[13px] text-[var(--muted-clr)]">{t("subtitle")}</p>
        </div>
      </div>

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
        className="mt-1 space-y-3.5"
        onSubmit={form.handleSubmit(onSubmit)}
        noValidate
      >
        <div>
          <AuthFieldLabel htmlFor="full_name" required>
            {t("fullName")}
          </AuthFieldLabel>
          <input
            id="full_name"
            type="text"
            autoComplete="name"
            autoFocus
            placeholder={t("fullNamePlaceholder")}
            className={cx(
              inputBase,
              form.formState.errors.full_name
                ? "border-red-300 bg-red-50/40"
                : "border-[var(--line)] hover:border-[var(--line-3)]",
            )}
            {...form.register("full_name")}
          />
          {form.formState.errors.full_name && (
            <p className="mt-1 text-[12px] text-red-600">
              {form.formState.errors.full_name.message}
            </p>
          )}
        </div>

        <div>
          <AuthFieldLabel htmlFor="organization_name" required>
            {t("organizationName")}
          </AuthFieldLabel>
          <input
            id="organization_name"
            type="text"
            autoComplete="organization"
            placeholder={t("organizationPlaceholder")}
            className={cx(
              inputBase,
              form.formState.errors.organization_name
                ? "border-red-300 bg-red-50/40"
                : "border-[var(--line)] hover:border-[var(--line-3)]",
            )}
            {...form.register("organization_name")}
          />
          {form.formState.errors.organization_name && (
            <p className="mt-1 text-[12px] text-red-600">
              {form.formState.errors.organization_name.message}
            </p>
          )}
        </div>

        <div>
          <AuthFieldLabel htmlFor="email" required>
            {t("email")}
          </AuthFieldLabel>
          <input
            id="email"
            type="email"
            autoComplete="email"
            placeholder="you@company.com"
            className={cx(
              inputBase,
              form.formState.errors.email
                ? "border-red-300 bg-red-50/40"
                : "border-[var(--line)] hover:border-[var(--line-3)]",
            )}
            {...form.register("email")}
          />
          {form.formState.errors.email && (
            <p className="mt-1 text-[12px] text-red-600">
              {form.formState.errors.email.message}
            </p>
          )}
        </div>

        <PasswordField
          id="signup-pw"
          label={t("password")}
          autoComplete="new-password"
          error={form.formState.errors.password?.message}
          {...form.register("password")}
        />
        {password ? <PasswordStrengthMeter password={password} /> : null}
        {password ? (
          <PasswordRequirements
            password={password}
            labels={{
              minLength: t("passwordRules.minLength"),
              uppercase: t("passwordRules.uppercase"),
              lowercase: t("passwordRules.lowercase"),
              number: t("passwordRules.number"),
              special: t("passwordRules.special"),
            }}
          />
        ) : null}

        <PasswordField
          id="signup-confirm"
          label={t("confirmPassword")}
          autoComplete="new-password"
          error={form.formState.errors.confirm_password?.message}
          {...form.register("confirm_password")}
        />

        <CheckboxField
          id="terms"
          checked={form.watch("terms") ?? false}
          onChange={(v) => form.setValue("terms", v, { shouldValidate: true })}
          error={form.formState.errors.terms?.message}
          label={
            <span>
              {t("termsPrefix")}{" "}
              <span className="font-medium text-blue-600">{t("termsOfService")}</span>{" "}
              {t("termsAnd")}{" "}
              <span className="font-medium text-blue-600">{t("privacyPolicy")}</span>
            </span>
          }
        />

        <div className="pt-0.5">
          <button
            type="submit"
            disabled={form.formState.isSubmitting}
            className="flex h-9 w-full items-center justify-center gap-2 rounded-lg bg-blue-600 text-[13px] font-medium text-white transition-colors hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {form.formState.isSubmitting ? t("submitting") : t("submit")}
          </button>
        </div>
      </form>

      <p className="mt-5 text-center text-[13px] text-[var(--muted-clr)]">
        {t("hasAccount")}{" "}
        <Link
          href="/login"
          className="font-medium text-blue-600 hover:text-blue-700 hover:underline"
        >
          {t("loginLink")}
        </Link>
      </p>
    </AuthCard>
  );
}
