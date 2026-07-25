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
import { useAuth } from "@/providers/auth-provider";
import { useWorkspaces } from "@/features/workspaces/hooks";
import type { Workspace } from "@/features/workspaces/types";
import {
  readWorkspaceCookie,
  writeWorkspaceCookie,
} from "@/lib/workspace";

interface WorkspaceContextValue {
  workspaces: Workspace[];
  activeWorkspace: Workspace | null;
  activeWorkspaceId: string | null;
  isLoading: boolean;
  setActiveWorkspace: (workspaceId: string) => void;
}

const WorkspaceContext = createContext<WorkspaceContextValue | null>(null);

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const { user, isAuthenticated } = useAuth();
  const organizationId = user?.organization_id;
  const { data, isLoading } = useWorkspaces(
    isAuthenticated ? organizationId : undefined,
  );

  const workspaces = useMemo(() => data ?? [], [data]);
  const [activeId, setActiveId] = useState<string | null>(null);

  // Restore the persisted selection on mount.
  useEffect(() => {
    const stored = readWorkspaceCookie();
    if (stored) setActiveId(stored);
  }, []);

  // Reconcile the selection with the fetched list.
  useEffect(() => {
    if (workspaces.length === 0) return;
    const exists = activeId && workspaces.some((w) => w.id === activeId);
    if (!exists) {
      const fallback = workspaces[0]!.id;
      setActiveId(fallback);
      writeWorkspaceCookie(fallback);
    }
  }, [workspaces, activeId]);

  const setActiveWorkspace = useCallback((workspaceId: string) => {
    setActiveId(workspaceId);
    writeWorkspaceCookie(workspaceId);
  }, []);

  const value = useMemo<WorkspaceContextValue>(() => {
    const activeWorkspace =
      workspaces.find((w) => w.id === activeId) ?? null;
    return {
      workspaces,
      activeWorkspace,
      activeWorkspaceId: activeWorkspace?.id ?? null,
      isLoading,
      setActiveWorkspace,
    };
  }, [workspaces, activeId, isLoading, setActiveWorkspace]);

  return (
    <WorkspaceContext.Provider value={value}>
      {children}
    </WorkspaceContext.Provider>
  );
}

export function useWorkspace(): WorkspaceContextValue {
  const ctx = useContext(WorkspaceContext);
  if (!ctx) {
    throw new Error("useWorkspace must be used within WorkspaceProvider");
  }
  return ctx;
}

/**
 * Convenience for feature hooks that require an active workspace.
 * Returns the id or null; callers should disable queries when null.
 */
export function useActiveWorkspaceId(): string | null {
  return useWorkspace().activeWorkspaceId;
}
