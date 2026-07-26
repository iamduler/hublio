import { type NextRequest, NextResponse } from "next/server";
import { cookies } from "next/headers";
import { goBaseUrl, proxyErrorResponse, ProxyError } from "@/lib/api/proxy-go";
import {
  OAUTH_PKCE_COOKIE,
  OAUTH_STATE_COOKIE,
  OAUTH_PROVIDER_COOKIE,
} from "@/lib/auth";

function oauthRedirectURI(request: NextRequest): string {
  const configured = process.env.OAUTH_REDIRECT_URI;
  if (configured) return configured;
  return `${request.nextUrl.origin}/api/auth/oauth/callback`;
}

function providerAuthorizeURL(
  provider: string,
  opts: { clientId: string; redirectURI: string; state: string; challenge: string },
): string | null {
  const params = new URLSearchParams({
    client_id: opts.clientId,
    redirect_uri: opts.redirectURI,
    response_type: "code",
    state: opts.state,
    code_challenge: opts.challenge,
    code_challenge_method: "S256",
  });

  switch (provider) {
    case "google":
      params.set("scope", "openid email profile");
      params.set("access_type", "online");
      params.set("prompt", "select_account");
      return `https://accounts.google.com/o/oauth2/v2/auth?${params}`;
    case "microsoft":
      params.set("scope", "openid email profile User.Read");
      params.set("response_mode", "query");
      return `https://login.microsoftonline.com/common/oauth2/v2.0/authorize?${params}`;
    case "github":
      params.set("scope", "read:user user:email");
      return `https://github.com/login/oauth/authorize?${params}`;
    default:
      return null;
  }
}

function clientIdFor(provider: string): string {
  switch (provider) {
    case "google":
      return process.env.GOOGLE_OAUTH_CLIENT_ID ?? "";
    case "microsoft":
      return process.env.MICROSOFT_OAUTH_CLIENT_ID ?? "";
    case "github":
      return process.env.GITHUB_OAUTH_CLIENT_ID ?? "";
    default:
      return "";
  }
}

function randomURLSafe(bytes = 32): string {
  const arr = new Uint8Array(bytes);
  crypto.getRandomValues(arr);
  return Buffer.from(arr).toString("base64url");
}

async function pkceChallenge(verifier: string): Promise<string> {
  const digest = await crypto.subtle.digest(
    "SHA-256",
    new TextEncoder().encode(verifier),
  );
  return Buffer.from(digest).toString("base64url");
}

/** GET /api/auth/oauth/start?provider=google */
export async function GET(request: NextRequest) {
  try {
    const provider = (request.nextUrl.searchParams.get("provider") || "").toLowerCase();
    const clientId = clientIdFor(provider);
    if (!clientId) {
      throw new ProxyError(400, "OAUTH_NOT_CONFIGURED", `OAuth provider ${provider} is not configured.`);
    }

    const state = randomURLSafe(16);
    const verifier = randomURLSafe(32);
    const challenge = await pkceChallenge(verifier);
    const redirectURI = oauthRedirectURI(request);
    const authorizeURL = providerAuthorizeURL(provider, {
      clientId,
      redirectURI,
      state,
      challenge,
    });
    if (!authorizeURL) {
      throw new ProxyError(400, "INVALID_PROVIDER", "Unsupported OAuth provider.");
    }

    const store = await cookies();
    const secure = process.env.NODE_ENV === "production";
    const cookieOpts = {
      httpOnly: true,
      sameSite: "lax" as const,
      secure,
      path: "/",
      maxAge: 10 * 60,
    };
    store.set(OAUTH_STATE_COOKIE, state, cookieOpts);
    store.set(OAUTH_PKCE_COOKIE, verifier, cookieOpts);
    store.set(OAUTH_PROVIDER_COOKIE, provider, cookieOpts);

    return NextResponse.redirect(authorizeURL);
  } catch (err) {
    return proxyErrorResponse(err);
  }
}
