import { setRequestLocale } from "next-intl/server";
import { IntentDetail } from "@/features/intents/components/intent-detail";

export default async function IntentDetailPage({
  params,
}: {
  params: Promise<{ locale: string; intentId: string }>;
}) {
  const { locale, intentId } = await params;
  setRequestLocale(locale);

  return <IntentDetail intentId={intentId} />;
}
