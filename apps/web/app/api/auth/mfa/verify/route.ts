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
  data?: {
    access_token?: string;
    refresh_token?: string;
    user?: unknown;
  };
};

/** POST /api/auth/mfa/verify — completes the login challenge and sets httpOnly JWT cookies. */
export async function POST(request: NextRequest) {
  try {
    const body = (await request.json()) as unknown;
    const res = await fetch(`${goBaseUrl()}/auth/mfa/verify`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
      cache: "no-store",
    });
    const text = await res.text();
    if (!res.ok) {
      return new NextResponse(text, {
        status: res.status,
        headers: { "Content-Type": "application/json" },
      });
    }

    let json: TokenPayload = {};
    try {
      json = text ? (JSON.parse(text) as TokenPayload) : {};
    } catch {
      // leave empty
    }

    const access = json.data?.access_token ?? json.access_token;
    const refresh = json.data?.refresh_token ?? json.refresh_token;
    const user = json.data?.user ?? json.user;
    if (!access || !refresh) {
      throw new ProxyError(502, "AUTH_FAILED", "MFA response missing tokens.");
    }
    await setAuthCookiesOnStore({ accessToken: access, refreshToken: refresh });

    // Strip tokens from the browser-visible payload.
    return NextResponse.json({
      status: "success",
      message: "login",
      data: { user },
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
