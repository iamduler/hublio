import { setRequestLocale } from "next-intl/server";
import { ConnectorDetail } from "@/features/connectors/components/connector-detail";

export default async function ConnectorDetailPage({
  params,
}: {
  params: Promise<{ locale: string; connectorId: string }>;
}) {
  const { locale, connectorId } = await params;
  setRequestLocale(locale);

  return <ConnectorDetail connectorId={connectorId} />;
}
