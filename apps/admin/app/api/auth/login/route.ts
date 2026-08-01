import { type NextRequest, NextResponse } from "next/server";
import {
  goBaseUrl,
  proxyErrorResponse,
  ProxyError,
  setAuthCookiesOnStore,
} from "@/lib/api/proxy-go";

type TokenPayload = {
  access_token?: string;
  refresh_token?: string;
  user?: unknown;
  mfa_required?: boolean;
  mfa_token?: string;
  data?: {
    access_token?: string;
    refresh_token?: string;
    user?: unknown;
    mfa_required?: boolean;
    mfa_token?: string;
  };
};

async function callGoAuth(path: string, body: unknown) {
  const res = await fetch(`${goBaseUrl()}${path}`, {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
    cache: "no-store",
  });
  const text = await res.text();
  let json: TokenPayload = {};
  try {
    json = text ? (JSON.parse(text) as TokenPayload) : {};
  } catch {
    // leave empty
  }
  return { res, text, json };
}

function extractTokens(json: TokenPayload) {
  const access = json.data?.access_token ?? json.access_token;
  const refresh = json.data?.refresh_token ?? json.refresh_token;
  const user = json.data?.user ?? json.user;
  return { access, refresh, user };
}

function extractMFAChallenge(json: TokenPayload) {
  const required = json.data?.mfa_required ?? json.mfa_required;
  const token = json.data?.mfa_token ?? json.mfa_token;
  return required && token ? token : undefined;
}

/** POST /api/auth/login — sets httpOnly JWT cookies; returns envelope without tokens. */
export async function POST(request: NextRequest) {
  try {
    const body = (await request.json()) as unknown;
    const { res, text, json } = await callGoAuth("/auth/login", body);
    if (!res.ok) {
      return new NextResponse(text, {
        status: res.status,
        headers: { "Content-Type": "application/json" },
      });
    }

    const mfaToken = extractMFAChallenge(json);
    if (mfaToken) {
      return NextResponse.json({
        status: "success",
        message: "mfa required",
        data: { mfa_required: true, mfa_token: mfaToken },
      });
    }

    const { access, refresh, user } = extractTokens(json);
    if (!access || !refresh) {
      throw new ProxyError(502, "AUTH_FAILED", "Login response missing tokens.");
    }
    await setAuthCookiesOnStore({ accessToken: access, refreshToken: refresh });

    return NextResponse.json({
      status: "success",
      message: "login",
      data: { user },
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
