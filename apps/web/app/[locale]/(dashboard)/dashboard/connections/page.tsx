import { getTranslations, setRequestLocale } from "next-intl/server";
import { Plus } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { Button } from "@hublio/ui/ui/button";
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
          <Button asChild size="sm">
            <Link href="/dashboard/connections/new">
              <Plus size={14} />
              {t("create")}
            </Link>
          </Button>
        }
      />
      <ConnectionsList />
    </div>
  );
}
