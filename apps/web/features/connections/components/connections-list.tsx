"use client";

import { useTranslations } from "next-intl";
import { Cable, Plus } from "lucide-react";
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
import { buttonVariants } from "@hublio/ui/ui/button";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { EnvBadge } from "@hublio/ui/common/env-badge";
import { EmptyState } from "@hublio/ui/ui/empty-state";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useConnections } from "../hooks";

export function ConnectionsList() {
  const t = useTranslations("connections");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useConnections();

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
    return (
      <EmptyState
        icon={Cable}
        title={t("empty")}
        description={t("emptyBody")}
        action={
          <Link
            href="/dashboard/connections/new"
            className={buttonVariants({ size: "sm" })}
          >
            <Plus size={14} />
            {t("create")}
          </Link>
        }
      />
    );
  }

  return (
    <Card className="overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("columns.name")}</TableHead>
            <TableHead>{t("columns.environment")}</TableHead>
            <TableHead>{t("columns.status")}</TableHead>
            <TableHead>{t("columns.default")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((connection) => (
            <TableRow key={connection.id}>
              <TableCell>
                <Link
                  href={`/dashboard/connections/${connection.id}`}
                  className="font-medium text-(--ink) no-underline hover:text-primary"
                >
                  {connection.name}
                </Link>
                {connection.description ? (
                  <div className="text-xs text-(--muted-clr)">
                    {connection.description}
                  </div>
                ) : null}
              </TableCell>
              <TableCell>
                <EnvBadge environment={connection.environment} />
              </TableCell>
              <TableCell>
                <StatusBadge status={connection.status} />
              </TableCell>
              <TableCell className="text-(--ink-2)">
                {connection.is_default ? "★" : "—"}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Card>
  );
}
