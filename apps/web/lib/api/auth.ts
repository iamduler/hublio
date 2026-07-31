import type { User } from "@/types/auth";
import { unwrapData, type SuccessEnvelope } from "./client";
import { ApiError } from "./client";

export type AuthTokenData = {
  user: User;
};

export type MeData = {
  user: User;
  organization: {
    id: string;
    name: string;
    status: string;
    created_at?: string;
    updated_at?: string;
  };
};

export type MFAChallengeData = {
  mfa_required: true;
  mfa_token: string;
};

export type LoginResult = AuthTokenData | MFAChallengeData;

export function isMFAChallenge(
  data: LoginResult,
): data is MFAChallengeData {
  return (
    "mfa_required" in data &&
    data.mfa_required === true &&
    typeof data.mfa_token === "string"
  );
}

export type RegisterPayload = {
  organization_name: string;
  email: string;
  password: string;
  full_name: string;
  workspace_name?: string;
  environment?: string;
};

export type RegisterResult = {
  organization: Record<string, unknown>;
  workspace: Record<string, unknown>;
  user: User;
};

export type MFAStatusData = {
  enabled: boolean;
  pending_enrollment: boolean;
  remaining_recovery_codes: number;
  can_enroll: boolean;
};

export type MFASetupData = {
  secret: string;
  otpauth_url: string;
  recovery_codes: string[];
  warning?: string;
};

/**
 * Auth goes through Next `/api/auth/*` so JWT cookies stay httpOnly.
 * Do not call Go `/auth/*` from the browser.
 */
async function authFetch<T>(
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const res = await fetch(`/api/auth${path}`, {
    ...init,
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      ...(init.headers as Record<string, string>),
    },
  });

  if (!res.ok) {
    let code = "UNKNOWN";
    let message = res.statusText;
    try {
      const errBody = (await res.json()) as {
        code?: string;
        detail?: string;
        error?: string;
        message?: string;
      };
      code = errBody.code ?? code;
      message = errBody.detail ?? errBody.error ?? errBody.message ?? message;
    } catch {
      // ignore
    }
    throw new ApiError(res.status, code, message);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const authApi = {
  login(email: string, password: string, deviceId?: string) {
    return authFetch<SuccessEnvelope<LoginResult>>("/login", {
      method: "POST",
      body: JSON.stringify({
        email,
        password,
        ...(deviceId ? { device_id: deviceId } : {}),
      }),
    }).then((res) => unwrapData<LoginResult>(res));
  },

  register(payload: RegisterPayload) {
    return authFetch<SuccessEnvelope<RegisterResult>>("/register", {
      method: "POST",
      body: JSON.stringify(payload),
    }).then((res) => unwrapData<RegisterResult>(res));
  },

  logout() {
    return authFetch("/logout", { method: "POST" })
      .then(() => undefined)
      .catch(() => undefined);
  },

  session() {
    return authFetch<{ authenticated: boolean; workspace_id: string | null }>(
      "/session",
      { method: "GET" },
    );
  },

  me() {
    return authFetch<SuccessEnvelope<MeData>>("/me", { method: "GET" }).then(
      (res) => unwrapData<MeData>(res),
    );
  },

  oauthProviders() {
    return authFetch<SuccessEnvelope<{ providers: string[] }>>(
      "/oauth/providers",
      { method: "GET" },
    ).then((res) => {
      const data = unwrapData<{ providers: string[] }>(res);
      return (data.providers ?? []).filter(
        (p): p is "google" | "microsoft" | "github" =>
          p === "google" || p === "microsoft" || p === "github",
      );
    });
  },

  oauthOnboardingPreview() {
    return authFetch<
      SuccessEnvelope<{ email: string; full_name: string; provider: string }>
    >("/oauth/onboarding", { method: "GET" }).then((res) =>
      unwrapData<{ email: string; full_name: string; provider: string }>(res),
    );
  },

  completeOAuthRegistration(payload: {
    organization_name: string;
    workspace_name?: string;
    environment?: string;
  }) {
    return authFetch<SuccessEnvelope<AuthTokenData>>(
      "/oauth/complete-registration",
      {
        method: "POST",
        body: JSON.stringify(payload),
      },
    ).then((res) => unwrapData<AuthTokenData>(res));
  },

  forgotPassword(email: string) {
    return authFetch<SuccessEnvelope<Record<string, unknown>>>(
      "/forgot-password",
      {
        method: "POST",
        body: JSON.stringify({ email }),
      },
    ).then((res) => unwrapData(res));
  },

  resetPassword(token: string, password: string) {
    return authFetch<SuccessEnvelope<Record<string, unknown>>>(
      "/reset-password",
      {
        method: "POST",
        body: JSON.stringify({ token, password }),
      },
    ).then((res) => unwrapData(res));
  },

  requestEmailVerification(email: string) {
    return authFetch<SuccessEnvelope<Record<string, unknown>>>(
      "/verify-email/request",
      {
        method: "POST",
        body: JSON.stringify({ email }),
      },
    ).then((res) => unwrapData(res));
  },

  verifyEmail(email: string, code: string) {
    return authFetch<SuccessEnvelope<Record<string, unknown>>>(
      "/verify-email",
      {
        method: "POST",
        body: JSON.stringify({ email, code }),
      },
    ).then((res) => unwrapData(res));
  },

  verifyMFA(payload: {
    mfa_token: string;
    code?: string;
    recovery_code?: string;
    trust_device?: boolean;
    device_id?: string;
  }) {
    return authFetch<SuccessEnvelope<AuthTokenData>>("/mfa/verify", {
      method: "POST",
      body: JSON.stringify(payload),
    }).then((res) => unwrapData<AuthTokenData>(res));
  },

  mfaStatus() {
    return authFetch<SuccessEnvelope<MFAStatusData>>("/mfa/status", {
      method: "GET",
    }).then((res) => unwrapData<MFAStatusData>(res));
  },

  mfaSetup() {
    return authFetch<SuccessEnvelope<MFASetupData>>("/mfa/setup", {
      method: "POST",
      body: JSON.stringify({}),
    }).then((res) => unwrapData<MFASetupData>(res));
  },

  mfaEnable(code: string) {
    return authFetch<SuccessEnvelope<Record<string, unknown>>>("/mfa/enable", {
      method: "POST",
      body: JSON.stringify({ code }),
    }).then((res) => unwrapData(res));
  },

  mfaDisable(password: string) {
    return authFetch<SuccessEnvelope<Record<string, unknown>>>("/mfa/disable", {
      method: "POST",
      body: JSON.stringify({ password }),
    }).then((res) => unwrapData(res));
  },
};
