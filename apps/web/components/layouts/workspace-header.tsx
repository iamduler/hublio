"use client";

import { useTranslations } from "next-intl";
import { LogOut } from "lucide-react";
import { AppSidebarTrigger } from "@hublio/ui/common/app-sidebar";
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
    <header className="flex h-(--nav-h) items-center justify-between gap-3 border-b border-(--line) bg-(--white) px-3 md:gap-4 md:px-5">
      <div className="flex min-w-0 items-center gap-2 md:gap-4">
        <AppSidebarTrigger menuLabel={t("nav.menu")} />
        <WorkspaceSwitcher />
        {title ? (
          <h1 className="truncate text-base font-semibold text-(--ink)">
            {title}
          </h1>
        ) : null}
      </div>
      <div className="flex shrink-0 items-center gap-2">
        {user ? (
          <span className="hidden text-xs text-(--muted-clr) sm:inline">
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
          <span className="hidden sm:inline">{t("nav.signOut")}</span>
        </Button>
      </div>
    </header>
  );
}
