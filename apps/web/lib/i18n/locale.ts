import { defaultLocale, isLocale } from "./config";

export function getLocaleFromPathname(pathname: string): string {
  const segment = pathname.split("/")[1];
  return isLocale(segment) ? segment : defaultLocale;
}

export function getBrowserLocale(): string {
  if (typeof window === "undefined") {
    return defaultLocale;
  }

  return getLocaleFromPathname(window.location.pathname);
}
