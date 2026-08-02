"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { X } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { FilterSelect } from "@hublio/ui/common/filter-select";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useExecutions } from "../hooks";
import type { Execution } from "../types";

const PAGE_SIZE = 10;

export function ExecutionsList() {
  const t = useTranslations("executions");
  const tTable = useTranslations("common.table");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useExecutions({
    limit: 50,
  });
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);

  const filtered = useMemo(() => {
    const list = data ?? [];
    if (!statusFilter) return list;
    return list.filter(
      (e) => e.status.toLowerCase() === statusFilter.toLowerCase(),
    );
  }, [data, statusFilter]);

  const statusOptions = useMemo(() => {
    const set = new Set((data ?? []).map((e) => e.status));
    return Array.from(set).sort().map((s) => ({ value: s, label: s }));
  }, [data]);

  const columns: DataTableColumn<Execution>[] = [
    {
      id: "id",
      header: t("columns.id"),
      sortable: true,
      sortAccessor: (row) => row.id,
      cell: (execution) => (
        <Link
          href={`/dashboard/executions/${execution.id}`}
          className="text-sm font-medium text-(--ink) no-underline hover:text-primary"
        >
          {execution.id.slice(0, 8)}…
        </Link>
      ),
    },
    {
      id: "intent",
      header: t("columns.intent"),
      sortable: true,
      sortAccessor: (row) => row.intent_id,
      cell: (execution) => (
        <Link
          href={`/dashboard/intents/${execution.intent_id}`}
          className="text-xs text-(--ink-2) no-underline hover:text-primary"
        >
          {execution.intent_id.slice(0, 8)}…
        </Link>
      ),
    },
    {
      id: "status",
      header: t("columns.status"),
      sortable: true,
      sortAccessor: (row) => row.status,
      cell: (row) => <StatusBadge status={row.status} />,
    },
    {
      id: "step",
      header: t("columns.step"),
      sortable: true,
      sortAccessor: (row) => row.current_step_no ?? 0,
      cell: (row) => (
        <span className="text-(--ink-2)">{row.current_step_no ?? "—"}</span>
      ),
    },
    {
      id: "created",
      header: t("columns.created"),
      sortable: true,
      sortAccessor: (row) => row.created_at ?? "",
      cell: (row) =>
        row.created_at ? <FormattedDate value={row.created_at} /> : "—",
    },
  ];

  if (isError) {
    return (
      <ErrorState
        title={t("listLoadError")}
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
