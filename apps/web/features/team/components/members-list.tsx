"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { X } from "lucide-react";
import { RoleBadge } from "@hublio/ui/common/role-badge";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { FilterSelect } from "@hublio/ui/common/filter-select";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useMembers } from "../hooks";
import type { WorkspaceMember } from "../types";

const PAGE_SIZE = 10;

export function MembersList() {
  const t = useTranslations("team");
  const tTable = useTranslations("common.table");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useMembers();
  const [roleFilter, setRoleFilter] = useState("");
  const [page, setPage] = useState(1);

  const filtered = useMemo(() => {
    const list = data ?? [];
    if (!roleFilter) return list;
    return list.filter(
      (m) => m.role.toLowerCase() === roleFilter.toLowerCase(),
    );
  }, [data, roleFilter]);

  const roleOptions = useMemo(() => {
    const set = new Set((data ?? []).map((m) => m.role));
    return Array.from(set).sort().map((s) => ({ value: s, label: s }));
  }, [data]);

  const columns: DataTableColumn<WorkspaceMember>[] = [
    {
      id: "name",
      header: t("columns.name"),
      sortable: true,
      sortAccessor: (row) => row.full_name ?? "",
      cell: (member) => (
        <span className="font-medium text-(--ink)">
          {member.full_name || "—"}
        </span>
      ),
    },
    {
      id: "email",
      header: t("columns.email"),
      sortable: true,
      sortAccessor: (row) => row.email,
      cell: (row) => <span className="text-(--ink-2)">{row.email}</span>,
    },
    {
      id: "role",
      header: t("columns.role"),
      sortable: true,
      sortAccessor: (row) => row.role,
      cell: (row) => <RoleBadge role={row.role} />,
    },
    {
      id: "joined",
      header: t("columns.joined"),
      sortable: true,
      sortAccessor: (row) => row.created_at ?? "",
      cell: (row) =>
        row.created_at ? <FormattedDate value={row.created_at} /> : "—",
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
      getRowId={(row) => row.user_id}
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
            placeholder={tTable("filterRole")}
            value={roleFilter}
            onChange={(v) => {
              setRoleFilter(v);
              setPage(1);
            }}
            options={roleOptions}
          />
          {roleFilter ? (
            <button
              type="button"
              className="inline-flex h-8 items-center gap-1.5 rounded-lg px-3 text-xs font-medium text-primary hover:bg-primary/5"
              onClick={() => {
                setRoleFilter("");
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
