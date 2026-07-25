import { getLocale } from "next-intl/server";
import { apiFetch } from "./client";

type ServerInit = RequestInit & {
  token?: string;
  params?: Record<string, unknown>;
  locale?: string;
};

/**
 * Server-side API fetch that always sends `Accept-Language`
 * from the active next-intl locale.
 */
export async function serverApiFetch<T = unknown>(
  path: string,
  init: ServerInit = {},
): Promise<T> {
  const locale = init.locale ?? (await getLocale());
  return apiFetch<T>(path, { ...init, locale });
}

export const serverApi = {
  get: async <T>(
    path: string,
    params?: Record<string, unknown>,
    init?: Omit<ServerInit, "params" | "method" | "body">,
  ) => serverApiFetch<T>(path, { ...init, method: "GET", params }),

  post: async <T>(
    path: string,
    body: unknown,
    init?: Omit<ServerInit, "method" | "body">,
  ) =>
    serverApiFetch<T>(path, {
      ...init,
      method: "POST",
      body: JSON.stringify(body),
    }),

  put: async <T>(
    path: string,
    body: unknown,
    init?: Omit<ServerInit, "method" | "body">,
  ) =>
    serverApiFetch<T>(path, {
      ...init,
      method: "PUT",
      body: JSON.stringify(body),
    }),

  patch: async <T>(
    path: string,
    body: unknown,
    init?: Omit<ServerInit, "method" | "body">,
  ) =>
    serverApiFetch<T>(path, {
      ...init,
      method: "PATCH",
      body: JSON.stringify(body),
    }),

  del: async <T>(path: string, init?: Omit<ServerInit, "method" | "body">) =>
    serverApiFetch<T>(path, { ...init, method: "DELETE" }),
};
