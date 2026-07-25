import { setRequestLocale } from "next-intl/server";
import { SyncRouteDetail } from "@/features/sync-routes/components/sync-route-detail";

export default async function SyncRouteDetailPage({
  params,
}: {
  params: Promise<{ locale: string; syncRouteId: string }>;
}) {
  const { locale, syncRouteId } = await params;
  setRequestLocale(locale);

  return <SyncRouteDetail syncRouteId={syncRouteId} />;
}
