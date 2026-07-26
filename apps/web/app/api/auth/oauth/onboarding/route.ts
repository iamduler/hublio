import { NextResponse } from "next/server";
import { cookies } from "next/headers";
import { goBaseUrl, proxyErrorResponse, ProxyError } from "@/lib/api/proxy-go";
import { OAUTH_ONBOARDING_COOKIE } from "@/lib/auth";

/** GET /api/auth/oauth/onboarding — preview email for Organization form */
export async function GET() {
  try {
    const store = await cookies();
    const token = store.get(OAUTH_ONBOARDING_COOKIE)?.value;
    if (!token) {
      throw new ProxyError(401, "ONBOARDING_EXPIRED", "Onboarding session expired.");
    }

    const res = await fetch(`${goBaseUrl()}/auth/oauth/onboarding`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "X-OAuth-Onboarding-Token": token,
      },
      cache: "no-store",
    });
    const text = await res.text();
    return new NextResponse(text, {
      status: res.status,
      headers: { "Content-Type": "application/json" },
    });
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
