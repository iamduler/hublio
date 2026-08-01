"use client";

import { useTranslations } from "next-intl";
import { LogOut } from "lucide-react";
import { AppHeader } from "@hublio/ui/common/app-shell";
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
    <AppHeader
      menuLabel={t("nav.menu")}
      email={user?.email}
      leading={
        <>
          <WorkspaceSwitcher />
          {title ? (
            <h1 className="truncate text-base font-semibold text-(--ink)">
              {title}
            </h1>
          ) : null}
        </>
      }
      trailing={
        <>
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
        </>
      }
    />
  );
}
