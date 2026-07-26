/**
 * Client-only draft state for the multi-step onboarding wizard.
 * Redis/session cookies are not required — values are disposable UI context.
 */

const KEY = "hublio_onboarding_draft";

export type OnboardingDraft = {
  organizationName?: string;
  workspace?: {
    id: string;
    name: string;
    environment: string;
  };
  invitedEmails?: string[];
  pendingInviteEmails?: string[];
};

function canUseStorage(): boolean {
  return typeof window !== "undefined" && typeof sessionStorage !== "undefined";
}

export function readOnboardingDraft(): OnboardingDraft {
  if (!canUseStorage()) return {};
  try {
    const raw = sessionStorage.getItem(KEY);
    if (!raw) return {};
    return JSON.parse(raw) as OnboardingDraft;
  } catch {
    return {};
  }
}

export function writeOnboardingDraft(patch: Partial<OnboardingDraft>): OnboardingDraft {
  const next = { ...readOnboardingDraft(), ...patch };
  if (canUseStorage()) {
    sessionStorage.setItem(KEY, JSON.stringify(next));
  }
  return next;
}

export function clearOnboardingDraft(): void {
  if (canUseStorage()) {
    sessionStorage.removeItem(KEY);
  }
}
