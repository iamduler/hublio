"use client";

import { useState } from "react";
import {
  Activity,
  Building2,
  CheckCircle,
  Layers,
  Plug,
  RefreshCw,
  Server,
  Users,
  Zap,
} from "lucide-react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import { Button } from "@hublio/ui/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { EnvBadge } from "@hublio/ui/common/env-badge";
import { KpiCard } from "@hublio/ui/common/kpi-card";
import { PageHeader } from "@hublio/ui/common/page-header";
import { formatNumber, formatTime } from "@hublio/ui/lib/format";
import { useAuth } from "@/providers/auth-provider";
import { useOrganizations } from "@/features/organizations/hooks";

const UNAVAILABLE = "—";

function platformEnvironment(): string {
  if (process.env.NEXT_PUBLIC_PLATFORM_ENV) {
    return process.env.NEXT_PUBLIC_PLATFORM_ENV;
  }
  return process.env.NODE_ENV === "production" ? "production" : "development";
}

export function MissionControl() {
  const t = useTranslations("missionControl");
  const { user } = useAuth();
  const { data: orgs, isFetching, refetch, dataUpdatedAt } = useOrganizations();
  const [refreshing, setRefreshing] = useState(false);

  const orgCount = orgs?.length ?? 0;
  const env = platformEnvironment();
  const lastRefresh =
    dataUpdatedAt > 0 ? formatTime(dataUpdatedAt) : UNAVAILABLE;

  async function handleRefresh() {
    setRefreshing(true);
    try {
      await refetch();
    } finally {
      setRefreshing(false);
    }
  }

  const unavailableHint = t("unavailable");

  return (
    <div className="space-y-5">
      <PageHeader
        title={
          <span className="flex flex-wrap items-center gap-3">
            {t("title")}
            <EnvBadge environment={env} />
          </span>
        }
        description={t("description")}
        actions={
          <div className="flex flex-wrap items-center gap-3">
            <p className="text-xs text-(--muted-clr)">
              {t("lastRefreshed")}{" "}
              <span className="text-(--ink-2)">{lastRefresh}</span>
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              disabled={refreshing || isFetching}
              onClick={() => void handleRefresh()}
            >
              <RefreshCw
                size={13}
                className={refreshing || isFetching ? "animate-spin" : ""}
              />
              {t("refresh")}
            </Button>
          </div>
        }
      />

      <div className="flex flex-col gap-3 rounded-xl border border-dashed border-(--line) bg-(--surface) px-5 py-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-start gap-3">
          <Activity size={16} className="mt-0.5 text-(--faint)" />
          <div>
            <p className="text-sm font-semibold text-(--ink)">
              {t("health.title")}
            </p>
            <p className="mt-0.5 text-xs text-(--muted-clr)">
              {t("health.body")}
            </p>
          </div>
        </div>
        <p className="text-xs font-medium text-(--faint)">{unavailableHint}</p>
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <KpiCard
          label={t("kpi.organizations")}
          value={formatNumber(orgCount)}
          icon={Building2}
          hint={user?.email ? t("kpi.signedIn", { email: user.email }) : undefined}
        />
        <KpiCard
          label={t("kpi.workspaces")}
          value={UNAVAILABLE}
          icon={Layers}
          hint={unavailableHint}
          unavailable
        />
        <KpiCard
          label={t("kpi.activeUsers")}
          value={UNAVAILABLE}
          icon={Users}
          hint={unavailableHint}
          unavailable
        />
        <KpiCard
          label={t("kpi.connections")}
          value={UNAVAILABLE}
          icon={Plug}
          hint={unavailableHint}
          unavailable
        />
        <KpiCard
          label={t("kpi.executions")}
          value={UNAVAILABLE}
          icon={Zap}
          hint={unavailableHint}
          unavailable
        />
        <KpiCard
          label={t("kpi.failed")}
          value={UNAVAILABLE}
          icon={Activity}
          hint={unavailableHint}
          unavailable
        />
        <KpiCard
          label={t("kpi.apiRequests")}
          value={UNAVAILABLE}
          icon={Server}
          hint={unavailableHint}
          unavailable
        />
        <KpiCard
          label={t("kpi.queueDepth")}
          value={UNAVAILABLE}
          icon={Server}
          hint={unavailableHint}
          unavailable
        />
      </div>

      <div className="grid grid-cols-1 gap-5 lg:grid-cols-12">
        <div className="space-y-5 lg:col-span-8">
          <Card className="overflow-hidden shadow-sm">
            <CardHeader className="flex flex-row items-center justify-between border-b border-(--line) px-5 py-3.5">
              <CardTitle className="flex items-center gap-2 text-[13px] font-semibold">
                <Activity size={14} className="text-(--muted-clr)" />
                {t("runtime.title")}
              </CardTitle>
              <span className="text-[11px] text-(--faint)">
                {unavailableHint}
              </span>
            </CardHeader>
            <CardContent className="divide-y divide-(--line) p-0">
              {(
                [
                  "successRate",
                  "avgTime",
                  "retries",
                  "dlq",
                  "processingRate",
                ] as const
              ).map((key) => (
                <div
                  key={key}
                  className="flex items-center justify-between px-5 py-3"
                >
                  <span className="text-[13px] text-(--ink-2)">
                    {t(`runtime.${key}`)}
                  </span>
                  <span className="text-sm text-(--faint)">
                    {UNAVAILABLE}
                  </span>
                </div>
              ))}
            </CardContent>
          </Card>

          <Card className="overflow-hidden shadow-sm">
            <CardHeader className="flex flex-row items-center justify-between border-b border-(--line) px-5 py-3.5">
              <CardTitle className="flex items-center gap-2 text-[13px] font-semibold">
                <Server size={14} className="text-(--muted-clr)" />
                {t("workers.title")}
              </CardTitle>
              <span className="text-[11px] text-(--faint)">
                {unavailableHint}
              </span>
            </CardHeader>
            <CardContent className="py-10 text-center">
              <p className="text-sm text-(--muted-clr)">{t("workers.empty")}</p>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-5 lg:col-span-4">
          <Card className="overflow-hidden shadow-sm">
            <CardHeader className="border-b border-(--line) px-5 py-3.5">
              <CardTitle className="flex items-center gap-2 text-[13px] font-semibold">
                <CheckCircle size={14} className="text-(--muted-clr)" />
                {t("alerts.title")}
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col items-center py-8 text-center">
              <CheckCircle size={22} className="mb-2 text-(--faint)" />
              <p className="text-[13px] font-medium text-(--ink-2)">
                {t("alerts.empty")}
              </p>
              <p className="mt-1 text-xs text-(--faint)">{unavailableHint}</p>
            </CardContent>
          </Card>

          <Card className="overflow-hidden shadow-sm">
            <CardHeader className="border-b border-(--line) px-5 py-3.5">
              <CardTitle className="flex items-center gap-2 text-[13px] font-semibold">
                <Zap size={14} className="text-(--muted-clr)" />
                {t("quickActions.title")}
              </CardTitle>
            </CardHeader>
            <CardContent className="grid grid-cols-2 gap-2 p-3 sm:grid-cols-4 lg:grid-cols-2">
              <QuickAction
                href="/organizations"
                icon={<Building2 size={15} />}
                label={t("quickActions.orgs")}
              />
              <QuickAction
                href="/connectors"
                icon={<Plug size={15} />}
                label={t("quickActions.connectors")}
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function QuickAction({
  href,
  icon,
  label,
}: {
  href: string;
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <Link
      href={href}
      className="flex flex-col items-center gap-1.5 rounded-lg border border-(--line) px-2 py-3 text-center no-underline transition-colors hover:border-primary/40 hover:bg-(--surface)"
    >
      <span className="text-(--muted-clr)">{icon}</span>
      <span className="text-[12px] font-medium text-(--ink-2)">{label}</span>
    </Link>
  );
}
