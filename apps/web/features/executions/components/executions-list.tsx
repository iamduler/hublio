"use client";

import { useTranslations } from "next-intl";
import { Activity } from "lucide-react";
import { Link } from "@/i18n/navigation";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@hublio/ui/ui/table";
import { Card } from "@hublio/ui/ui/card";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { EmptyState } from "@hublio/ui/ui/empty-state";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useExecutions } from "../hooks";

export function ExecutionsList() {
  const t = useTranslations("executions");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useExecutions({
    limit: 50,
  });

  if (isLoading) return <LoadingState rows={5} />;
  if (isError) {
    return (
      <ErrorState
        title={t("listLoadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }
  if (!data || data.length === 0) {
    return (
      <EmptyState
        icon={Activity}
        title={t("empty")}
        description={t("emptyBody")}
      />
    );
  }

  return (
    <Card className="overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("columns.id")}</TableHead>
            <TableHead>{t("columns.intent")}</TableHead>
            <TableHead>{t("columns.status")}</TableHead>
            <TableHead>{t("columns.step")}</TableHead>
            <TableHead>{t("columns.created")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((execution) => (
            <TableRow key={execution.id}>
              <TableCell>
                <Link
                  href={`/dashboard/executions/${execution.id}`}
                  className="font-mono text-sm font-medium text-[var(--ink)] no-underline hover:text-primary"
                >
                  {execution.id.slice(0, 8)}…
                </Link>
              </TableCell>
              <TableCell>
                <Link
                  href={`/dashboard/intents/${execution.intent_id}`}
                  className="font-mono text-xs text-[var(--ink-2)] no-underline hover:text-primary"
                >
                  {execution.intent_id.slice(0, 8)}…
                </Link>
              </TableCell>
              <TableCell>
                <StatusBadge status={execution.status} />
              </TableCell>
              <TableCell className="text-[var(--ink-2)]">
                {execution.current_step_no ?? "—"}
              </TableCell>
              <TableCell className="text-[var(--muted-clr)]">
                {execution.created_at
                  ? new Date(execution.created_at).toLocaleString()
                  : "—"}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Card>
  );
}
