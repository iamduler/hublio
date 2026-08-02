"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Plus, X } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { buttonVariants } from "@hublio/ui/ui/button";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { EnvBadge } from "@hublio/ui/common/env-badge";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { FilterSelect } from "@hublio/ui/common/filter-select";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useConnections } from "../hooks";
import type { Connection } from "../types";

const PAGE_SIZE = 10;

export function ConnectionsList() {
  const t = useTranslations("connections");
  const tTable = useTranslations("common.table");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useConnections();
  const [statusFilter, setStatusFilter] = useState("");
  const [envFilter, setEnvFilter] = useState("");
  const [page, setPage] = useState(1);

  const filtered = useMemo(() => {
    return (data ?? []).filter((c) => {
      const matchStatus =
        !statusFilter ||
        c.status.toLowerCase() === statusFilter.toLowerCase();
      const matchEnv =
        !envFilter ||
        c.environment.toLowerCase() === envFilter.toLowerCase();
      return matchStatus && matchEnv;
    });
  }, [data, statusFilter, envFilter]);

  const statusOptions = useMemo(() => {
    const set = new Set((data ?? []).map((c) => c.status));
    return Array.from(set).sort().map((s) => ({ value: s, label: s }));
  }, [data]);

  const envOptions = useMemo(() => {
    const set = new Set((data ?? []).map((c) => c.environment));
    return Array.from(set).sort().map((s) => ({ value: s, label: s }));
  }, [data]);

  const columns: DataTableColumn<Connection>[] = [
    {
      id: "name",
      header: t("columns.name"),
      sortable: true,
      sortAccessor: (row) => row.name,
      cell: (connection) => (
        <div>
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
        </div>
      ),
    },
    {
      id: "environment",
      header: t("columns.environment"),
      sortable: true,
      sortAccessor: (row) => row.environment,
      cell: (row) => <EnvBadge environment={row.environment} />,
    },
    {
      id: "status",
      header: t("columns.status"),
      sortable: true,
      sortAccessor: (row) => row.status,
      cell: (row) => <StatusBadge status={row.status} />,
    },
    {
      id: "default",
      header: t("columns.default"),
      sortable: true,
      sortAccessor: (row) => (row.is_default ? 1 : 0),
      cell: (row) => (
        <span className="text-(--ink-2)">{row.is_default ? "★" : "—"}</span>
      ),
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

  const hasFilters = Boolean(statusFilter || envFilter);

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
          href="/dashboard/connections/new"
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
          <FilterSelect
            placeholder={tTable("filterEnv")}
            value={envFilter}
            onChange={(v) => {
              setEnvFilter(v);
              setPage(1);
            }}
            options={envOptions}
          />
          {hasFilters ? (
            <button
              type="button"
              className="inline-flex h-8 items-center gap-1.5 rounded-lg px-3 text-xs font-medium text-primary hover:bg-primary/5"
              onClick={() => {
                setStatusFilter("");
                setEnvFilter("");
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
