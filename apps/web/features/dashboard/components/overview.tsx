"use client";

import { useTranslations } from "next-intl";
import { Activity, Blocks, Cable, GitBranch } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { Badge } from "@hublio/ui/ui/badge";
import { KpiCard } from "@hublio/ui/common/kpi-card";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { EmptyState } from "@hublio/ui/ui/empty-state";
import { useConnections } from "@/features/connections/hooks";
import { useConnectors } from "@/features/connectors/hooks";
import { useSyncRoutes } from "@/features/sync-routes/hooks";
import { useEvents } from "@/features/events/hooks";

export function DashboardOverview() {
  const t = useTranslations("dashboard.overview");
  const connections = useConnections();
  const connectors = useConnectors();
  const syncRoutes = useSyncRoutes();
  const events = useEvents({ limit: 8 });

  const activeConnections =
    connections.data?.filter((c) => c.status === "active").length ?? 0;

  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <KpiCard
          label={t("kpis.connections")}
          value={connections.data?.length ?? "—"}
          hint={t("kpis.activeConnections", { count: activeConnections })}
          icon={Cable}
        />
        <KpiCard
          label={t("kpis.connectors")}
          value={connectors.data?.length ?? "—"}
          icon={Blocks}
        />
        <KpiCard
          label={t("kpis.syncRoutes")}
          value={syncRoutes.data?.length ?? "—"}
          icon={GitBranch}
        />
        <KpiCard
          label={t("kpis.recentEvents")}
          value={events.data?.length ?? "—"}
          icon={Activity}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("recentActivity")}</CardTitle>
          <Link
            href="/dashboard/events"
            className="text-sm text-primary no-underline hover:underline"
          >
            {t("viewAll")}
          </Link>
        </CardHeader>
        <CardContent>
          {events.isLoading ? (
            <LoadingState rows={4} />
          ) : !events.data || events.data.length === 0 ? (
            <EmptyState icon={Activity} title={t("noActivity")} size="sm" />
          ) : (
            <ul className="space-y-2">
              {events.data.map((event) => (
                <li
                  key={event.id}
                  className="flex items-center justify-between rounded-md border border-(--line) px-3 py-2"
                >
                  <div className="flex items-center gap-3">
                    <Badge variant="sky">{event.category}</Badge>
                    <span className="text-sm font-medium text-(--ink)">
                      {event.event_name}
                    </span>
                  </div>
                  <span className="text-xs text-(--muted-clr)">
                    <FormattedDate value={event.created_at} relative />
                  </span>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
