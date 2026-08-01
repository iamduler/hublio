"use client";

import { useTranslations } from "next-intl";
import { PlayCircle } from "lucide-react";
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
import { useIntents } from "../hooks";

export function IntentsList() {
  const t = useTranslations("intents");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useIntents({ limit: 50 });

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
        icon={PlayCircle}
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
            <TableHead>{t("columns.capability")}</TableHead>
            <TableHead>{t("columns.status")}</TableHead>
            <TableHead>{t("columns.created")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((intent) => (
            <TableRow key={intent.id}>
              <TableCell>
                <Link
                  href={`/dashboard/intents/${intent.id}`}
                  className="font-medium text-(--ink) no-underline hover:text-primary"
                >
                  {intent.capability}
                </Link>
                <div className="font-mono text-xs text-(--muted-clr)">
                  {intent.id}
                </div>
              </TableCell>
              <TableCell>
                <StatusBadge status={intent.status} />
              </TableCell>
              <TableCell className="text-(--muted-clr)">
                {intent.created_at
                  ? new Date(intent.created_at).toLocaleString()
                  : "—"}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Card>
  );
}
