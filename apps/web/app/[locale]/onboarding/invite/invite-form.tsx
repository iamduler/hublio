"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { ChevronLeft, ChevronRight, Loader2, Plus, Trash2 } from "lucide-react";
import { useRouter } from "@/i18n/navigation";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { teamApi } from "@/features/team/api";
import type { WorkspaceRole } from "@/features/team/types";
import { OnboardingShell, cx, inputBase } from "@/features/auth/auth-ui";
import {
  readOnboardingDraft,
  writeOnboardingDraft,
} from "@/lib/onboarding-state";
import { toast } from "@/lib/toast";

type InviteRow = { id: string; email: string; role: WorkspaceRole };

const ROLES: WorkspaceRole[] = ["admin", "member"];

function newRow(): InviteRow {
  return { id: String(Date.now() + Math.random()), email: "", role: "member" };
}

export function InviteOnboardingForm() {
  const t = useTranslations("auth.onboarding");
  const router = useRouter();
  const getError = useApiErrorMessage();
  const workspaceId = useActiveWorkspaceId();
  const [rows, setRows] = useState<InviteRow[]>([newRow(), newRow()]);
  const [loading, setLoading] = useState(false);

  function updateRow(id: string, field: "email" | "role", value: string) {
    setRows((prev) =>
      prev.map((row) => (row.id === id ? { ...row, [field]: value } : row)),
    );
  }

  function removeRow(id: string) {
    setRows((prev) => prev.filter((row) => row.id !== id));
  }

  function goComplete(invited: string[], pending: string[]) {
    writeOnboardingDraft({
      invitedEmails: invited,
      pendingInviteEmails: pending,
    });
    router.push("/onboarding/complete");
  }

  async function handleSend() {
    const valid = rows.filter(
      (r) => r.email.trim() && /\S+@\S+\.\S+/.test(r.email.trim()),
    );
    if (!valid.length) {
      toast.error(t("inviteNeedEmail"));
      return;
    }
    if (!workspaceId) {
      toast.error(t("inviteNoWorkspace"));
      router.replace("/onboarding/workspace");
      return;
    }

    setLoading(true);
    const invited: string[] = [];
    const pending: string[] = [];
    try {
      for (const row of valid) {
        try {
          await teamApi.addMember(workspaceId, {
            email: row.email.trim().toLowerCase(),
            role: row.role,
          });
          invited.push(row.email.trim().toLowerCase());
        } catch {
          pending.push(row.email.trim().toLowerCase());
        }
      }
      if (invited.length) {
        toast.success(t("inviteSent", { count: invited.length }));
      }
      if (pending.length) {
        toast.error(t("invitePendingHint", { count: pending.length }));
      }
      goComplete(invited, pending);
    } catch (err) {
      toast.error(getError(err));
    } finally {
      setLoading(false);
    }
  }

  function handleSkip() {
    const draft = readOnboardingDraft();
    goComplete(draft.invitedEmails ?? [], draft.pendingInviteEmails ?? []);
  }

  return (
    <OnboardingShell step={2}>
      <div className="mb-6">
        <h2 className="text-[18px] font-semibold tracking-tight text-(--ink)">
          {t("inviteTitle")}
        </h2>
        <p className="mt-1.5 text-[13px] text-(--muted-clr)">
          {t("inviteSubtitle")}
        </p>
      </div>

      <div className="space-y-2.5">
        {rows.map((row) => (
          <div key={row.id} className="flex items-center gap-2">
            <input
              type="email"
              value={row.email}
              onChange={(e) => updateRow(row.id, "email", e.target.value)}
              placeholder={t("inviteEmailPlaceholder")}
              className={cx(inputBase, "flex-1 border-(--line) hover:border-(--line-3)")}
            />
            <select
              value={row.role}
              onChange={(e) => updateRow(row.id, "role", e.target.value)}
              className="h-9 w-28 flex-shrink-0 rounded-lg border border-(--line) bg-(--white) px-2.5 text-[13px] text-(--ink-2) transition-colors hover:border-(--line-3) focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-600/20"
            >
              {ROLES.map((role) => (
                <option key={role} value={role}>
                  {t(`roles.${role}`)}
                </option>
              ))}
            </select>
            <button
              type="button"
              onClick={() => removeRow(row.id)}
              disabled={rows.length <= 1}
              className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg text-(--faint) transition-colors hover:bg-red-50 hover:text-red-500 disabled:cursor-not-allowed disabled:opacity-30"
            >
              <Trash2 size={13} />
            </button>
          </div>
        ))}
        <button
          type="button"
          onClick={() => setRows((r) => [...r, newRow()])}
          className="mt-1 flex items-center gap-1.5 text-[13px] font-medium text-blue-600 transition-colors hover:text-blue-700"
        >
          <Plus size={13} />
          {t("addAnother")}
        </button>
      </div>

      <div className="mt-3 rounded-xl border border-(--line) bg-(--line-2) p-3.5">
        <p className="mb-2 text-[11px] font-semibold uppercase tracking-wider text-(--muted-clr)">
          {t("roleDescriptions")}
        </p>
        <div className="space-y-1.5">
          {(["admin", "member"] as const).map((role) => (
            <div key={role} className="flex items-baseline gap-2">
              <span className="w-16 flex-shrink-0 text-[11px] font-semibold text-(--ink-2)">
                {t(`roles.${role}`)}
              </span>
              <span className="text-[11px] text-(--muted-clr)">
                {t(`roleHelp.${role}`)}
              </span>
            </div>
          ))}
        </div>
        <p className="mt-2 text-[11px] text-(--faint)">{t("inviteExistingOnly")}</p>
      </div>

      <div className="mt-7 flex items-center justify-between">
        <button
          type="button"
          onClick={() => router.push("/onboarding/workspace")}
          className="flex items-center gap-1.5 text-[13px] text-(--muted-clr) transition-colors hover:text-(--ink-2)"
        >
          <ChevronLeft size={14} />
          {t("back")}
        </button>
        <div className="flex items-center gap-2.5">
          <button
            type="button"
            onClick={handleSkip}
            className="h-9 rounded-lg px-4 text-[13px] font-medium text-(--ink-2) transition-colors hover:bg-(--line-2) hover:text-(--ink)"
          >
            {t("skip")}
          </button>
          <button
            type="button"
            onClick={() => void handleSend()}
            disabled={loading}
            className="flex h-9 items-center gap-2 rounded-lg bg-blue-600 px-5 text-[13px] font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? <Loader2 size={13} className="animate-spin" /> : null}
            {t("sendInvites")}
            <ChevronRight size={14} />
          </button>
        </div>
      </div>
    </OnboardingShell>
  );
}
