"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import { Badge } from "@hublio/ui/ui/badge";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Label } from "@hublio/ui/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@hublio/ui/ui/select";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@hublio/ui/ui/dialog";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useInfiniteEvents } from "../hooks";
import type { DomainEvent, EventCategory } from "../types";

const CATEGORY_VARIANT: Record<EventCategory, "sky" | "violet" | "gray"> = {
  runtime: "sky",
  business: "violet",
  system: "gray",
};

const CATEGORIES: Array<EventCategory | "all"> = [
  "all",
  "runtime",
  "business",
  "system",
];

const PAGE_LIMIT = 50;

export function EventsExplorer() {
  const t = useTranslations("events");
  const getError = useApiErrorMessage();

  const [executionIdInput, setExecutionIdInput] = useState("");
  const [categoryInput, setCategoryInput] = useState<EventCategory | "all">(
    "all",
  );
  const [applied, setApplied] = useState<{
    execution_id?: string;
    category?: EventCategory | "all";
    limit: number;
  }>({ limit: PAGE_LIMIT });

  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteEvents({
    execution_id: applied.execution_id,
    category: applied.category,
    limit: applied.limit,
  });
  const [selected, setSelected] = useState<DomainEvent | null>(null);

  const events = data?.pages.flatMap((page) => page.items) ?? [];

  function applyFilters() {
    setApplied({
      execution_id: executionIdInput.trim() || undefined,
      category: categoryInput,
      limit: PAGE_LIMIT,
    });
  }

  function clearFilters() {
    setExecutionIdInput("");
    setCategoryInput("all");
    setApplied({ limit: PAGE_LIMIT });
  }

  const columns: DataTableColumn<DomainEvent>[] = [
    {
      id: "event",
      header: t("columns.event"),
      sortable: true,
      sortAccessor: (row) => row.event_name,
      cell: (event) => (
        <span className="font-medium text-(--ink)">{event.event_name}</span>
      ),
    },
    {
      id: "category",
      header: t("columns.category"),
      sortable: true,
      sortAccessor: (row) => row.category,
      cell: (event) => (
        <Badge variant={CATEGORY_VARIANT[event.category] ?? "gray"}>
          {event.category}
        </Badge>
      ),
    },
    {
      id: "aggregate",
      header: t("columns.aggregate"),
      sortable: true,
      sortAccessor: (row) => row.aggregate_type,
      cell: (event) => (
        <span className="text-(--ink-2)">{event.aggregate_type}</span>
      ),
    },
    {
      id: "execution",
      header: t("columns.execution"),
      sortable: true,
      sortAccessor: (row) => row.execution_id ?? "",
      cell: (event) =>
        event.execution_id ? (
          <Link
            href={`/dashboard/executions/${event.execution_id}`}
            className="text-xs text-primary no-underline hover:underline"
            onClick={(e) => e.stopPropagation()}
          >
            {event.execution_id.slice(0, 8)}…
          </Link>
        ) : (
          "—"
        ),
    },
    {
      id: "time",
      header: t("columns.time"),
      sortable: true,
      sortAccessor: (row) => row.created_at ?? "",
      cell: (event) => <FormattedDate value={event.created_at} relative />,
    },
  ];

  const cursorPagination =
    events.length > 0 ? (
      <div className="flex items-center justify-between gap-3 border-t border-(--line) px-4 py-3">
        <p className="text-xs text-(--muted-clr)">
          {t("loadedCount", { count: events.length })}
        </p>
        {hasNextPage ? (
          <Button
            variant="outline"
            size="sm"
            disabled={isFetchingNextPage}
            onClick={() => void fetchNextPage()}
          >
            {isFetchingNextPage ? t("loadingMore") : t("loadMore")}
          </Button>
        ) : (
          <span className="text-xs text-(--faint)">{t("endOfList")}</span>
        )}
      </div>
    ) : null;

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
    <div className="space-y-4">
      <DataTable
        columns={columns}
        rows={events}
        getRowId={(row) => row.id}
        loading={isLoading}
        mode="server"
        total={events.length}
        hidePagination={!cursorPagination}
        pagination={cursorPagination}
        onRowClick={(row) => setSelected(row)}
        emptyTitle={t("empty")}
        toolbar={
          <div className="flex w-full flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-end">
            <div className="min-w-0 flex-1 space-y-1.5">
              <Label htmlFor="events-execution">{t("filters.executionId")}</Label>
              <Input
                id="events-execution"
                placeholder={t("filters.executionIdHint")}
                value={executionIdInput}
                onChange={(e) => setExecutionIdInput(e.target.value)}
                className="text-xs"
              />
            </div>
            <div className="w-full space-y-1.5 sm:w-44">
              <Label>{t("filters.category")}</Label>
              <Select
                value={categoryInput}
                onValueChange={(v) =>
                  setCategoryInput(v as EventCategory | "all")
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {CATEGORIES.map((cat) => (
                    <SelectItem key={cat} value={cat}>
                      {cat === "all" ? t("filters.allCategories") : cat}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex gap-2">
              <Button size="sm" onClick={applyFilters}>
                {t("filters.apply")}
              </Button>
              <Button size="sm" variant="outline" onClick={clearFilters}>
                {t("filters.clear")}
              </Button>
            </div>
          </div>
        }
      />

      <Dialog
        open={Boolean(selected)}
        onOpenChange={(open) => !open && setSelected(null)}
      >
        <DialogContent wide>
          <DialogHeader>
            <DialogTitle>{selected?.event_name}</DialogTitle>
          </DialogHeader>
          <DialogBody className="space-y-4">
            <div className="grid grid-cols-2 gap-3 text-sm">
              <Meta label={t("columns.category")} value={selected?.category} />
              <Meta
                label={t("columns.aggregate")}
                value={selected?.aggregate_type}
              />
              <Meta
                label={t("detail.correlation")}
                value={selected?.correlation_id ?? "—"}
              />
              <Meta
                label={t("detail.publishedBy")}
                value={selected?.published_by ?? "—"}
              />
            </div>
            <div>
              <p className="mb-1 text-xs font-medium uppercase tracking-wide text-(--muted-clr)">
                {t("detail.payload")}
              </p>
              <pre className="max-h-80 overflow-auto rounded-md bg-(--line-2) p-4 text-xs text-(--ink-2)">
                {JSON.stringify(selected?.payload ?? {}, null, 2)}
              </pre>
            </div>
          </DialogBody>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function Meta({ label, value }: { label: string; value?: string }) {
  return (
    <div>
      <p className="text-xs font-medium uppercase tracking-wide text-(--muted-clr)">
        {label}
      </p>
      <p className="truncate text-xs text-(--ink)">
        {value ?? "—"}
      </p>
    </div>
  );
}
