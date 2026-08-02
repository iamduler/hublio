"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Plus, X } from "lucide-react";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Label } from "@hublio/ui/ui/label";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@hublio/ui/ui/dialog";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { CopyValue } from "@hublio/ui/common/copy-value";
import { ConfirmDialog } from "@hublio/ui/common/confirm-dialog";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { FilterSelect } from "@hublio/ui/common/filter-select";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  useApiKeys,
  useCreateApiKey,
  useDisableApiKey,
  useRotateApiKey,
} from "../hooks";
import type { ApiKey } from "../types";

const PAGE_SIZE = 10;

export function ApiKeysPanel() {
  const t = useTranslations("apiKeys");
  const tTable = useTranslations("common.table");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useApiKeys();
  const createKey = useCreateApiKey();
  const disableKey = useDisableApiKey();
  const rotateKey = useRotateApiKey();

  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [plaintext, setPlaintext] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);

  const filtered = useMemo(() => {
    const list = data ?? [];
    if (!statusFilter) return list;
    return list.filter(
      (k) => k.status.toLowerCase() === statusFilter.toLowerCase(),
    );
  }, [data, statusFilter]);

  const statusOptions = useMemo(() => {
    const set = new Set((data ?? []).map((k) => k.status));
    return Array.from(set).sort().map((s) => ({ value: s, label: s }));
  }, [data]);

  async function onCreate() {
    try {
      const result = await createKey.mutateAsync(name.trim());
      setPlaintext(result.plaintext);
      setName("");
      toast.success(t("created"));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  async function onRotate(id: string) {
    try {
      const result = await rotateKey.mutateAsync(id);
      setPlaintext(result.plaintext);
      setOpen(true);
      toast.success(t("rotated"));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  async function onDisable(id: string) {
    try {
      await disableKey.mutateAsync(id);
      toast.success(t("disabled"));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  const columns: DataTableColumn<ApiKey>[] = [
    {
      id: "name",
      header: t("columns.name"),
      sortable: true,
      sortAccessor: (row) => row.name,
      cell: (key) => (
        <span className="font-medium text-(--ink)">{key.name}</span>
      ),
    },
    {
      id: "prefix",
      header: t("columns.prefix"),
      sortable: true,
      sortAccessor: (row) => row.prefix ?? "",
      cell: (key) => (
        <span className="text-xs text-(--ink-2)">
          {key.prefix ?? "—"}
        </span>
      ),
    },
    {
      id: "lastUsed",
      header: t("columns.lastUsed"),
      sortable: true,
      sortAccessor: (row) => row.last_used_at ?? "",
      cell: (key) => <FormattedDate value={key.last_used_at} relative />,
    },
    {
      id: "status",
      header: t("columns.status"),
      sortable: true,
      sortAccessor: (row) => row.status,
      cell: (key) => <StatusBadge status={key.status} />,
    },
    {
      id: "actions",
      header: t("columns.actions"),
      className: "text-right",
      headerClassName: "text-right",
      cell: (key) => (
        <div className="flex justify-end gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={rotateKey.isPending}
            onClick={() => void onRotate(key.id)}
          >
            {t("rotate")}
          </Button>
          {key.status === "active" ? (
            <ConfirmDialog
              trigger={
                <Button variant="danger-soft" size="sm">
                  {t("disable")}
                </Button>
              }
              title={t("disableTitle")}
              description={t("disableBody")}
              confirmLabel={t("disable")}
              cancelLabel={t("cancel")}
              destructive
              pending={disableKey.isPending}
              onConfirm={() => onDisable(key.id)}
            />
          ) : null}
        </div>
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

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <Dialog
          open={open}
          onOpenChange={(next) => {
            setOpen(next);
            if (!next) setPlaintext(null);
          }}
        >
          <DialogTrigger asChild>
            <Button size="sm">
              <Plus size={14} />
              {t("create")}
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>{t("createTitle")}</DialogTitle>
            </DialogHeader>
            <DialogBody className="space-y-4">
              {plaintext ? (
                <div className="space-y-2">
                  <p className="text-sm text-(--ink-2)">{t("copyOnce")}</p>
                  <CopyValue value={plaintext} />
                </div>
              ) : (
                <div className="space-y-2">
                  <Label htmlFor="key-name">{t("nameLabel")}</Label>
                  <Input
                    id="key-name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder={t("namePlaceholder")}
                  />
                </div>
              )}
            </DialogBody>
            <DialogFooter>
              {plaintext ? (
                <Button onClick={() => setOpen(false)}>{t("done")}</Button>
              ) : (
                <Button
                  onClick={() => void onCreate()}
                  disabled={!name.trim() || createKey.isPending}
                >
                  {t("generate")}
                </Button>
              )}
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

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
    </div>
  );
}
