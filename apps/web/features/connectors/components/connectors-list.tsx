"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { X } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { Badge } from "@hublio/ui/ui/badge";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { FilterSelect } from "@hublio/ui/common/filter-select";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useConnectors } from "../hooks";
import type { Connector } from "../types";

const PAGE_SIZE = 10;

export function ConnectorsList() {
  const t = useTranslations("connectors");
  const tTable = useTranslations("common.table");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useConnectors();
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);

  const filtered = useMemo(() => {
    const list = data ?? [];
    if (!statusFilter) return list;
    return list.filter(
      (c) => c.status.toLowerCase() === statusFilter.toLowerCase(),
    );
  }, [data, statusFilter]);

  const statusOptions = useMemo(() => {
    const set = new Set((data ?? []).map((c) => c.status));
    return Array.from(set).sort().map((s) => ({ value: s, label: s }));
  }, [data]);

  const columns: DataTableColumn<Connector>[] = [
    {
      id: "name",
      header: t("columns.name"),
      sortable: true,
      sortAccessor: (row) => row.name,
      cell: (connector) => (
        <div>
          <Link
            href={`/dashboard/connectors/${connector.id}`}
            className="font-medium text-(--ink) no-underline hover:text-primary"
          >
            {connector.name}
          </Link>
          <div className="text-xs text-(--muted-clr)">
            {connector.code}
          </div>
        </div>
      ),
    },
    {
      id: "vendor",
      header: t("columns.vendor"),
      sortable: true,
      sortAccessor: (row) => row.vendor,
      cell: (row) => <span className="text-(--ink-2)">{row.vendor}</span>,
    },
    {
      id: "category",
      header: t("columns.category"),
      sortable: true,
      sortAccessor: (row) => row.category,
      cell: (row) => <Badge variant="gray">{row.category}</Badge>,
    },
    {
      id: "version",
      header: t("columns.version"),
      sortable: true,
      sortAccessor: (row) => row.version,
      cell: (row) => (
        <span className="text-xs text-(--ink-2)">{row.version}</span>
      ),
    },
    {
      id: "capabilities",
      header: t("columns.capabilities"),
      sortable: true,
      sortAccessor: (row) => row.capabilities?.length ?? 0,
      cell: (row) => (
        <span className="text-(--ink-2)">{row.capabilities?.length ?? 0}</span>
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
