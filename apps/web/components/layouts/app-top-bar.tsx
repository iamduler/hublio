"use client";

import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import { LocaleSwitcher } from "@/components/locale-switcher";
import { ThemeSwitcher } from "@/components/theme-switcher";
import { buttonVariants } from "@hublio/ui/ui/button";

export function AppTopBar() {
  const t = useTranslations("common");

  return (
    <header className="sticky top-0 z-40 border-b border-[var(--line)] bg-[var(--slate-900)] text-white">
      <div className="mx-auto flex h-[var(--nav-h)] max-w-6xl items-center justify-between gap-4 px-4">
        <Link href="/" className="font-display text-lg font-semibold tracking-tight text-white no-underline">
          {t("meta.brand")}
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
