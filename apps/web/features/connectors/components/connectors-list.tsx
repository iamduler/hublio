"use client";

import { useTranslations } from "next-intl";
import { Blocks } from "lucide-react";
import { Link } from "@/i18n/navigation";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@hublio/ui/ui/table";
import { Badge } from "@hublio/ui/ui/badge";
import { Card } from "@hublio/ui/ui/card";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { EmptyState } from "@hublio/ui/ui/empty-state";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useConnectors } from "../hooks";

export function ConnectorsList() {
  const t = useTranslations("connectors");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useConnectors();

  if (isLoading) return <LoadingState rows={5} />;
  if (isError) {
    return (
      <ErrorState
        title={t("loadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }
  if (!data || data.length === 0) {
    return <EmptyState icon={Blocks} title={t("empty")} />;
  }

  return (
    <Card className="overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("columns.name")}</TableHead>
            <TableHead>{t("columns.vendor")}</TableHead>
            <TableHead>{t("columns.category")}</TableHead>
            <TableHead>{t("columns.version")}</TableHead>
            <TableHead>{t("columns.capabilities")}</TableHead>
            <TableHead>{t("columns.status")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((connector) => (
            <TableRow key={connector.id}>
              <TableCell>
                <Link
                  href={`/dashboard/connectors/${connector.id}`}
                  className="font-medium text-[var(--ink)] no-underline hover:text-primary"
                >
                  {connector.name}
                </Link>
                <div className="font-mono text-xs text-[var(--muted-clr)]">
                  {connector.code}
                </div>
              </TableCell>
              <TableCell className="text-[var(--ink-2)]">
                {connector.vendor}
              </TableCell>
              <TableCell>
                <Badge variant="gray">{connector.category}</Badge>
              </TableCell>
              <TableCell className="font-mono text-xs text-[var(--ink-2)]">
                {connector.version}
              </TableCell>
              <TableCell className="text-[var(--ink-2)]">
                {connector.capabilities?.length ?? 0}
              </TableCell>
              <TableCell>
                <StatusBadge status={connector.status} />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Card>
  );
}
