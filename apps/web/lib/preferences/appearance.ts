import { THEME_OPTIONS, type UserTheme } from "@/lib/preferences/types";

export const GUEST_THEME_STORAGE_KEY = "hublio_preferred_theme";

export const DEFAULT_APPEARANCE_THEME: UserTheme = "system";

export function isUserTheme(value: string | null | undefined): value is UserTheme {
  return THEME_OPTIONS.includes(value as UserTheme);
}

export function readGuestPreferredTheme(): UserTheme | null {
  if (typeof window === "undefined") return null;
  try {
    const value = window.localStorage.getItem(GUEST_THEME_STORAGE_KEY);
    if (!value || !isUserTheme(value)) return null;
    return value;
  } catch {
    return null;
  }
}

export function writeGuestPreferredTheme(theme: UserTheme): void {
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(GUEST_THEME_STORAGE_KEY, theme);
  } catch {
    // ignore
  }
}

export function resolveGuestTheme(): UserTheme {
  return readGuestPreferredTheme() ?? DEFAULT_APPEARANCE_THEME;
}
