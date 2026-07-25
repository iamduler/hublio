import type { User } from "@/types/auth";

/** Access token (Bearer). Set by the client after login. */
export const SESSION_COOKIE = "hublio_session";

/** Refresh token for POST /auth/logout (and future refresh). */
export const REFRESH_COOKIE = "hublio_refresh";

/** Cached user snapshot (JWT payload is encrypted server-side). */
export const USER_STORAGE_KEY = "hublio_user";

const ACCESS_MAX_AGE_DEFAULT = 15 * 60; // matches Go AccessTokenTTL
const REFRESH_MAX_AGE_DEFAULT = 7 * 24 * 60 * 60;

export function readCookie(name: string): string | undefined {
  if (typeof document === "undefined") return undefined;
  const match = document.cookie.match(
    new RegExp(`(?:^|; )${name.replace(/[$()*+.?[\\\]^{|}]/g, "\\$&")}=([^;]*)`),
  );
  return match ? decodeURIComponent(match[1]!) : undefined;
}

function writeCookie(name: string, value: string, maxAge: number): void {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${maxAge}; SameSite=Lax`;
}

function deleteCookie(name: string): void {
  if (typeof document === "undefined") return;
  document.cookie = `${name}=; path=/; max-age=0; SameSite=Lax`;
}

export function setAuthCookies(opts: {
  accessToken: string;
  refreshToken: string;
  accessMaxAge?: number;
  refreshMaxAge?: number;
}): void {
  writeCookie(
    SESSION_COOKIE,
    opts.accessToken,
    opts.accessMaxAge ?? ACCESS_MAX_AGE_DEFAULT,
  );
  writeCookie(
    REFRESH_COOKIE,
    opts.refreshToken,
    opts.refreshMaxAge ?? REFRESH_MAX_AGE_DEFAULT,
  );
}

export function clearAuthCookies(): void {
  deleteCookie(SESSION_COOKIE);
  deleteCookie(REFRESH_COOKIE);
  if (typeof window !== "undefined") {
    try {
      window.localStorage.removeItem(USER_STORAGE_KEY);
    } catch {
      // ignore
    }
  }
}

export function persistUser(user: User): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user));
  } catch {
    // ignore
  }
}

export function readPersistedUser(): User | null {
  if (typeof window === "undefined") return null;
  try {
    const raw = window.localStorage.getItem(USER_STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as User;
  } catch {
    return null;
  }
}
