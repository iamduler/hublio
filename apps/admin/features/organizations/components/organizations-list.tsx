"use client";

import { useMemo, useState } from "react";
import { useTranslations, useLocale } from "next-intl";
import {
  Download,
  ExternalLink,
  Plus,
  UserCheck,
  UserX,
  X,
} from "lucide-react";
import { Link, useRouter } from "@/i18n/navigation";
import { Button } from "@hublio/ui/ui/button";
import { Breadcrumb } from "@hublio/ui/common/breadcrumb";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { ConfirmDialog } from "@hublio/ui/common/confirm-dialog";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { FilterSelect } from "@hublio/ui/common/filter-select";
import { OrgAvatar } from "@hublio/ui/common/org-avatar";
import { ActionMenu } from "@hublio/ui/common/action-menu";
import { SearchField } from "@hublio/ui/ui/search-field";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  useActivateOrganization,
  useOrganizations,
  useSuspendOrganization,
} from "../hooks";
import type { Organization } from "../types";

const PAGE_SIZE = 10;

type PendingAction =
  | { type: "suspend"; org: Organization }
  | { type: "activate"; org: Organization }
  | null;

export function OrganizationsList() {
  const t = useTranslations("organizations");
  const tCommon = useTranslations("table");
  const locale = useLocale();
  const getError = useApiErrorMessage();
  const router = useRouter();
  const { data, isLoading, isError, error, refetch } = useOrganizations();
  const suspend = useSuspendOrganization();
  const activate = useActivateOrganization();

  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);
  const [pending, setPending] = useState<PendingAction>(null);

  const filtered = useMemo(() => {
    const list = data ?? [];
    const q = search.trim().toLowerCase();
    return list.filter((org) => {
      const matchSearch =
        !q ||
        org.name.toLowerCase().includes(q) ||
        org.id.toLowerCase().includes(q);
      const matchStatus =
        !statusFilter ||
        org.status.toLowerCase() === statusFilter.toLowerCase();
      return matchSearch && matchStatus;
    });
  }, [data, search, statusFilter]);

  async function confirmPending() {
    if (!pending) return;
    try {
      if (pending.type === "suspend") {
        await suspend.mutateAsync(pending.org.id);
        toast.success(t("actions.suspendDone"));
      } else {
        await activate.mutateAsync(pending.org.id);
        toast.success(t("actions.activateDone"));
      }
      setPending(null);
    } catch (err) {
      toast.error(getError(err));
    }
  }

  const columns: DataTableColumn<Organization>[] = [
    {
      id: "name",
      header: t("columns.organization"),
      sortable: true,
      sortAccessor: (row) => row.name,
      cell: (org) => (
        <div className="flex items-center gap-2.5">
          <OrgAvatar name={org.name} size="sm" />
          <div className="min-w-0">
            <Link
              href={`/organizations/${org.id}`}
              className="block truncate text-[13px] font-semibold leading-snug text-(--ink) no-underline hover:text-primary"
              onClick={(e) => e.stopPropagation()}
            >
              {org.name}
            </Link>
            <p className="mt-0.5 truncate text-[11px] text-(--faint)">
              {org.id}
            </p>
          </div>
        </div>
      ),
    },
    {
      id: "status",
      header: t("columns.status"),
      sortable: true,
      sortAccessor: (row) => row.status,
      cell: (org) => <StatusBadge status={org.status} />,
    },
    {
      id: "created",
      header: t("columns.created"),
      sortable: true,
      sortAccessor: (row) => row.created_at ?? "",
      headerClassName: "hidden sm:table-cell",
      className: "hidden text-(--muted-clr) sm:table-cell",
      cell: (org) =>
        org.created_at ? (
          <FormattedDate value={org.created_at} mode="datetime" locale={locale} />
        ) : (
          "—"
        ),
    },
    {
      id: "updated",
      header: t("columns.updated"),
      sortable: true,
      sortAccessor: (row) => row.updated_at ?? "",
      headerClassName: "hidden md:table-cell",
      className: "hidden text-(--muted-clr) md:table-cell",
      cell: (org) =>
        org.updated_at ? (
          <FormattedDate value={org.updated_at} mode="datetime" locale={locale} />
        ) : (
          "—"
        ),
    },
    {
      id: "actions",
      header: t("columns.actions"),
      className: "text-right",
      headerClassName: "text-right",
      cell: (org) => {
        const menuItems = [
          {
            label: t("actions.viewDetails"),
            icon: <ExternalLink size={13} />,
            onSelect: () => router.push(`/organizations/${org.id}`),
          },
          ...(org.status === "active"
            ? [
              {
                label: t("actions.suspend"),
                icon: <UserX size={13} />,
                separator: true as const,
                onSelect: () => setPending({ type: "suspend", org }),
              },
            ]
            : []),
          ...(org.status === "suspended"
            ? [
              {
                label: t("actions.activate"),
                icon: <UserCheck size={13} />,
                separator: true as const,
                onSelect: () => setPending({ type: "activate", org }),
              },
            ]
            : []),
        ];

        return (
          <div
            className="flex items-center justify-end gap-1"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              type="button"
              className="h-7 rounded-lg px-2.5 text-[12px] font-medium text-primary opacity-0 transition-opacity hover:bg-primary/5 group-hover:opacity-100 focus-visible:opacity-100"
              onClick={() => router.push(`/organizations/${org.id}`)}
            >
              {t("actions.view")}
            </button>
            <ActionMenu items={menuItems} label={t("actions.menu")} />
          </div>
        );
      },
    },
  ];

  const hasFilters = Boolean(search || statusFilter);

  const header = (
    <div className="mb-1 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <Breadcrumb
          items={[
            { label: t("breadcrumb.admin"), href: "/" },
            { label: t("breadcrumb.organizations") },
          ]}
          renderLink={({ href, className, children }) => (
            <Link href={href} className={className}>
              {children}
            </Link>
          )}
        />
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="font-display text-xl font-semibold tracking-tight text-(--ink)">
            {t("title")}
          </h1>
          {data ? (
            <span className="rounded-md bg-(--line-2) px-2 py-0.5 text-xs font-semibold text-(--ink-2)">
              {t("totalCount", { count: data.length })}
            </span>
          ) : null}
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-2.5">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => toast.info(t("actions.unavailable"))}
        >
          <Download size={13} />
          {t("actions.export")}
        </Button>
        <Button
          type="button"
          size="sm"
          onClick={() => toast.info(t("actions.unavailable"))}
        >
          <Plus size={13} />
          {t("actions.create")}
        </Button>
      </div>
    </div>
  );

  if (isError) {
    return (
      <div className="space-y-6">
        {header}
        <ErrorState
          title={t("loadError")}
          description={getError(error)}
          onRetry={() => void refetch()}
        />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {header}

      <DataTable
        columns={columns}
        rows={filtered}
        getRowId={(row) => row.id}
        loading={isLoading}
        mode="client"
        page={page}
        pageSize={PAGE_SIZE}
        onPageChange={setPage}
        onRowClick={(row) => router.push(`/organizations/${row.id}`)}
        emptyTitle={t("empty")}
        emptyDescription={
          hasFilters ? t("emptyFiltered") : t("emptyDescription")
        }
        prevLabel={tCommon("prev")}
        nextLabel={tCommon("next")}
        paginationShowingLabel={(from, to, total) =>
          tCommon("showing", { from, to, total })
        }
        toolbar={
          <>
            <SearchField
              size="sm"
              className="min-w-[220px] flex-1 sm:max-w-sm"
              placeholder={t("searchPlaceholder")}
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
            />
            <FilterSelect
              placeholder={t("filters.status")}
              value={statusFilter}
              onChange={(v) => {
                setStatusFilter(v);
                setPage(1);
              }}
              options={[
                { value: "active", label: t("filters.active") },
                { value: "suspended", label: t("filters.suspended") },
              ]}
            />
            {hasFilters ? (
              <button
                type="button"
                className="inline-flex h-8 items-center gap-1.5 rounded-lg px-3 text-xs font-medium text-primary hover:bg-primary/5"
                onClick={() => {
                  setSearch("");
                  setStatusFilter("");
                  setPage(1);
                }}
              >
                <X size={11} />
                {t("filters.clear")}
              </button>
            ) : null}
          </>
        }
      />

      {pending ? (
        <ConfirmDialog
          open
          onOpenChange={(next) => {
            if (!next) setPending(null);
          }}
          title={
            pending.type === "suspend"
              ? t("suspendTitle")
              : t("activateTitle")
          }
          description={
            pending.type === "suspend"
              ? t("suspendBody", { name: pending.org.name })
              : t("activateBody", { name: pending.org.name })
          }
          confirmLabel={
            pending.type === "suspend"
              ? t("actions.suspend")
              : t("actions.activate")
          }
          cancelLabel={t("actions.cancel")}
          destructive={pending.type === "suspend"}
          pending={suspend.isPending || activate.isPending}
          onConfirm={() => confirmPending()}
        />
      ) : null}
    </div>
  );
}
