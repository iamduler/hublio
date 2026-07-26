import { clearAuthCookies } from "@/lib/auth";
import { ApiError } from "./client";
import { unwrapData, type ErrorEnvelope } from "./types";

/**
 * @deprecated Prefer `lib/api/client` (`/api/go`) after JWT is accepted on
 * orchestration/events. Kept for any residual `/api/intents|executions|events`
 * route aliases during migration.
 */
async function bffFetch<T = unknown>(
  path: string,
  init: RequestInit & { params?: Record<string, unknown> } = {},
): Promise<T> {
  const { params, body, headers, ...rest } = init;

  const url = buildUrl(path, params);
  const res = await fetch(url, {
    ...rest,
    body,
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      ...(body !== undefined && !(body instanceof FormData)
        ? { "Content-Type": "application/json" }
        : {}),
      ...(headers as Record<string, string>),
    },
  });

  if (!res.ok) {
    if (res.status === 401) {
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
      // ignore parse error
    }
    throw new ApiError(res.status, code, message);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

function buildUrl(path: string, params?: Record<string, unknown>): string {
  const base = `/api${path}`;
  if (!params) return base;
  const qs = Object.entries(params)
    .filter(([, v]) => v !== undefined && v !== null && v !== "")
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
    .join("&");
  return qs ? `${base}?${qs}` : base;
}

export const bff = {
  get: <T>(path: string, params?: Record<string, unknown>) =>
    bffFetch<T>(path, { method: "GET", params }),

  post: <T>(
    path: string,
    body: unknown,
    headers?: Record<string, string>,
  ) =>
    bffFetch<T>(path, {
      method: "POST",
      body: JSON.stringify(body),
      headers,
    }),
};

export { unwrapData };
