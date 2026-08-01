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
  resolveGuestTheme,
  writeGuestPreferredTheme,
  type UserTheme,
} from "@hublio/ui/lib/theme";

interface PreferencesContextValue {
  theme: UserTheme;
  isLoading: boolean;
  setTheme: (theme: UserTheme) => void;
}

const PreferencesContext = createContext<PreferencesContextValue | null>(null);

export function PreferencesProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<UserTheme>("system");
  const [isLoading, setIsLoading] = useState(true);

  const applyTheme = useCallback((next: UserTheme) => {
    setThemeState(next);
    applyThemeToDocument(next);
  }, []);

  useEffect(() => {
    applyTheme(resolveGuestTheme());
    setIsLoading(false);
  }, [applyTheme]);

  useEffect(() => {
    if (theme !== "system") return;
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = () => applyThemeToDocument("system");
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, [theme]);

  const setTheme = useCallback(
    (next: UserTheme) => {
      writeGuestPreferredTheme(next);
      applyTheme(next);
    },
    [applyTheme],
  );

  const value = useMemo(
    () => ({ theme, isLoading, setTheme }),
    [theme, isLoading, setTheme],
  );

  return (
    <PreferencesContext.Provider value={value}>
      {children}
    </PreferencesContext.Provider>
  );
}

export function usePreferences(): PreferencesContextValue {
  const ctx = useContext(PreferencesContext);
  if (!ctx) {
    throw new Error("usePreferences must be used within PreferencesProvider");
  }
  return ctx;
}
