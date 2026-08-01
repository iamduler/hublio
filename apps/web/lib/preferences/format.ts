import type { UserDateFormat, UserPreferences, UserTimeFormat } from "./types";
import {
  applyThemeToDocument as applyThemeToDocumentShared,
  resolveEffectiveTheme as resolveEffectiveThemeShared,
  type UserTheme,
} from "@hublio/ui/lib/theme";

function dateParts(
  date: Date,
  timeZone: string,
): { year: string; month: string; day: string } {
  const parts = new Intl.DateTimeFormat("en-GB", {
    timeZone,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  }).formatToParts(date);

  const pick = (type: Intl.DateTimeFormatPartTypes) =>
    parts.find((p) => p.type === type)?.value ?? "";

  return { year: pick("year"), month: pick("month"), day: pick("day") };
}

export function formatDateWithPrefs(
  value: Date | string | number,
  prefs: Pick<UserPreferences, "date_format" | "timezone">,
): string {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return "";

  const { year, month, day } = dateParts(date, prefs.timezone);
  const format = prefs.date_format as UserDateFormat;

  switch (format) {
    case "MM/DD/YYYY":
      return `${month}/${day}/${year}`;
    case "DD/MM/YYYY":
      return `${day}/${month}/${year}`;
    case "YYYY-MM-DD":
    default:
      return `${year}-${month}-${day}`;
  }
}

export function formatTimeWithPrefs(
  value: Date | string | number,
  prefs: Pick<UserPreferences, "time_format" | "timezone">,
): string {
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) return "";

  const hour12 = (prefs.time_format as UserTimeFormat) === "12h";

  return new Intl.DateTimeFormat("en-GB", {
    timeZone: prefs.timezone,
    hour: "2-digit",
    minute: "2-digit",
    hour12,
  }).format(date);
}

export function formatDateTimeWithPrefs(
  value: Date | string | number,
  prefs: Pick<UserPreferences, "date_format" | "time_format" | "timezone">,
): string {
  const datePart = formatDateWithPrefs(value, prefs);
  const timePart = formatTimeWithPrefs(value, prefs);
  if (!datePart) return "";
  return `${datePart} ${timePart}`.trim();
}

export function resolveEffectiveTheme(
  theme: UserPreferences["theme"] | UserTheme,
): "light" | "dark" {
  return resolveEffectiveThemeShared(theme);
}

export function applyThemeToDocument(
  theme: UserPreferences["theme"] | UserTheme,
): void {
  applyThemeToDocumentShared(theme);
}
