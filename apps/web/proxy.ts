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

export function proxy(request: NextRequest) {
  const { locale, pathname } = stripLocale(request.nextUrl.pathname);

  const isDashboard =
    pathname === "/dashboard" || pathname.startsWith("/dashboard/");

  if (isDashboard) {
    // Access cookie TTL ~15m; refresh (~7d) can rotate it via /api/auth/session + /api/go.
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
