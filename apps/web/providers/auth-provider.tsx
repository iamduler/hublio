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
  clearAuthCookies,
  persistUser,
  readPersistedUser,
  setAuthPresentCookie,
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
    // httpOnly JWT cookies are already set by /api/auth/login.
    setAuthPresentCookie();
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
    let cancelled = false;
    void (async () => {
      try {
        const session = await authApi.session();
        if (cancelled) return;
        if (!session.authenticated) {
          clearAuthCookies();
          setUser(null);
          setHasToken(false);
          return;
        }
        setAuthPresentCookie();
        setUser(readPersistedUser());
        setHasToken(true);
      } catch {
        if (cancelled) return;
        clearAuthCookies();
        setUser(null);
        setHasToken(false);
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
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
