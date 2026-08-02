"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Plus, X } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { buttonVariants } from "@hublio/ui/ui/button";
import { Badge } from "@hublio/ui/ui/badge";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { FilterSelect } from "@hublio/ui/common/filter-select";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useSyncRoutes } from "../hooks";
import type { SyncRoute } from "../types";

const PAGE_SIZE = 10;

export function SyncRoutesList() {
  const t = useTranslations("syncRoutes");
  const tTable = useTranslations("common.table");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useSyncRoutes();
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);

  const filtered = useMemo(() => {
    const list = data ?? [];
    if (!statusFilter) return list;
    return list.filter(
      (r) => r.status.toLowerCase() === statusFilter.toLowerCase(),
    );
  }, [data, statusFilter]);

  const statusOptions = useMemo(() => {
    const set = new Set((data ?? []).map((r) => r.status));
    return Array.from(set).sort().map((s) => ({ value: s, label: s }));
  }, [data]);

  const columns: DataTableColumn<SyncRoute>[] = [
    {
      id: "name",
      header: t("columns.name"),
      sortable: true,
      sortAccessor: (row) => row.name,
      cell: (route) => (
        <Link
          href={`/dashboard/sync-routes/${route.id}`}
          className="font-medium text-(--ink) no-underline hover:text-primary"
        >
          {route.name}
        </Link>
      ),
    },
    {
      id: "trigger",
      header: t("columns.trigger"),
      sortable: true,
      sortAccessor: (row) => row.trigger_type,
      cell: (row) => <Badge variant="sky">{row.trigger_type}</Badge>,
    },
    {
      id: "resources",
      header: t("columns.resources"),
      sortable: true,
      sortAccessor: (row) => (row.resource_types ?? []).join(", "),
      cell: (row) => (
        <span className="text-(--ink-2)">
          {(row.resource_types ?? []).join(", ") || "—"}
        </span>
      ),
    },
    {
      id: "status",
      header: t("columns.status"),
      sortable: true,
      sortAccessor: (row) => row.status,
      cell: (row) => <StatusBadge status={row.status} />,
    },
  ];

  if (isError) {
    return (
      <ErrorState
        title={t("loadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <DataTable
      columns={columns}
      rows={filtered}
      getRowId={(row) => row.id}
      loading={isLoading}
      mode="client"
      page={page}
      pageSize={PAGE_SIZE}
      onPageChange={setPage}
      emptyTitle={t("empty")}
      emptyDescription={t("emptyBody")}
      emptyAction={
        <Link
          href="/dashboard/sync-routes/new"
          className={buttonVariants({ size: "sm" })}
        >
          <Plus size={14} />
          {t("create")}
        </Link>
      }
      prevLabel={tTable("prev")}
      nextLabel={tTable("next")}
      paginationShowingLabel={(from, to, total) =>
        tTable("showing", { from, to, total })
      }
      toolbar={
        <>
          <FilterSelect
            placeholder={tTable("filterStatus")}
            value={statusFilter}
            onChange={(v) => {
              setStatusFilter(v);
              setPage(1);
            }}
            options={statusOptions}
          />
          {statusFilter ? (
            <button
              type="button"
              className="inline-flex h-8 items-center gap-1.5 rounded-lg px-3 text-xs font-medium text-primary hover:bg-primary/5"
              onClick={() => {
                setStatusFilter("");
                setPage(1);
              }}
            >
              <X size={11} />
              {tTable("clearFilters")}
            </button>
          ) : null}
        </>
      }
    />
  );
}
