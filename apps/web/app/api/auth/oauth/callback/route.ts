import { type NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import {
  clearAuthCookiesOnStore,
  goBaseUrl,
  setAuthCookiesOnStore,
} from "@/lib/api/proxy-go";
import {
  OAUTH_ONBOARDING_COOKIE,
  OAUTH_PKCE_COOKIE,
  OAUTH_PROVIDER_COOKIE,
  OAUTH_STATE_COOKIE,
  USER_BOOT_COOKIE,
} from "@/lib/auth";

type CallbackPayload = {
  status?: string;
  access_token?: string;
  refresh_token?: string;
  onboarding_token?: string;
  email?: string;
  data?: {
    status?: string;
    access_token?: string;
    refresh_token?: string;
    onboarding_token?: string;
    email?: string;
    user?: unknown;
  };
};

function oauthRedirectURI(request: NextRequest): string {
  const configured = process.env.OAUTH_REDIRECT_URI;
  if (configured) return configured;
  return `${request.nextUrl.origin}/api/auth/oauth/callback`;
}

function localeHome(request: NextRequest, path: string): string {
  const referer = request.headers.get("referer") || "";
  const match = referer.match(/\/(en|vi)\//);
  const locale = match?.[1] || "en";
  return `${request.nextUrl.origin}/${locale}${path}`;
}

/** GET /api/auth/oauth/callback — IdP returns here with ?code&state */
export async function GET(request: NextRequest) {
  const store = await cookies();
  const code = request.nextUrl.searchParams.get("code") || "";
  const state = request.nextUrl.searchParams.get("state") || "";
  const error = request.nextUrl.searchParams.get("error");

  const expectedState = store.get(OAUTH_STATE_COOKIE)?.value || "";
  const verifier = store.get(OAUTH_PKCE_COOKIE)?.value || "";
  const provider = store.get(OAUTH_PROVIDER_COOKIE)?.value || "";

  store.delete(OAUTH_STATE_COOKIE);
  store.delete(OAUTH_PKCE_COOKIE);
  store.delete(OAUTH_PROVIDER_COOKIE);

  if (error) {
    return NextResponse.redirect(
      localeHome(request, `/login?oauth_error=${encodeURIComponent(error)}`),
    );
  }
  if (!code || !state || !expectedState || state !== expectedState || !verifier || !provider) {
    return NextResponse.redirect(
      localeHome(request, `/login?oauth_error=${encodeURIComponent("invalid_oauth_state")}`),
    );
  }

  try {
    const res = await fetch(`${goBaseUrl()}/auth/oauth/callback`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        provider,
        code,
        code_verifier: verifier,
        redirect_uri: oauthRedirectURI(request),
      }),
      cache: "no-store",
    });
    const json = (await res.json()) as CallbackPayload;
    if (!res.ok) {
      const detail =
        (json as { detail?: string; error?: string }).detail ||
        (json as { error?: string }).error ||
        "oauth_failed";
      return NextResponse.redirect(
        localeHome(request, `/login?oauth_error=${encodeURIComponent(detail)}`),
      );
    }

    const status = json.data?.status ?? json.status;
    if (status === "onboarding_required") {
      const token = json.data?.onboarding_token ?? json.onboarding_token;
      if (!token) {
        return NextResponse.redirect(
          localeHome(request, `/login?oauth_error=${encodeURIComponent("missing_onboarding_token")}`),
        );
      }
      const secure = process.env.NODE_ENV === "production";
      store.set(OAUTH_ONBOARDING_COOKIE, token, {
        httpOnly: true,
        sameSite: "lax",
        secure,
        path: "/",
        maxAge: 10 * 60,
      });
      return NextResponse.redirect(localeHome(request, "/onboarding/organization"));
    }

    const access = json.data?.access_token ?? json.access_token;
    const refresh = json.data?.refresh_token ?? json.refresh_token;
    const user = json.data?.user ?? (json as { user?: unknown }).user;
    if (!access || !refresh) {
      return NextResponse.redirect(
        localeHome(request, `/login?oauth_error=${encodeURIComponent("missing_tokens")}`),
      );
    }
    await clearAuthCookiesOnStore();
    await setAuthCookiesOnStore({ accessToken: access, refreshToken: refresh });
    if (user) {
      const secure = process.env.NODE_ENV === "production";
      store.set(USER_BOOT_COOKIE, JSON.stringify(user), {
        httpOnly: false,
        sameSite: "lax",
        secure,
        path: "/",
        maxAge: 120,
      });
    }
    return NextResponse.redirect(localeHome(request, "/dashboard"));
  } catch {
    return NextResponse.redirect(
      localeHome(request, `/login?oauth_error=${encodeURIComponent("oauth_callback_failed")}`),
    );
  }
}
