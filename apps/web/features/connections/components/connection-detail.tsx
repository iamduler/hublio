"use client";

import { useTranslations } from "next-intl";
import { CheckCircle2, Power, RefreshCw } from "lucide-react";
import { Card, CardContent } from "@hublio/ui/ui/card";
import { Button } from "@hublio/ui/ui/button";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { EnvBadge } from "@hublio/ui/common/env-badge";
import { PageHeader } from "@hublio/ui/common/page-header";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useConnection, useConnectionAction } from "../hooks";

export function ConnectionDetail({ connectionId }: { connectionId: string }) {
  const t = useTranslations("connections");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } =
    useConnection(connectionId);
  const action = useConnectionAction();

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

  async function run(kind: "verify" | "enable" | "disable" | "rotate") {
    try {
      await action.mutateAsync({ id: connectionId, action: kind });
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
        description={data.description}
        actions={
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={action.isPending}
              onClick={() => void run("verify")}
            >
              <CheckCircle2 size={14} />
              {t("actions.verify")}
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={action.isPending}
              onClick={() => void run("rotate")}
            >
              <RefreshCw size={14} />
              {t("actions.rotate")}
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
          </div>
        }
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardContent className="p-4">
            <p className="text-xs font-medium uppercase tracking-wide text-[var(--muted-clr)]">
              {t("columns.environment")}
            </p>
            <div className="mt-1">
              <EnvBadge environment={data.environment} />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs font-medium uppercase tracking-wide text-[var(--muted-clr)]">
              {t("detail.timeout")}
            </p>
            <p className="mt-1 text-sm text-[var(--ink)]">
              {data.timeout_seconds ?? 30}s
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs font-medium uppercase tracking-wide text-[var(--muted-clr)]">
              {t("detail.updated")}
            </p>
            <p className="mt-1 text-sm text-[var(--ink)]">
              <FormattedDate value={data.updated_at} relative />
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
