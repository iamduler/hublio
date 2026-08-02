import { type NextRequest, NextResponse } from "next/server";
import createIntlMiddleware from "next-intl/middleware";
import { REFRESH_COOKIE, SESSION_COOKIE } from "@/lib/auth";
import { routing } from "@/i18n/routing";
import { isLocale } from "@/lib/i18n/config";

const intlMiddleware = createIntlMiddleware(routing);

function stripLocale(pathname: string): { locale: string; pathname: string } {
  const segment = pathname.split("/")[1];

  if (isLocale(segment)) {
    const rest = pathname.slice(segment.length + 1) || "/";
    return { locale: segment, pathname: rest };
  }

  return { locale: routing.defaultLocale, pathname };
}

const PUBLIC_PATHS = new Set([
  "/login",
  "/forbidden",
  "/forgot-password",
  "/reset-password",
]);

function isPublicPath(pathname: string): boolean {
  return PUBLIC_PATHS.has(pathname);
}

/**
 * Soft-gate: console routes require a session cookie.
 * Accept access (`hublio_session`) OR refresh (`hublio_refresh`) — access TTL is
 * ~15m; refresh can rotate it via `/api/auth/session` and `/api/go`.
 * Platform-admin claim is enforced client-side after /auth/me.
 */
export function proxy(request: NextRequest) {
  const { locale, pathname } = stripLocale(request.nextUrl.pathname);

  if (!isPublicPath(pathname)) {
    const access = request.cookies.get(SESSION_COOKIE)?.value;
    const refresh = request.cookies.get(REFRESH_COOKIE)?.value;
    if (!access && !refresh) {
      const loginUrl = new URL(`/${locale}/login`, request.url);
      const search = request.nextUrl.search;
      loginUrl.searchParams.set(
        "callbackUrl",
        search ? `${pathname}${search}` : pathname,
      );
      return NextResponse.redirect(loginUrl);
    }
  }

  return intlMiddleware(request);
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|site.webmanifest|.*\\.(?:ico|png|jpg|jpeg|gif|svg|webp|webmanifest|js|css|woff2?|ttf|eot|map)$|api/).*)",
  ],
};
