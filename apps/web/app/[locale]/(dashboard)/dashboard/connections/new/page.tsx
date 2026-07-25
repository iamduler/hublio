import { getTranslations, setRequestLocale } from "next-intl/server";
import { PageHeader } from "@hublio/ui/common/page-header";
import { ConnectionForm } from "@/features/connections/components/connection-form";

export default async function NewConnectionPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("connections");

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <PageHeader title={t("create")} description={t("createSubtitle")} />
      <ConnectionForm />
    </div>
  );
}
