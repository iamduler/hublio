import type { Organization, User } from "@/types/auth";
import { unwrapData, type SuccessEnvelope } from "./client";
import { ApiError } from "./client";

export type AuthTokenData = {
  user: User;
};

export type MeData = {
  user: User;
  organization: Organization;
};

export type MFAChallengeData = {
  mfa_required: true;
  mfa_token: string;
};

export type LoginResult = AuthTokenData | MFAChallengeData;

export function isMFAChallenge(data: LoginResult): data is MFAChallengeData {
  return (
    "mfa_required" in data &&
    data.mfa_required === true &&
    typeof data.mfa_token === "string"
  );
}

/**
 * Auth goes through Next `/api/auth/*` so JWT cookies stay httpOnly.
 * Do not call Go `/auth/*` from the browser.
 */
async function authFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
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
  login(email: string, password: string) {
    return authFetch<SuccessEnvelope<LoginResult>>("/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }).then((res) => unwrapData<LoginResult>(res));
  },

  logout() {
    return authFetch("/logout", { method: "POST" })
      .then(() => undefined)
      .catch(() => undefined);
  },

  session() {
    return authFetch<{ authenticated: boolean }>("/session", {
      method: "GET",
    });
  },

  me() {
    return authFetch<SuccessEnvelope<MeData>>("/me", { method: "GET" }).then(
      (res) => unwrapData<MeData>(res),
    );
  },

  refresh() {
    return authFetch("/refresh", { method: "POST" })
      .then(() => undefined)
      .catch(() => undefined);
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
};
