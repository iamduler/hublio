import { setRequestLocale } from "next-intl/server";
import { SyncRouteEdit } from "@/features/sync-routes/components/sync-route-edit";

export default async function EditSyncRoutePage({
  params,
}: {
  params: Promise<{ locale: string; syncRouteId: string }>;
}) {
  const { locale, syncRouteId } = await params;
  setRequestLocale(locale);

  return <SyncRouteEdit syncRouteId={syncRouteId} />;
}
