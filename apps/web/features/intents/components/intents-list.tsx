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
import { useIntents } from "../hooks";
import type { Intent } from "../types";

const PAGE_SIZE = 10;

export function IntentsList() {
  const t = useTranslations("intents");
  const tTable = useTranslations("common.table");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useIntents({ limit: 50 });
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);

  const filtered = useMemo(() => {
    const list = data ?? [];
    if (!statusFilter) return list;
    return list.filter(
      (i) => i.status.toLowerCase() === statusFilter.toLowerCase(),
    );
  }, [data, statusFilter]);

  const statusOptions = useMemo(() => {
    const set = new Set((data ?? []).map((i) => i.status));
    return Array.from(set).sort().map((s) => ({ value: s, label: s }));
  }, [data]);

  const columns: DataTableColumn<Intent>[] = [
    {
      id: "capability",
      header: t("columns.capability"),
      sortable: true,
      sortAccessor: (row) => row.capability,
      cell: (intent) => (
        <div>
          <Link
            href={`/dashboard/intents/${intent.id}`}
            className="font-medium text-(--ink) no-underline hover:text-primary"
          >
            {intent.capability}
          </Link>
          <div className="text-xs text-(--muted-clr)">{intent.id}</div>
        </div>
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
      id: "created",
      header: t("columns.created"),
      sortable: true,
      sortAccessor: (row) => row.created_at ?? "",
      cell: (row) =>
        row.created_at ? (
          <FormattedDate value={row.created_at} />
        ) : (
          "—"
        ),
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
