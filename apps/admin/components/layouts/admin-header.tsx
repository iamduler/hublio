"use client";

import { useTranslations } from "next-intl";
import { LogOut } from "lucide-react";
import { AppSidebarTrigger } from "@hublio/ui/common/app-sidebar";
import { Badge } from "@hublio/ui/ui/badge";
import { Button } from "@hublio/ui/ui/button";
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
    <header className="flex h-(--nav-h) items-center justify-between gap-3 border-b border-(--line) bg-(--white) px-3 md:gap-4 md:px-5">
      <div className="flex min-w-0 items-center gap-2 md:gap-3">
        <AppSidebarTrigger menuLabel={t("menu")} />
        {organization ? (
          <>
            <span className="truncate text-sm font-medium text-(--ink)">
              {organization.name}
            </span>
            <Badge variant="outline" className="hidden sm:inline-flex">
              {organization.status}
            </Badge>
          </>
        ) : (
          <span className="text-sm text-(--muted-clr)">{t("noOrg")}</span>
        )}
      </div>
      <div className="flex shrink-0 items-center gap-2">
        {user ? (
          <span className="hidden text-xs text-(--muted-clr) sm:inline">
            {user.email}
          </span>
        ) : null}
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => void onLogout()}
        >
          <LogOut size={14} />
          <span className="hidden sm:inline">{t("signOut")}</span>
        </Button>
      </div>
    </header>
  );
}
