import { getTranslations, setRequestLocale } from "next-intl/server";
import { PageHeader } from "@hublio/ui/common/page-header";
import { ExecutionsList } from "@/features/executions/components/executions-list";

export default async function ExecutionsPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("executions");

  return (
    <div className="space-y-6">
      <PageHeader title={t("listTitle")} description={t("listSubtitle")} />
      <ExecutionsList />
    </div>
  );
}
