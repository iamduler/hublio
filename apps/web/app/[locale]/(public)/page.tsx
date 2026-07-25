import { getTranslations, setRequestLocale } from "next-intl/server";
import { Link } from "@/i18n/navigation";
import { Button } from "@hublio/ui/ui/button";

export default async function LandingPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("common");

  return (
    <section className="landing-hero">
      <div className="relative z-10 mx-auto max-w-3xl text-center">
        <p className="eyebrow mb-4">{t("meta.brand")}</p>
        <h1 className="mb-4 text-4xl font-semibold tracking-tight text-[var(--ink)] md:text-5xl">
          {t("meta.brand")}
        </h1>
        <p className="mx-auto mb-8 max-w-xl text-base text-[var(--ink-2)] md:text-lg">
          {t("meta.description")}
        </p>
        <div className="flex flex-wrap items-center justify-center gap-3">
          <Button asChild size="lg">
            <Link href="/login">{t("nav.signIn")}</Link>
          </Button>
          <Button asChild size="lg" variant="outline">
            <Link href="/register">{t("nav.getStarted")}</Link>
          </Button>
        </div>
      </div>
    </section>
  );
}
