import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { SESSION_COOKIE } from "@/lib/auth";
import { WORKSPACE_COOKIE } from "@/lib/workspace";

/**
 * Server-only BFF helper.
 *
 * The browser dashboard authenticates with a JWT (identity/integration
 * routes). Orchestration + Events routes on the Go API are gated behind
 * `X-API-KEY` (machine auth). This module bridges the two: it mints a
 * per-workspace API key using the user's JWT, caches the plaintext in an
 * httpOnly cookie (never exposed to the browser), and proxies requests to Go
 * with that key.
 */

const WS_KEY_COOKIE = "hublio_ws_key";
const WS_KEY_MAX_AGE = 12 * 60 * 60; // 12h — re-mint after expiry
const BFF_KEY_NAME = "hublio-ui-bff";

function goBaseUrl(): string {
  return (
    process.env.API_INTERNAL_URL ??
    process.env.NEXT_PUBLIC_API_URL ??
    "http://localhost:8080/api/v1"
  );
}

export class BffError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
  ) {
    super(message);
    this.name = "BffError";
  }
}

export interface WorkspaceContext {
  token: string;
  workspaceId: string;
  apiKey: string;
}

type CachedKey = { workspaceId: string; key: string };

function readCachedKey(raw: string | undefined, workspaceId: string) {
  if (!raw) return undefined;
  try {
    const parsed = JSON.parse(raw) as CachedKey;
    if (parsed.workspaceId === workspaceId && parsed.key) return parsed.key;
  } catch {
    // ignore malformed cache
  }
  return undefined;
}

async function mintWorkspaceKey(
  token: string,
  workspaceId: string,
): Promise<string> {
  const res = await fetch(
    `${goBaseUrl()}/identity/workspaces/${workspaceId}/api-keys`,
    {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({ name: BFF_KEY_NAME }),
      cache: "no-store",
    },
  );

  if (!res.ok) {
    throw new BffError(
      res.status === 401 ? 401 : 502,
      "WORKSPACE_KEY_FAILED",
      "Unable to provision workspace credentials.",
    );
  }

  const body = (await res.json()) as {
    data?: { plaintext?: string };
    plaintext?: string;
  };
  const key = body.data?.plaintext ?? body.plaintext;
  if (!key) {
    throw new BffError(
      502,
      "WORKSPACE_KEY_FAILED",
      "Workspace credentials response was empty.",
    );
  }
  return key;
}

/**
 * Resolve the active workspace API key, minting + caching it if needed.
 * Must be called inside a Route Handler (can write cookies).
 */
export async function resolveWorkspaceContext(): Promise<WorkspaceContext> {
  const store = await cookies();
  const token = store.get(SESSION_COOKIE)?.value;
  const workspaceId = store.get(WORKSPACE_COOKIE)?.value;

  if (!token) {
    throw new BffError(401, "UNAUTHORIZED", "No active session.");
  }
  if (!workspaceId) {
    throw new BffError(400, "NO_WORKSPACE", "No active workspace selected.");
  }

  const cached = readCachedKey(store.get(WS_KEY_COOKIE)?.value, workspaceId);
  if (cached) {
    return { token, workspaceId, apiKey: cached };
  }

  const key = await mintWorkspaceKey(token, workspaceId);
  store.set(WS_KEY_COOKIE, JSON.stringify({ workspaceId, key }), {
    httpOnly: true,
    sameSite: "lax",
    secure: process.env.NODE_ENV === "production",
    path: "/",
    maxAge: WS_KEY_MAX_AGE,
  });

  return { token, workspaceId, apiKey: key };
}

/**
 * Proxy a request to the Go API using the workspace API key and stream the
 * Go envelope back to the browser verbatim (same status + body).
 */
export async function proxyToGo(
  ctx: WorkspaceContext,
  path: string,
  init: {
    method?: string;
    body?: unknown;
    headers?: Record<string, string>;
    params?: Record<string, string | undefined>;
  } = {},
): Promise<NextResponse> {
  const url = new URL(`${goBaseUrl()}${path}`);
  if (init.params) {
    for (const [k, v] of Object.entries(init.params)) {
      if (v !== undefined && v !== "") url.searchParams.set(k, v);
    }
  }

  const res = await fetch(url, {
    method: init.method ?? "GET",
    headers: {
      "X-API-KEY": ctx.apiKey,
      Accept: "application/json",
      ...(init.body !== undefined
        ? { "Content-Type": "application/json" }
        : {}),
      ...init.headers,
    },
    body: init.body !== undefined ? JSON.stringify(init.body) : undefined,
    cache: "no-store",
  });

  const text = await res.text();
  return new NextResponse(text, {
    status: res.status,
    headers: { "Content-Type": "application/json" },
  });
}

export function bffErrorResponse(err: unknown): NextResponse {
  if (err instanceof BffError) {
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
