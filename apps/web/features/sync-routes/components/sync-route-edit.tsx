"use client";

import { useTranslations } from "next-intl";
import { PageHeader } from "@hublio/ui/common/page-header";
import { Button } from "@hublio/ui/ui/button";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { Link } from "@/i18n/navigation";
import { useSyncRoute } from "../hooks";
import { SyncRouteForm } from "./sync-route-form";

export function SyncRouteEdit({ syncRouteId }: { syncRouteId: string }) {
  const t = useTranslations("syncRoutes");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } =
    useSyncRoute(syncRouteId);

  if (isLoading) return <LoadingState rows={4} />;
  if (isError || !data) {
    return (
      <ErrorState
        title={t("loadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }

  if (data.status === "enabled") {
    return (
      <div className="mx-auto max-w-lg space-y-4 py-12 text-center">
        <h2 className="text-lg font-semibold text-(--ink)">
          {t("editBlockedTitle")}
        </h2>
        <p className="text-sm text-(--muted-clr)">{t("editBlockedBody")}</p>
        <Button asChild>
          <Link href={`/dashboard/sync-routes/${syncRouteId}`}>
            {t("actions.disable")}
          </Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <PageHeader title={t("edit")} description={t("editSubtitle")} />
      <SyncRouteForm mode="edit" initial={data} />
    </div>
  );
}
