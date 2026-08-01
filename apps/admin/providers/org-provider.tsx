"use client";

import { createContext, useContext, type ReactNode } from "react";
import type { Organization } from "@/types/auth";
import { useAuth } from "@/providers/auth-provider";

interface OrgContextValue {
  organization: Organization | null;
}

const OrgContext = createContext<OrgContextValue | null>(null);

/** Org context from `/auth/me` (no org switcher until list API exists). */
export function OrgProvider({ children }: { children: ReactNode }) {
  const { organization } = useAuth();
  return (
    <OrgContext.Provider value={{ organization }}>{children}</OrgContext.Provider>
  );
}

export function useOrg(): OrgContextValue {
  const ctx = useContext(OrgContext);
  if (!ctx) throw new Error("useOrg must be used within OrgProvider");
  return ctx;
}
