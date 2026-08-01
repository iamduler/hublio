import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import {
  AUTH_PRESENT_COOKIE,
  SESSION_COOKIE,
  REFRESH_COOKIE,
} from "@/lib/auth";

/**
 * Server-only Go proxy helpers for the admin console.
 * Browser never reads access/refresh tokens; Next Route Handlers attach Bearer.
 */

export function goBaseUrl(): string {
  return (
    process.env.API_INTERNAL_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    "http://localhost:8080/api/v1"
  );
}

export class ProxyError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "ProxyError";
  }
}

export interface SessionTokens {
  accessToken?: string;
  refreshToken?: string;
}

export async function readSessionTokens(): Promise<SessionTokens | null> {
  const store = await cookies();
  const accessToken = store.get(SESSION_COOKIE)?.value;
  const refreshToken = store.get(REFRESH_COOKIE)?.value;
  if (!accessToken && !refreshToken) return null;
  return { accessToken, refreshToken };
}

const ACCESS_MAX_AGE = 15 * 60;
const REFRESH_MAX_AGE = 7 * 24 * 60 * 60;

export async function setAuthCookiesOnStore(opts: {
  accessToken: string;
  refreshToken: string;
  accessMaxAge?: number;
  refreshMaxAge?: number;
}): Promise<void> {
  const store = await cookies();
  const secure = process.env.NODE_ENV === "production";
  store.set(SESSION_COOKIE, opts.accessToken, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    maxAge: opts.accessMaxAge ?? ACCESS_MAX_AGE,
  });
  store.set(REFRESH_COOKIE, opts.refreshToken, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    maxAge: opts.refreshMaxAge ?? REFRESH_MAX_AGE,
  });
  store.set(AUTH_PRESENT_COOKIE, "1", {
    httpOnly: false,
    sameSite: "lax",
    secure,
    path: "/",
    maxAge: opts.refreshMaxAge ?? REFRESH_MAX_AGE,
  });
}

export async function clearAuthCookiesOnStore(): Promise<void> {
  const store = await cookies();
  store.delete(SESSION_COOKIE);
  store.delete(REFRESH_COOKIE);
  store.delete(AUTH_PRESENT_COOKIE);
}

type TokenEnvelope = {
  access_token?: string;
  refresh_token?: string;
  data?: {
    access_token?: string;
    refresh_token?: string;
  };
};

function extractRotatedTokens(json: TokenEnvelope): {
  access?: string;
  refresh?: string;
} {
  return {
    access: json.data?.access_token ?? json.access_token,
    refresh: json.data?.refresh_token ?? json.refresh_token,
  };
}

export async function refreshSessionAccessToken(
  refreshToken: string,
): Promise<string | null> {
  try {
    const res = await fetch(`${goBaseUrl()}/auth/refresh`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
      cache: "no-store",
    });
    const text = await res.text();
    if (!res.ok) {
      await clearAuthCookiesOnStore().catch(() => undefined);
      return null;
    }
    let json: TokenEnvelope = {};
    try {
      json = text ? (JSON.parse(text) as TokenEnvelope) : {};
    } catch {
      await clearAuthCookiesOnStore().catch(() => undefined);
      return null;
    }
    const { access, refresh } = extractRotatedTokens(json);
    if (!access || !refresh) {
      await clearAuthCookiesOnStore().catch(() => undefined);
      return null;
    }
    await setAuthCookiesOnStore({ accessToken: access, refreshToken: refresh });
    return access;
  } catch {
    await clearAuthCookiesOnStore().catch(() => undefined);
    return null;
  }
}

export async function proxyGoWithJWT(
  path: string,
  init: {
    method?: string;
    body?: unknown;
    headers?: Record<string, string>;
    params?: Record<string, string | undefined>;
    requireAuth?: boolean;
  } = {},
): Promise<NextResponse> {
  const requireAuth = init.requireAuth !== false;
  let session = await readSessionTokens();

  if (requireAuth && !session?.accessToken && !session?.refreshToken) {
    throw new ProxyError(401, "UNAUTHORIZED", "No active session.");
  }

  let accessToken = session?.accessToken;
  let didRefresh = false;

  if (requireAuth && !accessToken && session?.refreshToken) {
    accessToken =
      (await refreshSessionAccessToken(session.refreshToken)) ?? undefined;
    didRefresh = true;
    if (!accessToken) {
      throw new ProxyError(401, "UNAUTHORIZED", "Session expired.");
    }
    session = await readSessionTokens();
  }

  const url = new URL(
    `${goBaseUrl()}${path.startsWith("/") ? path : `/${path}`}`,
  );
  if (init.params) {
    for (const [k, v] of Object.entries(init.params)) {
      if (v !== undefined && v !== "") url.searchParams.set(k, v);
    }
  }

  const buildHeaders = (bearer?: string): Record<string, string> => {
    const headers: Record<string, string> = {
      Accept: "application/json",
      ...(init.body !== undefined ? { "Content-Type": "application/json" } : {}),
      ...init.headers,
    };
    if (bearer) {
      headers.Authorization = `Bearer ${bearer}`;
    }
    return headers;
  };

  const doFetch = (bearer?: string) =>
    fetch(url, {
      method: init.method ?? "GET",
      headers: buildHeaders(bearer),
      body: init.body !== undefined ? JSON.stringify(init.body) : undefined,
      cache: "no-store",
    });

  let res = await doFetch(accessToken);
  if (
    requireAuth &&
    res.status === 401 &&
    !didRefresh &&
    session?.refreshToken
  ) {
    const refreshed = await refreshSessionAccessToken(session.refreshToken);
    if (!refreshed) {
      return new NextResponse(
        JSON.stringify({ error: "Session expired.", code: "UNAUTHORIZED" }),
        { status: 401, headers: { "Content-Type": "application/json" } },
      );
    }
    accessToken = refreshed;
    session = await readSessionTokens();
    res = await doFetch(accessToken);
  }

  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

export function proxyErrorResponse(err: unknown): NextResponse {
  if (err instanceof ProxyError) {
    return NextResponse.json(
      { error: err.message, code: err.code },
      { status: err.status },
    );
  }
  return NextResponse.json(
    { error: "Unexpected proxy error.", code: "INTERNAL" },
    { status: 500 },
  );
}
