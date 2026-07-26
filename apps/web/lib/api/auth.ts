import type { User } from "@/types/auth";
import { unwrapData, type SuccessEnvelope } from "./client";
import { ApiError } from "./client";

export type AuthTokenData = {
  user: User;
};

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
  login(email: string, password: string) {
    return authFetch<SuccessEnvelope<AuthTokenData>>("/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }).then((res) => unwrapData<AuthTokenData>(res));
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
};
