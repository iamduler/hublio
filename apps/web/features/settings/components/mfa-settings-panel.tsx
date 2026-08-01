"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Check,
  Copy,
  KeyRound,
  Loader2,
  Lock,
  Shield,
  ShieldCheck,
  ShieldOff,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { Button } from "@hublio/ui/ui/button";
import { ConfirmDialog } from "@hublio/ui/common/confirm-dialog";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { OTPInput, PasswordField, cx } from "@/features/auth/auth-ui";
import { authApi, type MFASetupData } from "@/lib/api/auth";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { queryKeys } from "@/lib/query-keys";
import { toast } from "@/lib/toast";

type Phase = "idle" | "setup" | "disable";

export function MFASettingsPanel() {
  const t = useTranslations("workspaces.mfa");
  const getError = useApiErrorMessage();
  const queryClient = useQueryClient();
  const [phase, setPhase] = useState<Phase>("idle");
  const [setup, setSetup] = useState<MFASetupData | null>(null);
  const [code, setCode] = useState(["", "", "", "", "", ""]);
  const [password, setPassword] = useState("");
  const [copiedSecret, setCopiedSecret] = useState(false);
  const [copiedCodes, setCopiedCodes] = useState(false);

  const statusQuery = useQuery({
    queryKey: queryKeys.mfaStatus(),
    queryFn: () => authApi.mfaStatus(),
  });

  const startSetup = useMutation({
    mutationFn: () => authApi.mfaSetup(),
    onSuccess: (data) => {
      setSetup(data);
      setCode(["", "", "", "", "", ""]);
      setPhase("setup");
    },
  });

  const enable = useMutation({
    mutationFn: (totp: string) => authApi.mfaEnable(totp),
    onSuccess: async () => {
      toast.success(t("enabled"));
      setPhase("idle");
      setSetup(null);
      setCode(["", "", "", "", "", ""]);
      await queryClient.invalidateQueries({ queryKey: queryKeys.mfaStatus() });
    },
  });

  const disable = useMutation({
    mutationFn: (pw: string) => authApi.mfaDisable(pw),
    onSuccess: async () => {
      toast.success(t("disabled"));
      setPhase("idle");
      setPassword("");
      await queryClient.invalidateQueries({ queryKey: queryKeys.mfaStatus() });
    },
  });

  if (statusQuery.isLoading) return <LoadingState rows={3} />;
  if (statusQuery.isError) {
    return (
      <ErrorState
        title={t("loadError")}
        description={getError(statusQuery.error)}
        onRetry={() => void statusQuery.refetch()}
      />
    );
  }

  const status = statusQuery.data!;

  async function onStartEnroll() {
    try {
      await startSetup.mutateAsync();
    } catch (err) {
      toast.error(getError(err));
    }
  }

  async function onConfirmEnable() {
    const totp = code.join("");
    if (totp.length < 6) {
      toast.error(t("codeRequired"));
      return;
    }
    try {
      await enable.mutateAsync(totp);
    } catch (err) {
      toast.error(getError(err));
      setCode(["", "", "", "", "", ""]);
    }
  }

  async function onConfirmDisable() {
    if (!password.trim()) {
      toast.error(t("passwordRequired"));
      return;
    }
    try {
      await disable.mutateAsync(password);
    } catch (err) {
      toast.error(getError(err));
    }
  }

  function copyText(value: string, kind: "secret" | "codes") {
    void navigator.clipboard.writeText(value).then(() => {
      if (kind === "secret") {
        setCopiedSecret(true);
        window.setTimeout(() => setCopiedSecret(false), 2000);
      } else {
        setCopiedCodes(true);
        window.setTimeout(() => setCopiedCodes(false), 2000);
      }
      toast.success(t("copied"));
    });
  }

  if (phase === "setup" && setup) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield size={16} className="text-blue-600" />
            {t("enrollTitle")}
          </CardTitle>
          <p className="text-sm text-(--muted-clr)">{t("enrollSubtitle")}</p>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="flex flex-col items-start gap-4 sm:flex-row">
            <QRCodeImage data={setup.otpauth_url} label={t("scanQr")} />
            <div className="min-w-0 flex-1 space-y-3">
              <div>
                <p className="mb-1.5 text-[12px] font-semibold uppercase tracking-wider text-(--muted-clr)">
                  {t("manualSecret")}
                </p>
                <div className="flex items-center gap-2">
                  <code className="block flex-1 truncate rounded-lg border border-(--line) bg-(--line-2) px-3 py-2 font-mono text-[12px] text-(--ink)">
                    {setup.secret}
                  </code>
                  <Button
                    type="button"
                    size="icon-sm"
                    variant="ghost"
                    onClick={() => copyText(setup.secret, "secret")}
                  >
                    {copiedSecret ? <Check size={14} /> : <Copy size={14} />}
                  </Button>
                </div>
              </div>
              <p className="text-[12px] text-(--faint)">{t("manualHint")}</p>
            </div>
          </div>

          <div>
            <div className="mb-2 flex items-center justify-between">
              <p className="text-[13px] font-semibold text-(--ink-2)">
                {t("recoveryTitle")}
              </p>
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() =>
                  copyText(setup.recovery_codes.join("\n"), "codes")
                }
              >
                {copiedCodes ? <Check size={14} /> : <Copy size={14} />}
                {t("copyCodes")}
              </Button>
            </div>
            <p className="mb-2 text-[12px] text-(--muted-clr)">
              {t("recoveryWarning")}
            </p>
            <div className="grid grid-cols-2 gap-1.5 rounded-xl border border-amber-100 bg-amber-50/60 p-3 sm:grid-cols-2">
              {setup.recovery_codes.map((c) => (
                <code
                  key={c}
                  className="font-mono text-[12px] text-(--ink-2)"
                >
                  {c}
                </code>
              ))}
            </div>
          </div>

          <div>
            <p className="mb-3 text-[13px] font-semibold text-(--ink-2)">
              {t("confirmCode")}
            </p>
            <OTPInput value={code} onChange={setCode} />
          </div>

          <div className="flex flex-wrap items-center justify-end gap-2 border-t border-(--line) pt-4">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                setPhase("idle");
                setSetup(null);
              }}
            >
              {t("cancel")}
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={enable.isPending}
              onClick={() => void onConfirmEnable()}
            >
              {enable.isPending ? (
                <Loader2 size={14} className="animate-spin" />
              ) : (
                <ShieldCheck size={14} />
              )}
              {t("enable")}
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (phase === "disable") {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ShieldOff size={16} className="text-red-600" />
            {t("disableTitle")}
          </CardTitle>
          <p className="text-sm text-(--muted-clr)">{t("disableSubtitle")}</p>
        </CardHeader>
        <CardContent className="space-y-4">
          <PasswordField
            id="mfa-disable-password"
            label={t("password")}
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                setPhase("idle");
                setPassword("");
              }}
            >
              {t("cancel")}
            </Button>
            <Button
              type="button"
              variant="danger-soft"
              size="sm"
              disabled={disable.isPending}
              onClick={() => void onConfirmDisable()}
            >
              {disable.isPending ? (
                <Loader2 size={14} className="animate-spin" />
              ) : null}
              {t("disableConfirm")}
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Lock size={16} className="text-(--ink-2)" />
          {t("title")}
        </CardTitle>
        <p className="text-sm text-(--muted-clr)">{t("subtitle")}</p>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-start justify-between gap-4 rounded-xl border border-(--line) bg-(--line-2) p-4">
          <div className="flex items-start gap-3">
            <div
              className={cx(
                "mt-0.5 flex h-9 w-9 items-center justify-center rounded-xl border",
                status.enabled
                  ? "border-green-100 bg-green-50 text-green-700"
                  : "border-(--line) bg-(--white) text-(--muted-clr)",
              )}
            >
              {status.enabled ? <ShieldCheck size={16} /> : <KeyRound size={16} />}
            </div>
            <div>
              <p className="text-[13px] font-semibold text-(--ink)">
                {status.enabled ? t("statusOn") : t("statusOff")}
              </p>
              <p className="mt-0.5 text-[12px] text-(--muted-clr)">
                {status.enabled
                  ? t("statusOnHint", {
                    count: status.remaining_recovery_codes,
                  })
                  : status.pending_enrollment
                    ? t("statusPending")
                    : t("statusOffHint")}
              </p>
            </div>
          </div>
          <span
            className={cx(
              "rounded-md border px-2 py-0.5 text-[11px] font-semibold",
              status.enabled
                ? "border-green-100 bg-green-50 text-green-700"
                : "border-(--line) bg-(--white) text-(--muted-clr)",
            )}
          >
            {status.enabled ? t("badgeOn") : t("badgeOff")}
          </span>
        </div>

        {!status.can_enroll ? (
          <p className="text-[13px] text-(--muted-clr)">{t("oauthOnly")}</p>
        ) : status.enabled ? (
          <div className="flex justify-end">
            <Button
              type="button"
              variant="danger-soft"
              size="sm"
              onClick={() => setPhase("disable")}
            >
              <ShieldOff size={14} />
              {t("disable")}
            </Button>
          </div>
        ) : (
          <div className="flex flex-wrap items-center justify-between gap-2">
            <p className="text-[12px] text-(--faint)">{t("enrollHint")}</p>
            <ConfirmDialog
              trigger={
                <Button
                  type="button"
                  size="sm"
                  disabled={startSetup.isPending}
                >
                  {startSetup.isPending ? (
                    <Loader2 size={14} className="animate-spin" />
                  ) : (
                    <Shield size={14} />
                  )}
                  {status.pending_enrollment ? t("resumeEnroll") : t("startEnroll")}
                </Button>
              }
              title={t("startEnrollTitle")}
              description={t("startEnrollBody")}
              confirmLabel={t("startEnroll")}
              cancelLabel={t("cancel")}
              pending={startSetup.isPending}
              onConfirm={onStartEnroll}
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function QRCodeImage({ data, label }: { data: string; label: string }) {
  // External QR renderer avoids a new dependency; secret + otpauth remain usable offline.
  const src = `https://api.qrserver.com/v1/create-qr-code/?size=180x180&data=${encodeURIComponent(data)}`;

  return (
    <div className="flex flex-col items-center gap-2">
      <div className="rounded-xl border border-(--line) bg-white p-2 shadow-sm">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={src} alt={label} width={180} height={180} />
      </div>
      <p className="text-[11px] text-(--muted-clr)">{label}</p>
    </div>
  );
}
