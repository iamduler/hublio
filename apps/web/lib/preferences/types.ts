export type UserTheme = "system" | "light" | "dark";
export type UserTimeFormat = "24h" | "12h";
export type UserDateFormat = "DD/MM/YYYY" | "MM/DD/YYYY" | "YYYY-MM-DD";

export interface UserPreferences {
  theme: UserTheme;
  date_format: UserDateFormat | string;
  time_format: UserTimeFormat;
  timezone: string;
}

export const DEFAULT_PREFERENCES: UserPreferences = {
  theme: "system",
  date_format: "YYYY-MM-DD",
  time_format: "24h",
  timezone: "UTC",
};

export { THEME_OPTIONS } from "@hublio/ui/lib/theme";
