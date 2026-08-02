"use client";

import { useTranslations } from "next-intl";
import { ExternalLink, Power } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { Badge } from "@hublio/ui/ui/badge";
import { Button } from "@hublio/ui/ui/button";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { PageHeader } from "@hublio/ui/common/page-header";
import { ConfirmDialog } from "@hublio/ui/common/confirm-dialog";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useConnector, useToggleConnector } from "../hooks";

export function ConnectorDetail({ connectorId }: { connectorId: string }) {
  const t = useTranslations("connectors");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } =
    useConnector(connectorId);
  const toggle = useToggleConnector();

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

  const isEnabled = data.status === "enabled";
  const canToggle = data.status === "enabled" || data.status === "disabled" || data.status === "registered";

  async function runToggle(enable: boolean) {
    try {
      await toggle.mutateAsync({ id: connectorId, enable });
      toast.success(t(enable ? "actions.enableDone" : "actions.disableDone"));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={
          <span className="flex items-center gap-3">
            {data.name}
            <StatusBadge status={data.status} />
          </span>
        }
        description={data.description ?? data.code}
        actions={
          <div className="flex items-center gap-2">
            {data.documentation_url ? (
              <a
                href={data.documentation_url}
                target="_blank"
                rel="noreferrer"
                className="flex items-center gap-1.5 text-sm text-primary no-underline hover:underline"
              >
                {t("docs")}
                <ExternalLink size={14} />
              </a>
            ) : null}
            {canToggle ? (
              isEnabled ? (
                <ConfirmDialog
                  trigger={
                    <Button
                      variant="danger-soft"
                      size="sm"
                      disabled={toggle.isPending}
                    >
                      <Power size={14} />
                      {t("actions.disable")}
                    </Button>
                  }
                  title={t("disableTitle")}
                  description={t("disableBody")}
                  confirmLabel={t("actions.disable")}
                  cancelLabel={t("actions.cancel")}
                  destructive
                  pending={toggle.isPending}
                  onConfirm={() => runToggle(false)}
                />
              ) : (
                <Button
                  size="sm"
                  disabled={toggle.isPending}
                  onClick={() => void runToggle(true)}
                >
                  <Power size={14} />
                  {t("actions.enable")}
                </Button>
              )
            ) : null}
          </div>
        }
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Detail label={t("columns.vendor")} value={data.vendor} />
        <Detail label={t("columns.category")} value={data.category} />
        <Detail label={t("columns.version")} value={data.version} mono />
        <Detail label={t("columns.code")} value={data.code} mono />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("capabilities")}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-2">
          {data.capabilities && data.capabilities.length > 0 ? (
            data.capabilities.map((cap) => (
              <Badge
                key={cap.code ?? cap.capability_code ?? cap.display_name}
                variant={cap.is_async ? "violet" : "sky"}
              >
                {cap.display_name}
                {cap.is_async ? " · async" : ""}
              </Badge>
            ))
          ) : (
            <p className="text-sm text-(--muted-clr)">
              {t("noCapabilities")}
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

function Detail({
  label,
  value,
  mono,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <Card>
      <CardContent className="p-4">
        <p className="text-xs font-medium uppercase tracking-wide text-(--muted-clr)">
          {label}
        </p>
        <p
          className={
            mono
              ? "text-sm text-(--ink)"
              : "text-sm text-(--ink)"
          }
        >
          {value}
        </p>
      </CardContent>
    </Card>
  );
}
