"use client";

import { useTranslations } from "next-intl";
import { GitBranch, Plus } from "lucide-react";
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
import { Button } from "@hublio/ui/ui/button";
import { Badge } from "@hublio/ui/ui/badge";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { EmptyState } from "@hublio/ui/ui/empty-state";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useSyncRoutes } from "../hooks";

export function SyncRoutesList() {
  const t = useTranslations("syncRoutes");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useSyncRoutes();

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
        icon={GitBranch}
        title={t("empty")}
        description={t("emptyBody")}
        action={
          <Button asChild size="sm">
            <Link href="/dashboard/sync-routes/new">
              <Plus size={14} />
              {t("create")}
            </Link>
          </Button>
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
            <TableHead>{t("columns.trigger")}</TableHead>
            <TableHead>{t("columns.resources")}</TableHead>
            <TableHead>{t("columns.status")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((route) => (
            <TableRow key={route.id}>
              <TableCell>
                <Link
                  href={`/dashboard/sync-routes/${route.id}`}
                  className="font-medium text-[var(--ink)] no-underline hover:text-primary"
                >
                  {route.name}
                </Link>
              </TableCell>
              <TableCell>
                <Badge variant="sky">{route.trigger_type}</Badge>
              </TableCell>
              <TableCell className="text-[var(--ink-2)]">
                {(route.resource_types ?? []).join(", ") || "—"}
              </TableCell>
              <TableCell>
                <StatusBadge status={route.status} />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Card>
  );
}
