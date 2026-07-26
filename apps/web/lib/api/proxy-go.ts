import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { SESSION_COOKIE, REFRESH_COOKIE } from "@/lib/auth";
import { WORKSPACE_COOKIE } from "@/lib/workspace";

/**
 * Server-only Go proxy helpers.
 *
 * After httpOnly JWT migration, the browser never reads access/refresh tokens.
 * All Go calls go through Next Route Handlers which attach:
 *   Authorization: Bearer <hublio_session>
 *   X-Workspace-ID: <hublio_workspace>   (when present — required for
 *                                         orchestration/events JWT path)
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
  accessToken: string;
  refreshToken?: string;
  workspaceId?: string;
}

export async function readSessionTokens(): Promise<SessionTokens | null> {
  const store = await cookies();
  const accessToken = store.get(SESSION_COOKIE)?.value;
  if (!accessToken) return null;
  return {
    accessToken,
    refreshToken: store.get(REFRESH_COOKIE)?.value,
    workspaceId: store.get(WORKSPACE_COOKIE)?.value,
  };
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
  // Client-visible presence flag for soft-gate / AuthProvider.
  store.set("hublio_auth", "1", {
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
  store.delete("hublio_auth");
}

/**
 * Proxy a request to the Go API with the httpOnly JWT (+ optional workspace).
 * Streams Go status + body back verbatim.
 */
export async function proxyGoWithJWT(
  path: string,
  init: {
    method?: string;
    body?: unknown;
    headers?: Record<string, string>;
    params?: Record<string, string | undefined>;
    requireAuth?: boolean;
    requireWorkspace?: boolean;
  } = {},
): Promise<NextResponse> {
  const requireAuth = init.requireAuth !== false;
  const session = await readSessionTokens();

  if (requireAuth && !session?.accessToken) {
    throw new ProxyError(401, "UNAUTHORIZED", "No active session.");
  }
  if (init.requireWorkspace && !session?.workspaceId) {
    throw new ProxyError(400, "NO_WORKSPACE", "No active workspace selected.");
  }

  const url = new URL(`${goBaseUrl()}${path.startsWith("/") ? path : `/${path}`}`);
  if (init.params) {
    for (const [k, v] of Object.entries(init.params)) {
      if (v !== undefined && v !== "") url.searchParams.set(k, v);
    }
  }

  const headers: Record<string, string> = {
    Accept: "application/json",
    ...(init.body !== undefined ? { "Content-Type": "application/json" } : {}),
    ...init.headers,
  };
  if (session?.accessToken) {
    headers.Authorization = `Bearer ${session.accessToken}`;
  }
  if (session?.workspaceId) {
    headers["X-Workspace-ID"] = session.workspaceId;
  }

  const res = await fetch(url, {
    method: init.method ?? "GET",
    headers,
    body: init.body !== undefined ? JSON.stringify(init.body) : undefined,
    cache: "no-store",
  });

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
