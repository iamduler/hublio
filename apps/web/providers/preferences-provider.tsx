"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import {
  applyThemeToDocument,
  formatDateTimeWithPrefs,
  formatDateWithPrefs,
  formatTimeWithPrefs,
} from "@/lib/preferences/format";
import {
  DEFAULT_PREFERENCES,
  type UserPreferences,
} from "@/lib/preferences/types";
import { resolveGuestTheme, writeGuestPreferredTheme } from "@/lib/preferences/appearance";

interface PreferencesContextValue {
  preferences: UserPreferences;
  isLoading: boolean;
  setPreferencesLocal: (prefs: Partial<UserPreferences>) => void;
  setTheme: (theme: UserPreferences["theme"]) => void;
  formatDate: (value: Date | string | number) => string;
  formatTime: (value: Date | string | number) => string;
  formatDateTime: (value: Date | string | number) => string;
}

const PreferencesContext = createContext<PreferencesContextValue | null>(null);

function mergePrefs(
  base: UserPreferences,
  patch?: Partial<UserPreferences> | null,
): UserPreferences {
  if (!patch) return base;
  return { ...base, ...patch };
}

export function PreferencesProvider({ children }: { children: ReactNode }) {
  const [preferences, setPreferences] = useState<UserPreferences>(DEFAULT_PREFERENCES);
  const [isLoading, setIsLoading] = useState(true);

  const applyPrefs = useCallback((prefs: UserPreferences) => {
    setPreferences(prefs);
    applyThemeToDocument(prefs.theme);
  }, []);

  useEffect(() => {
    applyPrefs(
      mergePrefs(DEFAULT_PREFERENCES, {
        theme: resolveGuestTheme(),
      }),
    );
    setIsLoading(false);
  }, [applyPrefs]);

  useEffect(() => {
    if (preferences.theme !== "system") return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => applyThemeToDocument("system");
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, [preferences.theme]);

  const setPreferencesLocal = useCallback(
    (patch: Partial<UserPreferences>) => {
      applyPrefs(mergePrefs(preferences, patch));
    },
    [applyPrefs, preferences],
  );

  const setTheme = useCallback(
    (theme: UserPreferences["theme"]) => {
      writeGuestPreferredTheme(theme);
      applyPrefs(mergePrefs(preferences, { theme }));
    },
    [applyPrefs, preferences],
  );

  const value = useMemo<PreferencesContextValue>(
    () => ({
      preferences,
      isLoading,
      setPreferencesLocal,
      setTheme,
      formatDate: (v) => formatDateWithPrefs(v, preferences),
      formatTime: (v) => formatTimeWithPrefs(v, preferences),
      formatDateTime: (v) => formatDateTimeWithPrefs(v, preferences),
    }),
    [preferences, isLoading, setPreferencesLocal, setTheme],
  );

  return (
    <PreferencesContext.Provider value={value}>
      {children}
    </PreferencesContext.Provider>
  );
}

export function usePreferences(): PreferencesContextValue {
  const ctx = useContext(PreferencesContext);
  if (!ctx) throw new Error("usePreferences must be used within PreferencesProvider");
  return ctx;
}
