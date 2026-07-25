import { setRequestLocale } from "next-intl/server";
import { ConnectionDetail } from "@/features/connections/components/connection-detail";

export default async function ConnectionDetailPage({
  params,
}: {
  params: Promise<{ locale: string; connectionId: string }>;
}) {
  const { locale, connectionId } = await params;
  setRequestLocale(locale);

  return <ConnectionDetail connectionId={connectionId} />;
}
