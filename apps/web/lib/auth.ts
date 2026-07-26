import type { User } from "@/types/auth";

/** Access token (Bearer) — httpOnly; set only by Next `/api/auth/*` routes. */
export const SESSION_COOKIE = "hublio_session";

/** Refresh token — httpOnly; set only by Next `/api/auth/*` routes. */
export const REFRESH_COOKIE = "hublio_refresh";

/** OAuth PKCE / state cookies (httpOnly, short-lived). */
export const OAUTH_STATE_COOKIE = "hublio_oauth_state";
export const OAUTH_PKCE_COOKIE = "hublio_oauth_pkce";
export const OAUTH_PROVIDER_COOKIE = "hublio_oauth_provider";
export const OAUTH_ONBOARDING_COOKIE = "hublio_oauth_onboarding";

/** Cached user snapshot (JWT payload is encrypted server-side). */
export const USER_STORAGE_KEY = "hublio_user";

/** Short-lived non-httpOnly bootstrap after OAuth redirect (JSON user). */
export const USER_BOOT_COOKIE = "hublio_user_boot";

/**
 * Client-visible presence flag (NOT the JWT). Used by proxy.ts soft-gate and
 * AuthProvider while the real token stays httpOnly.
 */
export const AUTH_PRESENT_COOKIE = "hublio_auth";

const AUTH_PRESENT_MAX_AGE = 7 * 24 * 60 * 60;

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

/** Mark the browser as authenticated without exposing the JWT. */
export function setAuthPresentCookie(): void {
  writeCookie(AUTH_PRESENT_COOKIE, "1", AUTH_PRESENT_MAX_AGE);
}

export function clearAuthCookies(): void {
  deleteCookie(AUTH_PRESENT_COOKIE);
  // Best-effort: legacy JS-readable cookies from before httpOnly migration.
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
    const boot = readCookie(USER_BOOT_COOKIE);
    if (boot) {
      deleteCookie(USER_BOOT_COOKIE);
      const parsed = JSON.parse(boot) as User;
      persistUser(parsed);
      return parsed;
    }
    const raw = window.localStorage.getItem(USER_STORAGE_KEY);
    if (!raw) return null;
    return JSON.parse(raw) as User;
  } catch {
    return null;
  }
}
