"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  ArrowRight,
  BookOpen,
  Building2,
  Check,
  Globe2,
  Rocket,
  Users,
} from "lucide-react";
import { useRouter } from "@/i18n/navigation";
import { OnboardingShell } from "@/features/auth/auth-ui";
import {
  clearOnboardingDraft,
  readOnboardingDraft,
  type OnboardingDraft,
} from "@/lib/onboarding-state";

export function CompleteOnboardingForm() {
  const t = useTranslations("auth.onboarding");
  const router = useRouter();
  const [draft, setDraft] = useState<OnboardingDraft>({});

  useEffect(() => {
    setDraft(readOnboardingDraft());
  }, []);

  const invited = draft.invitedEmails?.length ?? 0;
  const pending = draft.pendingInviteEmails?.length ?? 0;
  const memberCount = 1 + invited;

  function enterWorkspace() {
    clearOnboardingDraft();
    router.replace("/dashboard");
  }

  return (
    <OnboardingShell step={3}>
      <div className="mb-7 text-center">
        <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-blue-600 shadow-lg shadow-blue-600/25">
          <Rocket size={24} className="text-white" />
        </div>
        <h2 className="text-[20px] font-semibold tracking-tight text-(--ink)">
          {t("completeTitle")}
        </h2>
        <p className="mt-2 text-[13px] leading-relaxed text-(--muted-clr)">
          {t("completeSubtitle")}
        </p>
      </div>

      <div className="mb-6 space-y-2.5">
        <div>
          <p className="mb-2 text-[11px] font-semibold uppercase tracking-wider text-(--muted-clr)">
            {t("summaryOrg")}
          </p>
          <div className="flex items-center gap-3 rounded-xl border border-(--line) bg-(--white) p-3.5">
            <div className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl border border-blue-100 bg-blue-50">
              <Building2 size={16} className="text-blue-600" />
            </div>
            <div>
              <p className="text-[13px] font-semibold text-(--ink)">
                {draft.organizationName || t("summaryOrgFallback")}
              </p>
              <p className="text-[11px] text-(--muted-clr)">{t("summaryOrgHint")}</p>
            </div>
            <div className="ml-auto">
              <span className="rounded-md border border-green-100 bg-green-50 px-2 py-0.5 text-[11px] font-semibold text-green-700">
                {t("active")}
              </span>
            </div>
          </div>
        </div>

        <div>
          <p className="mb-2 text-[11px] font-semibold uppercase tracking-wider text-(--muted-clr)">
            {t("summaryWorkspace")}
          </p>
          <div className="flex items-center gap-3 rounded-xl border border-(--line) bg-(--white) p-3.5">
            <div className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl border border-violet-100 bg-violet-50">
              <Globe2 size={16} className="text-violet-600" />
            </div>
            <div>
              <p className="text-[13px] font-semibold text-(--ink)">
                {draft.workspace?.name || t("summaryWorkspaceFallback")}
              </p>
              <p className="font-mono text-[11px] text-(--muted-clr)">
                {draft.workspace
                  ? `${draft.workspace.id.slice(0, 8)}… · ${draft.workspace.environment}`
                  : t("summaryWorkspaceHint")}
              </p>
            </div>
            {draft.workspace ? (
              <div className="ml-auto">
                <span className="rounded-md border border-green-100 bg-green-50 px-2 py-0.5 text-[11px] font-semibold text-green-700">
                  {draft.workspace.environment}
                </span>
              </div>
            ) : null}
          </div>
        </div>

        <div>
          <p className="mb-2 text-[11px] font-semibold uppercase tracking-wider text-(--muted-clr)">
            {t("summaryTeam")}
          </p>
          <div className="flex items-center gap-3 rounded-xl border border-(--line) bg-(--white) p-3.5">
            <div className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-xl border border-green-100 bg-green-50">
              <Users size={16} className="text-green-600" />
            </div>
            <div>
              <p className="text-[13px] font-semibold text-(--ink)">
                {t("memberCount", { count: memberCount })}
              </p>
              <p className="text-[11px] text-(--muted-clr)">
                {pending > 0
                  ? t("pendingInvites", { count: pending })
                  : t("noPendingInvites")}
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="mb-6 rounded-xl border border-blue-100 bg-blue-50 p-4">
        <p className="mb-2 text-[12px] font-semibold text-blue-800">{t("nextTitle")}</p>
        <ul className="space-y-1.5">
          {(
            [
              t("nextConnect"),
              t("nextInvite"),
              t("nextDocs"),
            ] as const
          ).map((item) => (
            <li key={item} className="flex items-start gap-2 text-[12px] text-blue-700">
              <Check size={12} className="mt-0.5 flex-shrink-0 text-blue-500" strokeWidth={2.5} />
              {item}
            </li>
          ))}
        </ul>
      </div>

      <div className="space-y-2.5">
        <button
          type="button"
          onClick={enterWorkspace}
          className="flex h-10 w-full items-center justify-center gap-2 rounded-lg bg-blue-600 text-[14px] font-semibold text-white shadow-sm shadow-blue-600/20 transition-colors hover:bg-blue-700"
        >
          {t("enterWorkspace")}
          <ArrowRight size={15} />
        </button>
        <a
          href="https://docs.hublio.dev"
          target="_blank"
          rel="noreferrer"
          className="flex h-10 w-full items-center justify-center gap-2 rounded-lg border border-(--line) bg-(--white) text-[13px] font-medium text-(--ink-2) transition-colors hover:border-(--line-3) hover:bg-(--line-2)"
        >
          <BookOpen size={14} className="text-(--muted-clr)" />
          {t("viewDocs")}
        </a>
      </div>
    </OnboardingShell>
  );
}
