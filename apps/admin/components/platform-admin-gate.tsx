"use client";

import { useEffect, type ReactNode } from "react";
import { useAuth } from "@/providers/auth-provider";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { usePathname, useRouter } from "@/i18n/navigation";

/**
 * Console routes require an authenticated platform admin.
 * Soft-gate in proxy.ts only checks session cookie; this enforces the claim.
 */
export function PlatformAdminGate({ children }: { children: ReactNode }) {
  const { user, isAuthenticated, isLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (isLoading) return;
    if (!isAuthenticated) {
      const callback = pathname && pathname !== "/" ? pathname : "/";
      router.replace(`/login?callbackUrl=${encodeURIComponent(callback)}`);
      return;
    }
    if (!user?.is_platform_admin) {
      router.replace("/forbidden");
    }
  }, [isLoading, isAuthenticated, user, router, pathname]);

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center p-6">
        <LoadingState rows={4} />
      </div>
    );
  }

  if (!isAuthenticated || !user?.is_platform_admin) {
    return null;
  }

  return <>{children}</>;
}
