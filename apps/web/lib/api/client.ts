import { SESSION_COOKIE, clearAuthCookies, readCookie } from "@/lib/auth";
import { getBrowserLocale } from "@/lib/i18n/locale";
import { unwrapData, type ErrorEnvelope, type SuccessEnvelope } from "./types";

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

const BASE_URL = process.env.NEXT_PUBLIC_API_URL ?? "";

function readBrowserToken(): string | undefined {
  return readCookie(SESSION_COOKIE);
}

function buildUrl(path: string, params?: Record<string, unknown>): string {
  if (!params) return path;
  const qs = Object.entries(params)
    .filter(([, v]) => v !== undefined && v !== null && v !== "")
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
    .join("&");
  return qs ? `${path}?${qs}` : path;
}

export type ApiFetchInit = RequestInit & {
  token?: string;
  params?: Record<string, unknown>;
  locale?: string;
};

export async function apiFetch<T = unknown>(
  path: string,
  init: ApiFetchInit = {},
): Promise<T> {
  const {
    token: explicitToken,
    params,
    locale: explicitLocale,
    headers: extraHeaders,
    body,
    ...rest
  } = init;
  const token = explicitToken ?? readBrowserToken();
  const locale = explicitLocale ?? getBrowserLocale();

  const url = `${BASE_URL}${buildUrl(path, params)}`;
  const isFormData = typeof FormData !== "undefined" && body instanceof FormData;

  const headers: Record<string, string> = {
    Accept: "application/json",
    "Accept-Language": locale,
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(extraHeaders as Record<string, string>),
  };

  if (!isFormData && body !== undefined && !headers["Content-Type"]) {
    headers["Content-Type"] = "application/json";
  }

  const res = await fetch(url, { ...rest, body, headers });

  if (!res.ok) {
    if (res.status === 401) {
      // No public refresh endpoint yet — clear session on unauthorized.
      clearAuthCookies();
    }

    let code = "UNKNOWN";
    let message = res.statusText;
    try {
      const errBody = (await res.json()) as ErrorEnvelope & {
        message?: string;
      };
      code = errBody.code ?? code;
      message = errBody.detail ?? errBody.error ?? errBody.message ?? message;
    } catch {
      // ignore parse errors
    }
    throw new ApiError(res.status, code, message);
  }

  if (res.status === 204) return undefined as T;

  return res.json() as Promise<T>;
}

export const api = {
  get: <T>(
    path: string,
    params?: Record<string, unknown>,
    init?: ApiFetchInit,
  ) => apiFetch<T>(path, { ...init, method: "GET", params }),

  post: <T>(path: string, body: unknown, init?: ApiFetchInit) =>
    apiFetch<T>(path, {
      ...init,
      method: "POST",
      body: body instanceof FormData ? body : JSON.stringify(body),
    }),

  put: <T>(path: string, body: unknown, init?: ApiFetchInit) =>
    apiFetch<T>(path, {
      ...init,
      method: "PUT",
      body: body instanceof FormData ? body : JSON.stringify(body),
    }),

  patch: <T>(path: string, body: unknown, init?: ApiFetchInit) =>
    apiFetch<T>(path, {
      ...init,
      method: "PATCH",
      body: body instanceof FormData ? body : JSON.stringify(body),
    }),

  del: <T>(path: string, init?: ApiFetchInit) =>
    apiFetch<T>(path, { ...init, method: "DELETE" }),
};

export { unwrapData };
export type { SuccessEnvelope };
