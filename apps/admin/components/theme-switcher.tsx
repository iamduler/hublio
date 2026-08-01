"use client";

import { useTranslations } from "next-intl";
import { ThemeSwitcher as ThemeSwitcherUI } from "@hublio/ui/common/theme-switcher";
import { THEME_OPTIONS, type UserTheme } from "@hublio/ui/lib/theme";
import { usePreferredTheme } from "@/hooks/use-preferred-theme";

type Variant = "topbar" | "auth" | "shell";

export function ThemeSwitcher({ variant = "shell" }: { variant?: Variant }) {
  const t = useTranslations("common.settings.appearance.themes");
  const { theme, setTheme, isReady } = usePreferredTheme();

  return (
    <ThemeSwitcherUI
      theme={theme}
      isReady={isReady}
      variant={variant}
      fallbackLabel={t("system")}
      options={THEME_OPTIONS.map((option) => ({
        value: option,
        label: t(option),
      }))}
      onSelect={async (next: UserTheme) => {
        await setTheme(next);
      }}
    />
  );
}
