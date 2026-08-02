"use client";

import { useTranslations } from "next-intl";
import { Ban, RefreshCw } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { Button } from "@hublio/ui/ui/button";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { PageHeader } from "@hublio/ui/common/page-header";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useExecution, useExecutionAction } from "../hooks";

const ACTIVE_STATUSES = new Set(["created", "queued", "running"]);
const RETRYABLE = new Set(["failed", "dead_letter", "cancelled", "expired"]);

export function ExecutionDetail({ executionId }: { executionId: string }) {
  const t = useTranslations("executions");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } =
    useExecution(executionId);
  const action = useExecutionAction(executionId);

  if (isLoading) return <LoadingState rows={5} />;
  if (isError || !data) {
    return (
      <ErrorState
        title={t("loadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }

  async function run(kind: "cancel" | "retry") {
    try {
      await action.mutateAsync(kind);
      toast.success(t(`actions.${kind}Done`));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  const steps = [...(data.steps ?? [])].sort((a, b) => a.step_no - b.step_no);
  const timeline = data.timeline ?? [];

  return (
    <div className="space-y-6">
      <PageHeader
        title={
          <span className="flex items-center gap-3">
            {t("title")}
            <StatusBadge status={data.status} />
          </span>
        }
        description={
          <Link
            href={`/dashboard/intents/${data.intent_id}`}
            className="text-xs text-primary no-underline hover:underline"
          >
            {t("fromIntent")}: {data.intent_id}
          </Link>
        }
        actions={
          <div className="flex items-center gap-2">
            {ACTIVE_STATUSES.has(data.status) ? (
              <Button
                variant="danger-soft"
                size="sm"
                disabled={action.isPending}
                onClick={() => void run("cancel")}
              >
                <Ban size={14} />
                {t("actions.cancel")}
              </Button>
            ) : null}
            {RETRYABLE.has(data.status) ? (
              <Button
                variant="outline"
                size="sm"
                disabled={action.isPending}
                onClick={() => void run("retry")}
              >
                <RefreshCw size={14} />
                {t("actions.retry")}
              </Button>
            ) : null}
          </div>
        }
      />

      {data.failure_reason ? (
        <div className="rounded-md border border-border-(--danger) bg-border-(--danger-soft) px-4 py-3 text-sm text-border-(--danger)">
          {data.failure_reason}
        </div>
      ) : null}

      <Card>
        <CardHeader>
          <CardTitle>{t("steps")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {steps.length === 0 ? (
            <p className="text-sm text-(--muted-clr)">{t("noSteps")}</p>
          ) : (
            steps.map((step) => (
              <div
                key={step.id}
                className="flex items-center justify-between rounded-md border border-(--line) px-3 py-2"
              >
                <div className="flex items-center gap-3">
                  <span className="grid h-6 w-6 place-items-center rounded-md bg-(--line-2) text-xs text-(--ink-2)">
                    {step.step_no}
                  </span>
                  <div>
                    <p className="text-sm font-medium text-(--ink)">
                      {step.step_type}
                    </p>
                    {step.error_message ? (
                      <p className="text-xs text-border-(--danger)">
                        {step.error_message}
                      </p>
                    ) : null}
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  {step.duration_ms != null ? (
                    <span className="text-xs text-(--muted-clr)">
                      {step.duration_ms}ms
                    </span>
                  ) : null}
                  <StatusBadge status={step.status} />
                </div>
              </div>
            ))
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("timeline")}</CardTitle>
        </CardHeader>
        <CardContent>
          {timeline.length === 0 ? (
            <p className="text-sm text-(--muted-clr)">{t("noTimeline")}</p>
          ) : (
            <ol className="space-y-3 border-l border-(--line) pl-4">
              {timeline.map((entry, index) => (
                <li key={entry.id ?? index} className="relative">
                  <span className="absolute -left-5.25 top-1 h-2.5 w-2.5 rounded-full bg-primary" />
                  <p className="text-sm font-medium text-(--ink)">
                    {entry.event}
                  </p>
                  {entry.message ? (
                    <p className="text-xs text-(--ink-2)">
                      {entry.message}
                    </p>
                  ) : null}
                  <p className="text-xs text-(--muted-clr)">
                    <FormattedDate value={entry.created_at} relative />
                  </p>
                </li>
              ))}
            </ol>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
