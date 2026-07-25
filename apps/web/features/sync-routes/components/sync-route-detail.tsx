"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import { Power, RefreshCw, Trash2 } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { Button } from "@hublio/ui/ui/button";
import { Badge } from "@hublio/ui/ui/badge";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { PageHeader } from "@hublio/ui/common/page-header";
import { CopyValue } from "@hublio/ui/common/copy-value";
import { ConfirmDialog } from "@hublio/ui/common/confirm-dialog";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useRouter } from "@/i18n/navigation";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { queryKeys } from "@/lib/query-keys";
import { syncRoutesApi } from "../api";
import { useSyncRoute, useSyncRouteAction } from "../hooks";

export function SyncRouteDetail({ syncRouteId }: { syncRouteId: string }) {
  const t = useTranslations("syncRoutes");
  const getError = useApiErrorMessage();
  const router = useRouter();
  const workspaceId = useActiveWorkspaceId();
  const { data, isLoading, isError, error, refetch } =
    useSyncRoute(syncRouteId);
  const action = useSyncRouteAction();
  const [secret, setSecret] = useState<string | null>(null);

  const watermarks = useQuery({
    queryKey: [...queryKeys.syncRoute(workspaceId ?? "none", syncRouteId), "watermarks"],
    queryFn: () => syncRoutesApi.watermarks(workspaceId!, syncRouteId),
    enabled: Boolean(workspaceId && data),
  });

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

  async function run(kind: "enable" | "disable" | "remove" | "rotate") {
    try {
      const result = await action.mutateAsync({ id: syncRouteId, action: kind });
      if (kind === "remove") {
        toast.success(t("deleted"));
        router.replace("/dashboard/sync-routes");
        return;
      }
      if (kind === "rotate" && result && "webhook_secret" in result) {
        setSecret((result as { webhook_secret?: string }).webhook_secret ?? null);
      }
      toast.success(t(`actions.${kind}Done`));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  const isDisabled = data.status === "disabled";

  return (
    <div className="space-y-6">
      <PageHeader
        title={
          <span className="flex items-center gap-3">
            {data.name}
            <StatusBadge status={data.status} />
          </span>
        }
        description={
          <span className="flex items-center gap-2">
            <Badge variant="sky">{data.trigger_type}</Badge>
            <span>{(data.resource_types ?? []).join(", ")}</span>
          </span>
        }
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={action.isPending}
              onClick={() => void run("rotate")}
            >
              <RefreshCw size={14} />
              {t("actions.rotateSecret")}
            </Button>
            <Button
              variant={isDisabled ? "default" : "danger-soft"}
              size="sm"
              disabled={action.isPending}
              onClick={() => void run(isDisabled ? "enable" : "disable")}
            >
              <Power size={14} />
              {isDisabled ? t("actions.enable") : t("actions.disable")}
            </Button>
            <ConfirmDialog
              trigger={
                <Button variant="danger-soft" size="sm">
                  <Trash2 size={14} />
                  {t("actions.delete")}
                </Button>
              }
              title={t("deleteTitle")}
              description={t("deleteBody")}
              confirmLabel={t("actions.delete")}
              cancelLabel={t("form.cancel")}
              destructive
              pending={action.isPending}
              onConfirm={() => run("remove")}
            />
          </div>
        }
      />

      {secret ? (
        <Card>
          <CardHeader>
            <CardTitle>{t("webhookSecret")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <p className="text-sm text-[var(--ink-2)]">{t("secretOnce")}</p>
            <CopyValue value={secret} />
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>{t("watermarks")}</CardTitle>
        </CardHeader>
        <CardContent>
          {watermarks.isLoading ? (
            <LoadingState rows={2} />
          ) : !watermarks.data || watermarks.data.length === 0 ? (
            <p className="text-sm text-[var(--muted-clr)]">
              {t("noWatermarks")}
            </p>
          ) : (
            <ul className="space-y-2">
              {watermarks.data.map((wm) => (
                <li
                  key={wm.resource_type}
                  className="flex items-center justify-between rounded-md border border-[var(--line)] px-3 py-2 text-sm"
                >
                  <span className="font-medium text-[var(--ink)]">
                    {wm.resource_type}
                  </span>
                  <code className="font-mono text-xs text-[var(--muted-clr)]">
                    {JSON.stringify(wm.cursor ?? {})}
                  </code>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
