"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Activity } from "lucide-react";
import { Link } from "@/i18n/navigation";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@hublio/ui/ui/table";
import { Card } from "@hublio/ui/ui/card";
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
import { EmptyState } from "@hublio/ui/ui/empty-state";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
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

  return (
    <div className="space-y-4">
      <Card className="p-4">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end">
          <div className="min-w-0 flex-1 space-y-2">
            <Label htmlFor="events-execution">{t("filters.executionId")}</Label>
            <Input
              id="events-execution"
              placeholder={t("filters.executionIdHint")}
              value={executionIdInput}
              onChange={(e) => setExecutionIdInput(e.target.value)}
              className="font-mono text-xs"
            />
          </div>
          <div className="w-full space-y-2 sm:w-44">
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
      </Card>

      {isLoading ? (
        <LoadingState rows={6} />
      ) : isError ? (
        <ErrorState
          title={t("loadError")}
          description={getError(error)}
          onRetry={() => void refetch()}
        />
      ) : events.length === 0 ? (
        <EmptyState icon={Activity} title={t("empty")} />
      ) : (
        <>
          <Card className="overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("columns.event")}</TableHead>
                  <TableHead>{t("columns.category")}</TableHead>
                  <TableHead>{t("columns.aggregate")}</TableHead>
                  <TableHead>{t("columns.execution")}</TableHead>
                  <TableHead>{t("columns.time")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {events.map((event) => (
                  <TableRow
                    key={event.id}
                    className="cursor-pointer"
                    onClick={() => setSelected(event)}
                  >
                    <TableCell className="font-medium text-(--ink)">
                      {event.event_name}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={CATEGORY_VARIANT[event.category] ?? "gray"}
                      >
                        {event.category}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-(--ink-2)">
                      {event.aggregate_type}
                    </TableCell>
                    <TableCell
                      onClick={(e) => e.stopPropagation()}
                      className="font-mono text-xs"
                    >
                      {event.execution_id ? (
                        <Link
                          href={`/dashboard/executions/${event.execution_id}`}
                          className="text-primary no-underline hover:underline"
                        >
                          {event.execution_id.slice(0, 8)}…
                        </Link>
                      ) : (
                        "—"
                      )}
                    </TableCell>
                    <TableCell className="text-(--muted-clr)">
                      <FormattedDate value={event.created_at} relative />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Card>

          {hasNextPage ? (
            <div className="flex justify-center">
              <Button
                variant="outline"
                size="sm"
                disabled={isFetchingNextPage}
                onClick={() => void fetchNextPage()}
              >
                {isFetchingNextPage ? t("loadingMore") : t("loadMore")}
              </Button>
            </div>
          ) : null}

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
                  <Meta
                    label={t("columns.category")}
                    value={selected?.category}
                  />
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
                  <pre className="max-h-80 overflow-auto rounded-md bg-(--line-2) p-4 font-mono text-xs text-(--ink-2)">
                    {JSON.stringify(selected?.payload ?? {}, null, 2)}
                  </pre>
                </div>
              </DialogBody>
            </DialogContent>
          </Dialog>
        </>
      )}
    </div>
  );
}

function Meta({ label, value }: { label: string; value?: string }) {
  return (
    <div>
      <p className="text-xs font-medium uppercase tracking-wide text-(--muted-clr)">
        {label}
      </p>
      <p className="truncate font-mono text-xs text-(--ink)">
        {value ?? "—"}
      </p>
    </div>
  );
}
