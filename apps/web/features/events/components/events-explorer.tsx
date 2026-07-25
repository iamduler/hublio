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
import { useEvents } from "../hooks";
import type { DomainEvent, EventCategory } from "../types";

const CATEGORY_VARIANT: Record<EventCategory, "sky" | "violet" | "gray"> = {
  runtime: "sky",
  business: "violet",
  system: "gray",
};

export function EventsExplorer() {
  const t = useTranslations("events");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useEvents({ limit: 100 });
  const [selected, setSelected] = useState<DomainEvent | null>(null);

  if (isLoading) return <LoadingState rows={6} />;
  if (isError) {
    return (
      <ErrorState
        title={t("loadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }
  if (!data || data.length === 0) {
    return <EmptyState icon={Activity} title={t("empty")} />;
  }

  return (
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
            {data.map((event) => (
              <TableRow
                key={event.id}
                className="cursor-pointer"
                onClick={() => setSelected(event)}
              >
                <TableCell className="font-medium text-[var(--ink)]">
                  {event.event_name}
                </TableCell>
                <TableCell>
                  <Badge variant={CATEGORY_VARIANT[event.category] ?? "gray"}>
                    {event.category}
                  </Badge>
                </TableCell>
                <TableCell className="text-[var(--ink-2)]">
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
                <TableCell className="text-[var(--muted-clr)]">
                  <FormattedDate value={event.created_at} relative />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>

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
              <p className="mb-1 text-xs font-medium uppercase tracking-wide text-[var(--muted-clr)]">
                {t("detail.payload")}
              </p>
              <pre className="max-h-80 overflow-auto rounded-md bg-[var(--line-2)] p-4 font-mono text-xs text-[var(--ink-2)]">
                {JSON.stringify(selected?.payload ?? {}, null, 2)}
              </pre>
            </div>
          </DialogBody>
        </DialogContent>
      </Dialog>
    </>
  );
}

function Meta({ label, value }: { label: string; value?: string }) {
  return (
    <div>
      <p className="text-xs font-medium uppercase tracking-wide text-[var(--muted-clr)]">
        {label}
      </p>
      <p className="truncate font-mono text-xs text-[var(--ink)]">
        {value ?? "—"}
      </p>
    </div>
  );
}
