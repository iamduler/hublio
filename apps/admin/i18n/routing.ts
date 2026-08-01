import { defineRouting } from "next-intl/routing";
import { defaultLocale, locales } from "@/lib/i18n/config";

export const routing = defineRouting({
  locales,
  defaultLocale,
  localePrefix: "always",
  localeCookie: {
    name: "NEXT_LOCALE",
    maxAge: 60 * 60 * 24 * 365,
    sameSite: "lax",
  },
  localeDetection: true,
});
