import { getTranslations, setRequestLocale } from "next-intl/server";
import { Plus } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { buttonVariants } from "@hublio/ui/ui/button-variants";
import { PageHeader } from "@hublio/ui/common/page-header";
import { SyncRoutesList } from "@/features/sync-routes/components/sync-routes-list";

export default async function SyncRoutesPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("syncRoutes");

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("title")}
        description={t("subtitle")}
        actions={
          <Link
            href="/dashboard/sync-routes/new"
            className={buttonVariants({ size: "sm" })}
          >
            <Plus size={14} />
            {t("create")}
          </Link>
        }
      />
      <SyncRoutesList />
    </div>
  );
}
