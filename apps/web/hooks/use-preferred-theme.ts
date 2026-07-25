"use client";

import { useMemo } from "react";
import { usePreferences } from "@/providers/preferences-provider";
import { resolveEffectiveTheme } from "@/lib/preferences/format";
import type { UserTheme } from "@/lib/preferences/types";

export function usePreferredTheme() {
  const { preferences, setTheme, isLoading } = usePreferences();

  const theme = preferences.theme;
  const effectiveTheme = useMemo(
    () => resolveEffectiveTheme(theme),
    [theme],
  );

  return {
    theme,
    effectiveTheme,
    setTheme: async (next: UserTheme) => {
      setTheme(next);
    },
    isReady: !isLoading,
  };
}
