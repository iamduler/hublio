import { hasLocale } from "next-intl";
import { getRequestConfig } from "next-intl/server";
import { loadMessages } from "@/lib/i18n/load-messages";
import type { Locale } from "@/lib/i18n/config";
import { routing } from "./routing";

export default getRequestConfig(async ({ requestLocale }) => {
  const requested = await requestLocale;
  const locale = hasLocale(routing.locales, requested)
    ? (requested as Locale)
    : routing.defaultLocale;

  return {
    locale,
    messages: await loadMessages(locale),
  };
});
