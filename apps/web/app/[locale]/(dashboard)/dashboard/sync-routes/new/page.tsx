import { getTranslations, setRequestLocale } from "next-intl/server";
import { PageHeader } from "@hublio/ui/common/page-header";
import { SyncRouteForm } from "@/features/sync-routes/components/sync-route-form";

export default async function NewSyncRoutePage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("syncRoutes");

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <PageHeader title={t("create")} description={t("createSubtitle")} />
      <SyncRouteForm />
    </div>
  );
}
