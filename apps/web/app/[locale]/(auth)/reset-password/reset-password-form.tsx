"use client";

import { useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { CheckCircle2, Lock } from "lucide-react";
import { Link, useRouter } from "@/i18n/navigation";
import { authApi } from "@/lib/api/auth";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  AuthCard,
  PasswordField,
  PasswordRequirements,
  PasswordStrengthMeter,
} from "@/features/auth/auth-ui";
import { toast } from "@/lib/toast";

export default function ResetPasswordForm() {
  const t = useTranslations("auth.reset");
  const tr = useTranslations("auth.register");
  const getError = useApiErrorMessage();
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";

  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [errors, setErrors] = useState<{ password?: string; confirm?: string }>(
    {},
  );
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);

  const missingToken = useMemo(() => !token.trim(), [token]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const next: typeof errors = {};
    if (!password) next.password = tr("password");
    else if (password.length < 8) next.password = tr("passwordRules.minLength");
    if (password !== confirm) next.confirm = tr("confirmPassword");
    if (Object.keys(next).length) {
      setErrors(next);
      return;
    }
    if (missingToken) {
      toast.error(t("invalidToken"));
      return;
    }
    setLoading(true);
    try {
      await authApi.resetPassword(token, password);
      setDone(true);
    } catch (err) {
      toast.error(getError(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthCard>
      {!done ? (
        <>
          <div className="mb-6">
            <div className="mb-4 flex h-9 w-9 items-center justify-center rounded-xl border border-(--line) bg-(--line-2)">
              <Lock size={16} className="text-(--ink-2)" />
            </div>
            <h1 className="text-[17px] font-semibold tracking-tight text-(--ink)">
              {t("title")}
            </h1>
            <p className="mt-1.5 text-[13px] text-(--muted-clr)">{t("subtitle")}</p>
            {missingToken ? (
              <p className="mt-2 text-[12px] text-red-600">{t("invalidToken")}</p>
            ) : null}
          </div>
          <form className="space-y-4" onSubmit={(e) => void handleSubmit(e)} noValidate>
            <div>
              <PasswordField
                id="new-password"
                label={t("password")}
                autoComplete="new-password"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value);
                  setErrors((p) => ({ ...p, password: undefined }));
                }}
                error={errors.password}
              />
              {password ? <PasswordStrengthMeter password={password} /> : null}
              {password ? (
                <PasswordRequirements
                  password={password}
                  labels={{
                    minLength: tr("passwordRules.minLength"),
                    uppercase: tr("passwordRules.uppercase"),
                    lowercase: tr("passwordRules.lowercase"),
                    number: tr("passwordRules.number"),
                    special: tr("passwordRules.special"),
                  }}
                />
              ) : null}
            </div>
            <PasswordField
              id="confirm-password"
              label={t("confirmPassword")}
              autoComplete="new-password"
              value={confirm}
              onChange={(e) => {
                setConfirm(e.target.value);
                setErrors((p) => ({ ...p, confirm: undefined }));
              }}
              error={errors.confirm}
            />
            <button
              type="submit"
              disabled={loading || missingToken}
              className="flex h-9 w-full items-center justify-center gap-2 rounded-lg bg-blue-600 text-[13px] font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? t("submitting") : t("submit")}
            </button>
          </form>
        </>
      ) : (
        <div className="py-6 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border border-green-100 bg-green-50">
            <CheckCircle2 size={22} className="text-green-600" />
          </div>
          <h2 className="text-[16px] font-semibold text-(--ink)">{t("doneTitle")}</h2>
          <p className="mt-2 text-[13px] leading-relaxed text-(--muted-clr)">
            {t("doneBody")}
          </p>
          <div className="mt-5">
            <button
              type="button"
              onClick={() => router.replace("/login")}
              className="flex h-9 w-full items-center justify-center rounded-lg bg-blue-600 text-[13px] font-medium text-white transition-colors hover:bg-blue-700"
            >
              {t("continueLogin")}
            </button>
          </div>
          <Link
            href="/login"
            className="mt-3 block text-[13px] text-(--muted-clr) hover:text-(--ink-2)"
          >
            {t("continueLogin")}
          </Link>
        </div>
      )}
    </AuthCard>
  );
}
