"use client";

import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { ThemeSwitcher } from "@/components/theme-switcher";
import { buttonVariants } from "@hublio/ui/ui/button";
import { Logo } from "@/features/auth/auth-ui";

export function AppTopBar() {
  const t = useTranslations("common");

  return (
    <header className="sticky top-0 z-40 border-b border-(--line) bg-(--slate-900) text-white">
      <div className="mx-auto flex h-(--nav-h) max-w-6xl items-center justify-between gap-4 px-4">
        <Link href="/" className="no-underline">
          <Logo size="sm" />
        </Link>
        <div className="flex items-center gap-2">
          <ThemeSwitcher variant="topbar" />
          <LocaleSwitcher variant="topbar" />
          <Link
            href="/login"
            className={buttonVariants({ size: "sm", variant: "default" })}
          >
            {t("nav.signIn")}
          </Link>
        </div>
      </div>
    </header>
  );
}
