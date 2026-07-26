"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { CheckCircle2, ChevronLeft, KeyRound, RefreshCw } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { authApi } from "@/lib/api/auth";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  AuthCard,
  AuthFieldLabel,
  cx,
  inputBase,
} from "@/features/auth/auth-ui";
import { toast } from "@/lib/toast";

export default function ForgotPasswordPage() {
  const t = useTranslations("auth.forgot");
  const getError = useApiErrorMessage();
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!email.trim()) {
      setError(t("email"));
      return;
    }
    if (!/\S+@\S+\.\S+/.test(email)) {
      setError(t("email"));
      return;
    }
    setLoading(true);
    setError("");
    try {
      await authApi.forgotPassword(email.trim().toLowerCase());
      setSent(true);
      toast.success(t("successGeneric"));
    } catch (err) {
      toast.error(getError(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthCard>
      <Link
        href="/login"
        className="mb-7 -ml-0.5 flex items-center gap-1.5 text-[13px] text-[var(--muted-clr)] transition-colors hover:text-[var(--ink-2)]"
      >
        <ChevronLeft size={14} />
        {t("backToLogin")}
      </Link>

      {!sent ? (
        <>
          <div className="mb-6">
            <div className="mb-4 flex h-9 w-9 items-center justify-center rounded-xl border border-[var(--line)] bg-[var(--line-2)]">
              <KeyRound size={16} className="text-[var(--ink-2)]" />
            </div>
            <h1 className="text-[17px] font-semibold tracking-tight text-[var(--ink)]">
              {t("title")}
            </h1>
            <p className="mt-1.5 text-[13px] leading-relaxed text-[var(--muted-clr)]">
              {t("subtitle")}
            </p>
          </div>
          <form className="space-y-4" onSubmit={(e) => void handleSubmit(e)} noValidate>
            <div>
              <AuthFieldLabel htmlFor="forgot-email" required>
                {t("email")}
              </AuthFieldLabel>
              <input
                id="forgot-email"
                type="email"
                autoFocus
                autoComplete="email"
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value);
                  setError("");
                }}
                className={cx(
                  inputBase,
                  error
                    ? "border-red-300 bg-red-50/40"
                    : "border-[var(--line)] hover:border-[var(--line-3)]",
                )}
              />
              {error ? <p className="mt-1 text-[12px] text-red-600">{error}</p> : null}
            </div>
            <button
              type="submit"
              disabled={loading}
              className="flex h-9 w-full items-center justify-center gap-2 rounded-lg bg-blue-600 text-[13px] font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? t("submitting") : t("submit")}
            </button>
          </form>
        </>
      ) : (
        <div className="py-4 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border border-green-100 bg-green-50">
            <CheckCircle2 size={22} className="text-green-600" />
          </div>
          <h2 className="text-[16px] font-semibold text-[var(--ink)]">{t("sentTitle")}</h2>
          <p className="mt-2 text-[13px] leading-relaxed text-[var(--muted-clr)]">
            {t("sentBody", { email })}
          </p>
          <div className="mt-5 space-y-2.5">
            <button
              type="button"
              onClick={() => setSent(false)}
              className="flex h-9 w-full items-center justify-center gap-2 rounded-lg border border-[var(--line)] bg-[var(--white)] text-[13px] font-medium text-[var(--ink-2)] transition-colors hover:bg-[var(--line-2)]"
            >
              <RefreshCw size={12} />
              {t("resend")}
            </button>
            <Link
              href="/login"
              className="block w-full text-center text-[13px] text-[var(--muted-clr)] transition-colors hover:text-[var(--ink-2)]"
            >
              {t("backToLogin")}
            </Link>
          </div>
        </div>
      )}
    </AuthCard>
  );
}
