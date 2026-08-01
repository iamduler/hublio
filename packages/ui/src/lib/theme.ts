export type UserTheme = "system" | "light" | "dark";

export const THEME_OPTIONS: UserTheme[] = ["system", "light", "dark"];

export const GUEST_THEME_STORAGE_KEY = "hublio_preferred_theme";

export const DEFAULT_APPEARANCE_THEME: UserTheme = "system";

export function isUserTheme(
  value: string | null | undefined,
): value is UserTheme {
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

export function resolveEffectiveTheme(theme: UserTheme): "light" | "dark" {
  if (theme === "light" || theme === "dark") return theme;
  if (typeof window === "undefined") return "light";
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light";
}

export function applyThemeToDocument(theme: UserTheme): void {
  if (typeof document === "undefined") return;
  const effective = resolveEffectiveTheme(theme);
  const root = document.documentElement;
  root.dataset.theme = theme;
  root.classList.toggle("dark", effective === "dark");
  root.style.colorScheme = effective;
}
