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
import type { AuthSession, User } from "@/types/auth";
import { authApi, type AuthTokenData } from "@/lib/api/auth";
import {
  SESSION_COOKIE,
  clearAuthCookies,
  persistUser,
  readCookie,
  readPersistedUser,
  setAuthCookies,
} from "@/lib/auth";

interface AuthContextValue extends AuthSession {
  establishSession: (data: AuthTokenData) => Promise<User>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [hasToken, setHasToken] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const establishSession = useCallback(async (data: AuthTokenData) => {
    setAuthCookies({
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
    });
    persistUser(data.user);
    setUser(data.user);
    setHasToken(true);
    return data.user;
  }, []);

  const logout = useCallback(async () => {
    await authApi.logout();
    clearAuthCookies();
    setUser(null);
    setHasToken(false);
  }, []);

  useEffect(() => {
    const token = readCookie(SESSION_COOKIE);
    if (!token) {
      setUser(null);
      setHasToken(false);
      setIsLoading(false);
      return;
    }

    // JWT payload is encrypted — restore last login snapshot until a /me API exists.
    const cached = readPersistedUser();
    setUser(cached);
    setHasToken(true);
    setIsLoading(false);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isAuthenticated: hasToken && Boolean(user),
      isLoading,
      establishSession,
      logout,
    }),
    [user, hasToken, isLoading, establishSession, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
