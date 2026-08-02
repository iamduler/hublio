"use client";

import { useTranslations } from "next-intl";
import { LogOut } from "lucide-react";
import { AppHeader } from "@hublio/ui/common/app-shell";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { Button } from "@hublio/ui/ui/button";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { ThemeSwitcher } from "@/components/theme-switcher";
import { useAuth } from "@/providers/auth-provider";
import { useOrg } from "@/providers/org-provider";
import { useRouter } from "@/i18n/navigation";

export function AdminHeader() {
  const t = useTranslations("shell.header");
  const { user, logout } = useAuth();
  const { organization } = useOrg();
  const router = useRouter();

  async function onLogout() {
    await logout();
    router.replace("/login");
  }

  return (
    <AppHeader
      menuLabel={t("menu")}
      email={user?.email}
      leading={
        organization ? (
          <>
            <span className="truncate text-sm font-medium text-(--ink)">
              {organization.name}
            </span>
            <StatusBadge
              status={organization.status}
              className="hidden sm:inline-flex"
            />
          </>
        ) : (
          <span className="text-sm text-(--muted-clr)">{t("noOrg")}</span>
        )
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
            <span className="hidden sm:inline">{t("signOut")}</span>
          </Button>
        </>
      }
    />
  );
}
