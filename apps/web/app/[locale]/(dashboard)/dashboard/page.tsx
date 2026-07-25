import { getTranslations, setRequestLocale } from "next-intl/server";
import { PageHeader } from "@hublio/ui/common/page-header";
import { DashboardOverview } from "@/features/dashboard/components/overview";

export default async function DashboardHomePage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("dashboard.home");

  return (
    <div className="space-y-6">
      <PageHeader title={t("title")} description={t("welcome")} />
      <DashboardOverview />
    </div>
  );
}
