"use client";

import { useTranslations } from "next-intl";
import { LogOut } from "lucide-react";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { ThemeSwitcher } from "@/components/theme-switcher";
import { WorkspaceSwitcher } from "@/components/layouts/workspace-switcher";
import { Button } from "@hublio/ui/ui/button";
import { useAuth } from "@/providers/auth-provider";
import { useRouter } from "@/i18n/navigation";

export function WorkspaceHeader({ title }: { title?: string }) {
  const t = useTranslations("common");
  const { user, logout } = useAuth();
  const router = useRouter();

  async function onLogout() {
    await logout();
    router.replace("/login");
  }

  return (
    <header className="flex h-[var(--nav-h)] items-center justify-between gap-4 border-b border-[var(--line)] bg-[var(--white)] px-5">
      <div className="flex items-center gap-4">
        <WorkspaceSwitcher />
        {title ? (
          <h1 className="text-base font-semibold text-[var(--ink)]">{title}</h1>
        ) : null}
      </div>
      <div className="flex items-center gap-2">
        {user ? (
          <span className="hidden text-xs text-[var(--muted-clr)] sm:inline">
            {user.email}
          </span>
        ) : null}
        <ThemeSwitcher variant="shell" />
        <LocaleSwitcher variant="shell" />
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => void onLogout()}
        >
          <LogOut size={14} />
          {t("nav.signOut")}
        </Button>
      </div>
    </header>
  );
}
