import { type NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import {
  goBaseUrl,
  proxyErrorResponse,
  ProxyError,
  setAuthCookiesOnStore,
} from "@/lib/api/proxy-go";
import { OAUTH_ONBOARDING_COOKIE } from "@/lib/auth";

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

/** POST /api/auth/oauth/complete-registration */
export async function POST(request: NextRequest) {
  try {
    const body = (await request.json()) as {
      organization_name?: string;
      workspace_name?: string;
      environment?: string;
    };
    const store = await cookies();
    const onboardingToken = store.get(OAUTH_ONBOARDING_COOKIE)?.value;
    if (!onboardingToken) {
      throw new ProxyError(401, "ONBOARDING_EXPIRED", "Onboarding session expired.");
    }

    const res = await fetch(`${goBaseUrl()}/auth/oauth/complete-registration`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        onboarding_token: onboardingToken,
        organization_name: body.organization_name,
        workspace_name: body.workspace_name,
        environment: body.environment,
      }),
      cache: "no-store",
    });
    const text = await res.text();
    let json: TokenPayload = {};
    try {
      json = text ? (JSON.parse(text) as TokenPayload) : {};
    } catch {
      // ignore
    }
    if (!res.ok) {
      return new NextResponse(text, {
        status: res.status,
        headers: { "Content-Type": "application/json" },
      });
    }

    const access = json.data?.access_token ?? json.access_token;
    const refresh = json.data?.refresh_token ?? json.refresh_token;
    const user = json.data?.user ?? json.user;
    if (!access || !refresh) {
      throw new ProxyError(502, "AUTH_FAILED", "Registration response missing tokens.");
    }

    store.delete(OAUTH_ONBOARDING_COOKIE);
    await setAuthCookiesOnStore({ accessToken: access, refreshToken: refresh });

    return NextResponse.json({
      status: "success",
      message: "registered",
      data: { user },
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
