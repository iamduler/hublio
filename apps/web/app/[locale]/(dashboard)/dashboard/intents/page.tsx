import { getTranslations, setRequestLocale } from "next-intl/server";
import { PageHeader } from "@hublio/ui/common/page-header";
import { IntentsList } from "@/features/intents/components/intents-list";
import { RunIntent } from "@/features/intents/components/run-intent";

export default async function IntentsPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("intents");

  return (
    <div className="space-y-6">
      <PageHeader title={t("title")} description={t("subtitle")} />
      <IntentsList />
      <RunIntent />
    </div>
  );
}
