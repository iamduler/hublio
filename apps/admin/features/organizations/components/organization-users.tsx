"use client";

import { useTranslations, useLocale } from "next-intl";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { Badge } from "@hublio/ui/ui/badge";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useOrganizationUsers } from "../hooks";
import type { OrganizationUser } from "../types";

export function OrganizationUsers({
  organizationId,
}: {
  organizationId: string;
}) {
  const t = useTranslations("organizations");
  const tCommon = useTranslations("table");
  const locale = useLocale();
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } =
    useOrganizationUsers(organizationId);

  const columns: DataTableColumn<OrganizationUser>[] = [
    {
      id: "name",
      header: t("users.columns.name"),
      sortable: true,
      sortAccessor: (row) => row.full_name,
      cell: (user) => (
        <div className="min-w-0">
          <p className="truncate text-[13px] font-semibold text-(--ink)">
            {user.full_name}
          </p>
          <p className="mt-0.5 truncate text-[11px] text-(--faint)">{user.id}</p>
        </div>
      ),
    },
    {
      id: "email",
      header: t("users.columns.email"),
      sortable: true,
      sortAccessor: (row) => row.email,
      cell: (user) => (
        <span className="text-[13px] text-(--ink)">{user.email}</span>
      ),
    },
    {
      id: "status",
      header: t("users.columns.status"),
      sortable: true,
      sortAccessor: (row) => row.status,
      cell: (user) => <StatusBadge status={user.status} />,
    },
    {
      id: "role",
      header: t("users.columns.platformAdmin"),
      headerClassName: "hidden sm:table-cell",
      className: "hidden sm:table-cell",
      cell: (user) =>
        user.is_platform_admin ? (
          <Badge variant="violet">{t("users.platformAdminYes")}</Badge>
        ) : (
          <span className="text-[12px] text-(--faint)">
            {t("users.platformAdminNo")}
          </span>
        ),
    },
    {
      id: "created",
      header: t("users.columns.created"),
      sortable: true,
      sortAccessor: (row) => row.created_at ?? "",
      headerClassName: "hidden md:table-cell",
      className: "hidden text-(--muted-clr) md:table-cell",
      cell: (user) =>
        user.created_at ? (
          <FormattedDate
            value={user.created_at}
            mode="datetime"
            locale={locale}
          />
        ) : (
          "—"
        ),
    },
    {
      id: "lastLogin",
      header: t("users.columns.lastLogin"),
      sortable: true,
      sortAccessor: (row) => row.last_login_at ?? "",
      headerClassName: "hidden lg:table-cell",
      className: "hidden text-(--muted-clr) lg:table-cell",
      cell: (user) =>
        user.last_login_at ? (
          <FormattedDate
            value={user.last_login_at}
            mode="datetime"
            locale={locale}
          />
        ) : (
          "—"
        ),
    },
  ];

  if (isError) {
    return (
      <ErrorState
        title={t("users.loadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">{t("users.description")}</p>
      <DataTable
        columns={columns}
        rows={data ?? []}
        getRowId={(row) => row.id}
        loading={isLoading}
        mode="client"
        pageSize={10}
        emptyTitle={t("users.empty")}
        emptyDescription={t("users.emptyDescription")}
        prevLabel={tCommon("prev")}
        nextLabel={tCommon("next")}
        paginationShowingLabel={(from, to, total) =>
          tCommon("showing", { from, to, total })
        }
      />
    </div>
  );
}
