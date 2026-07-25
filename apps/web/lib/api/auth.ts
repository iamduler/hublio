import type { User } from "@/types/auth";
import { api, unwrapData, type SuccessEnvelope } from "./client";
import { REFRESH_COOKIE, readCookie } from "@/lib/auth";

export type AuthTokenData = {
  access_token: string;
  refresh_token: string;
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

export const authApi = {
  login(email: string, password: string) {
    return api
      .post<SuccessEnvelope<AuthTokenData>>("/auth/login", { email, password })
      .then((res) => unwrapData<AuthTokenData>(res));
  },

  register(payload: RegisterPayload) {
    return api
      .post<SuccessEnvelope<RegisterResult>>("/auth/register", payload)
      .then((res) => unwrapData<RegisterResult>(res));
  },

  logout() {
    const refreshToken = readCookie(REFRESH_COOKIE);
    if (!refreshToken) return Promise.resolve();
    return api
      .post("/auth/logout", { refresh_token: refreshToken })
      .then(() => undefined)
      .catch(() => undefined);
  },
};
