"use client";

import { useMemo } from "react";
import {
  resolveEffectiveTheme,
  type UserTheme,
} from "@hublio/ui/lib/theme";
import { usePreferences } from "@/providers/preferences-provider";

export function usePreferredTheme() {
  const { theme, setTheme, isLoading } = usePreferences();

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
