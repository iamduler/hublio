"use client";

import { useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { CheckCircle2, ChevronLeft, Mail } from "lucide-react";
import { Link, useRouter } from "@/i18n/navigation";
import { authApi } from "@/lib/api/auth";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { AuthCard, OTPInput } from "@/features/auth/auth-ui";
import { toast } from "@/lib/toast";

export default function VerifyEmailPage() {
  const t = useTranslations("auth.verify");
  const getError = useApiErrorMessage();
  const router = useRouter();
  const searchParams = useSearchParams();
  const email = useMemo(
    () => (searchParams.get("email") ?? "").trim().toLowerCase(),
    [searchParams],
  );

  const [code, setCode] = useState(["", "", "", "", "", ""]);
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);
  const [resent, setResent] = useState(false);

  async function handleVerify() {
    if (code.join("").length < 6) {
      toast.error(t("codeRequired"));
      return;
    }
    if (!email) {
      toast.error(t("changeEmail"));
      router.replace("/register");
      return;
    }
    setLoading(true);
    try {
      await authApi.verifyEmail(email, code.join(""));
      toast.success(t("submit"));
      router.replace("/login");
    } catch (err) {
      toast.error(getError(err));
      setCode(["", "", "", "", "", ""]);
    } finally {
      setLoading(false);
    }
  }

  async function handleResend() {
    if (!email) return;
    setResending(true);
    try {
      await authApi.requestEmailVerification(email);
      setResent(true);
      toast.success(t("resent"));
      window.setTimeout(() => setResent(false), 4000);
    } catch (err) {
      toast.error(getError(err));
    } finally {
      setResending(false);
    }
  }

  return (
    <AuthCard>
      <Link
        href="/register"
        className="mb-7 -ml-0.5 flex items-center gap-1.5 text-[13px] text-[var(--muted-clr)] transition-colors hover:text-[var(--ink-2)]"
      >
        <ChevronLeft size={14} />
        {t("back")}
      </Link>

      <div className="mb-6">
        <div className="mb-4 flex h-10 w-10 items-center justify-center rounded-xl border border-blue-100 bg-blue-50">
          <Mail size={18} className="text-blue-600" />
        </div>
        <h1 className="text-[17px] font-semibold tracking-tight text-[var(--ink)]">
          {t("title")}
        </h1>
        <p className="mt-2 text-[13px] leading-relaxed text-[var(--muted-clr)]">
          {t("subtitle", { email: email || "—" })}
        </p>
      </div>

      <div className="space-y-5">
        <OTPInput value={code} onChange={setCode} />
        <button
          type="button"
          onClick={() => void handleVerify()}
          disabled={loading}
          className="flex h-9 w-full items-center justify-center gap-2 rounded-lg bg-blue-600 text-[13px] font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? t("submitting") : t("submit")}
        </button>
      </div>

      <div className="mt-5 space-y-2 text-center">
        {resent ? (
          <div className="flex items-center justify-center gap-1.5 text-[13px] text-green-700">
            <CheckCircle2 size={13} />
            <span>{t("resent")}</span>
          </div>
        ) : (
          <p className="text-[13px] text-[var(--muted-clr)]">
            {t("didntReceive")}{" "}
            <button
              type="button"
              onClick={() => void handleResend()}
              disabled={resending || !email}
              className="font-medium text-blue-600 transition-colors hover:text-blue-700 disabled:opacity-50"
            >
              {resending ? t("resending") : t("resend")}
            </button>
          </p>
        )}
        <Link
          href="/register"
          className="block w-full text-center text-[12px] text-[var(--faint)] transition-colors hover:text-[var(--ink-2)]"
        >
          {t("changeEmail")}
        </Link>
      </div>
    </AuthCard>
  );
}
