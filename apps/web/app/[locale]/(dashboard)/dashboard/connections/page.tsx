import { getTranslations, setRequestLocale } from "next-intl/server";
import { Plus } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { buttonVariants } from "@hublio/ui/ui/button-variants";
import { PageHeader } from "@hublio/ui/common/page-header";
import { ConnectionsList } from "@/features/connections/components/connections-list";

export default async function ConnectionsPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("connections");

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("title")}
        description={t("subtitle")}
        actions={
          <Link
            href="/dashboard/connections/new"
            className={buttonVariants({ size: "sm" })}
          >
            <Plus size={14} />
            {t("create")}
          </Link>
        }
      />
      <ConnectionsList />
    </div>
  );
}
