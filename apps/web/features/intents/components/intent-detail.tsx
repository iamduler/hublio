"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { PageHeader } from "@hublio/ui/common/page-header";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useIntent } from "../hooks";

export function IntentDetail({ intentId }: { intentId: string }) {
  const t = useTranslations("intents");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useIntent(intentId);

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
            {data.capability}
            <StatusBadge status={data.status} />
          </span>
        }
        description={
          <span className="font-mono text-xs">{data.id}</span>
        }
      />
      <div className="grid gap-4 sm:grid-cols-3">
        <Info label={t("detail.connection")} value={data.connection_id} mono />
        <Info
          label={t("detail.correlation")}
          value={data.correlation_id ?? "—"}
          mono
        />
        <Card>
          <CardContent className="p-4">
            <p className="text-xs font-medium uppercase tracking-wide text-(--muted-clr)">
              {t("detail.submitted")}
            </p>
            <p className="mt-1 text-sm text-(--ink)">
              <FormattedDate value={data.submitted_at ?? data.created_at} />
            </p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("detail.payload")}</CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="overflow-x-auto rounded-md bg-(--line-2) p-4 font-mono text-xs text-(--ink-2)">
            {JSON.stringify(data.payload ?? {}, null, 2)}
          </pre>
        </CardContent>
      </Card>
    </div>
  );
}

function Info({
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
              ? "mt-1 truncate font-mono text-xs text-(--ink)"
              : "mt-1 text-sm text-(--ink)"
          }
        >
          {value}
        </p>
      </CardContent>
    </Card>
  );
}
