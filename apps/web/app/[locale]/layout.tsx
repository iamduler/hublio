import type { Metadata } from "next";
import { NextIntlClientProvider } from "next-intl";
import { getMessages, getTranslations, setRequestLocale } from "next-intl/server";
import { hasLocale } from "next-intl";
import { notFound } from "next/navigation";
import { routing } from "@/i18n/routing";
import { QueryProvider } from "@/providers/query-provider";
import { AuthProvider } from "@/providers/auth-provider";
import { WorkspaceProvider } from "@/providers/workspace-provider";
import { PreferencesProvider } from "@/providers/preferences-provider";
import { NavigationProgress } from "@/components/navigation-progress";
import { Toaster } from "@hublio/ui/ui/toaster";

export function generateStaticParams() {
  return routing.locales.map((locale) => ({ locale }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ locale: string }>;
}): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: "common" });

  return {
    title: {
      default: t("meta.title"),
      template: `%s — ${t("meta.brand")}`,
    },
    description: t("meta.description"),
  };
}

export default async function LocaleLayout({
  children,
  params,
}: Readonly<{
  children: React.ReactNode;
  params: Promise<{ locale: string }>;
}>) {
  const { locale } = await params;

  if (!hasLocale(routing.locales, locale)) {
    notFound();
  }

  setRequestLocale(locale);
  const messages = await getMessages();

  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      <QueryProvider>
        <AuthProvider>
          <WorkspaceProvider>
            <PreferencesProvider>
              <NavigationProgress />
              {children}
              <Toaster />
            </PreferencesProvider>
          </WorkspaceProvider>
        </AuthProvider>
      </QueryProvider>
    </NextIntlClientProvider>
  );
}
