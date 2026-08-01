"use client";

import { useLocale } from "next-intl";
import { LocaleSwitcher as LocaleSwitcherUI } from "@hublio/ui/common/locale-switcher";
import { usePathname, useRouter } from "@/i18n/navigation";
import { locales, type Locale } from "@/lib/i18n/config";

const LOCALE_LABELS: Record<Locale, string> = {
  en: "English",
  vi: "Tiếng Việt",
};

type Variant = "topbar" | "default" | "auth" | "shell";

export function LocaleSwitcher({ variant = "default" }: { variant?: Variant }) {
  const locale = useLocale() as Locale;
  const router = useRouter();
  const pathname = usePathname();

  function selectLocale(next: Locale) {
    if (next === locale) return;

    const qs =
      typeof window !== "undefined"
        ? window.location.search.replace(/^\?/, "")
        : "";
    if (!qs) {
      router.replace(pathname, { locale: next });
      return;
    }

    const query: Record<string, string | string[]> = {};
    new URLSearchParams(qs).forEach((value, key) => {
      const existing = query[key];
      if (existing === undefined) {
        query[key] = value;
      } else if (Array.isArray(existing)) {
        existing.push(value);
      } else {
        query[key] = [existing, value];
      }
    });

    router.replace({ pathname, query }, { locale: next });
  }

  return (
    <LocaleSwitcherUI
      locale={locale}
      locales={locales}
      labels={LOCALE_LABELS}
      onSelect={selectLocale}
      variant={variant}
    />
  );
}
