"use client";

import { useTranslations } from "next-intl";
import { ExternalLink } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { Badge } from "@hublio/ui/ui/badge";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { PageHeader } from "@hublio/ui/common/page-header";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useConnector } from "../hooks";

export function ConnectorDetail({ connectorId }: { connectorId: string }) {
  const t = useTranslations("connectors");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } =
    useConnector(connectorId);

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
          data.documentation_url ? (
            <a
              href={data.documentation_url}
              target="_blank"
              rel="noreferrer"
              className="flex items-center gap-1.5 text-sm text-primary no-underline hover:underline"
            >
              {t("docs")}
              <ExternalLink size={14} />
            </a>
          ) : null
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
            <p className="text-sm text-[var(--muted-clr)]">
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
        <p className="text-xs font-medium uppercase tracking-wide text-[var(--muted-clr)]">
          {label}
        </p>
        <p
          className={
            mono
              ? "font-mono text-sm text-[var(--ink)]"
              : "text-sm text-[var(--ink)]"
          }
        >
          {value}
        </p>
      </CardContent>
    </Card>
  );
}
