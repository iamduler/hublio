"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { ChevronLeft, Lock } from "lucide-react";
import { Link, useRouter } from "@/i18n/navigation";
import { authApi } from "@/lib/api/auth";
import { useAuth } from "@/providers/auth-provider";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  AuthCard,
  AuthFieldLabel,
  CheckboxField,
  OTPInput,
  cx,
  inputBase,
} from "@/features/auth/auth-ui";
import {
  clearMFAToken,
  getOrCreateDeviceId,
  readMFAToken,
} from "@/lib/mfa-device";
import { toast } from "@/lib/toast";

export function MFAForm() {
  const t = useTranslations("auth.mfa");
  const getError = useApiErrorMessage();
  const { establishSession } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();

  const [code, setCode] = useState(["", "", "", "", "", ""]);
  const [recoveryCode, setRecoveryCode] = useState("");
  const [useRecovery, setUseRecovery] = useState(false);
  const [trustDevice, setTrustDevice] = useState(false);
  const [loading, setLoading] = useState(false);
  const [mfaToken, setMfaToken] = useState<string | null>(null);

  useEffect(() => {
    const token = readMFAToken();
    if (!token) {
      router.replace("/login");
      return;
    }
    setMfaToken(token);
  }, [router]);

  async function handleVerify() {
    if (!mfaToken) {
      router.replace("/login");
      return;
    }
    if (!useRecovery && code.join("").length < 6) {
      toast.error(t("invalidCode"));
      return;
    }
    if (useRecovery && recoveryCode.trim().length < 8) {
      toast.error(t("invalidCode"));
      return;
    }

    setLoading(true);
    try {
      const data = await authApi.verifyMFA({
        mfa_token: mfaToken,
        ...(useRecovery
          ? { recovery_code: recoveryCode.trim() }
          : { code: code.join("") }),
        trust_device: trustDevice,
        device_id: getOrCreateDeviceId(),
      });
      clearMFAToken();
      await establishSession(data);
      const callback = searchParams.get("callbackUrl") || "/dashboard";
      router.replace(callback);
    } catch (err) {
      toast.error(getError(err));
      setCode(["", "", "", "", "", ""]);
      setRecoveryCode("");
    } finally {
      setLoading(false);
    }
  }

  if (!mfaToken) {
    return (
      <AuthCard>
        <div className="py-8 text-center text-sm text-(--muted-clr)">…</div>
      </AuthCard>
    );
  }

  return (
    <AuthCard>
      <Link
        href="/login"
        className="mb-7 -ml-0.5 flex items-center gap-1.5 text-[13px] text-(--muted-clr) transition-colors hover:text-(--ink-2)"
      >
        <ChevronLeft size={14} />
        {t("back")}
      </Link>

      <div className="mb-6">
        <div className="mb-4 flex h-9 w-9 items-center justify-center rounded-xl border border-violet-100 bg-violet-50">
          <Lock size={16} className="text-violet-600" />
        </div>
        <h1 className="text-[17px] font-semibold tracking-tight text-(--ink)">
          {t("title")}
        </h1>
        <p className="mt-1.5 text-[13px] leading-relaxed text-(--muted-clr)">
          {useRecovery ? t("recoverySubtitle") : t("subtitle")}
        </p>
      </div>

      <div className="space-y-5">
        {!useRecovery ? (
          <OTPInput value={code} onChange={setCode} />
        ) : (
          <div>
            <AuthFieldLabel htmlFor="recovery" required>
              {t("recoveryLabel")}
            </AuthFieldLabel>
            <input
              id="recovery"
              type="text"
              autoFocus
              value={recoveryCode}
              onChange={(e) => setRecoveryCode(e.target.value)}
              placeholder={t("recoveryPlaceholder")}
              className={cx(
                inputBase,
                "border-(--line) font-mono tracking-wider hover:border-(--line-3)",
              )}
            />
          </div>
        )}

        <CheckboxField
          id="mfa-remember"
          checked={trustDevice}
          onChange={setTrustDevice}
          label={t("trustDevice")}
        />

        <button
          type="button"
          onClick={() => void handleVerify()}
          disabled={loading}
          className="flex h-9 w-full items-center justify-center gap-2 rounded-lg bg-blue-600 text-[13px] font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
        >
          {loading ? t("submitting") : t("submit")}
        </button>
      </div>

      <div className="mt-5 text-center">
        <button
          type="button"
          onClick={() => {
            setUseRecovery((v) => !v);
            setCode(["", "", "", "", "", ""]);
            setRecoveryCode("");
          }}
          className="text-[13px] font-medium text-blue-600 transition-colors hover:text-blue-700"
        >
          {useRecovery ? t("useAuthenticator") : t("useRecovery")}
        </button>
      </div>
    </AuthCard>
  );
}
