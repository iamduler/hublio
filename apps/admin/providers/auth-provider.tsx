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
import type { AuthSession, Organization, User } from "@/types/auth";
import { authApi, type AuthTokenData } from "@/lib/api/auth";
import {
  clearAuthCookies,
  persistUser,
  setAuthPresentCookie,
} from "@/lib/auth";

interface AuthContextValue extends AuthSession {
  establishSession: (data: AuthTokenData) => Promise<User>;
  setOrganization: (org: Organization | null) => void;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [hasToken, setHasToken] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const establishSession = useCallback(async (data: AuthTokenData) => {
    setAuthPresentCookie();
    persistUser(data.user);
    setUser(data.user);
    setHasToken(true);
    try {
      const me = await authApi.me();
      setOrganization(me.organization);
      persistUser(me.user);
      setUser(me.user);
      return me.user;
    } catch {
      setOrganization(null);
      return data.user;
    }
  }, []);

  const logout = useCallback(async () => {
    await authApi.logout();
    clearAuthCookies();
    setUser(null);
    setOrganization(null);
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
          setOrganization(null);
          setHasToken(false);
          return;
        }

        const me = await authApi.me();
        if (cancelled) return;
        setAuthPresentCookie();
        persistUser(me.user);
        setUser(me.user);
        setOrganization(me.organization);
        setHasToken(true);
      } catch {
        if (cancelled) return;
        clearAuthCookies();
        setUser(null);
        setOrganization(null);
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
      organization,
      isAuthenticated: hasToken,
      isLoading,
      establishSession,
      setOrganization,
      logout,
    }),
    [
      user,
      organization,
      hasToken,
      isLoading,
      establishSession,
      logout,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
