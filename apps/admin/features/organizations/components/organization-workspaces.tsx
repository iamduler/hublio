"use client";

import { useMemo, useState } from "react";
import { useTranslations, useLocale } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, Power, PowerOff } from "lucide-react";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Card, CardContent } from "@hublio/ui/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@hublio/ui/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@hublio/ui/ui/select";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { EnvBadge } from "@hublio/ui/common/env-badge";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { ConfirmDialog } from "@hublio/ui/common/confirm-dialog";
import { DataTable, type DataTableColumn } from "@hublio/ui/common/data-table";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  useCreateWorkspace,
  useDisableWorkspace,
  useEnableWorkspace,
  useOrganizationWorkspaces,
} from "../hooks";
import {
  makeCreateWorkspaceSchema,
  type CreateWorkspaceValues,
} from "../schemas";
import type { Workspace } from "../types";

const ENVIRONMENTS = ["production", "staging", "development"] as const;

type PendingWs =
  | { type: "enable"; ws: Workspace }
  | { type: "disable"; ws: Workspace }
  | null;

export function OrganizationWorkspaces({
  organizationId,
}: {
  organizationId: string;
}) {
  const t = useTranslations("organizations");
  const tv = useTranslations("validation");
  const tCommon = useTranslations("table");
  const locale = useLocale();
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } =
    useOrganizationWorkspaces(organizationId);
  const create = useCreateWorkspace(organizationId);
  const enable = useEnableWorkspace(organizationId);
  const disable = useDisableWorkspace(organizationId);
  const [pending, setPending] = useState<PendingWs>(null);
  const [showCreate, setShowCreate] = useState(false);

  const schema = useMemo(() => makeCreateWorkspaceSchema(tv), [tv]);
  const form = useForm<CreateWorkspaceValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", environment: "production" },
  });

  async function onCreate(values: CreateWorkspaceValues) {
    try {
      await create.mutateAsync(values);
      toast.success(t("workspaces.created"));
      form.reset({ name: "", environment: "production" });
      setShowCreate(false);
    } catch (err) {
      toast.error(getError(err));
    }
  }

  async function confirmPending() {
    if (!pending) return;
    try {
      if (pending.type === "enable") {
        await enable.mutateAsync(pending.ws.id);
        toast.success(t("workspaces.enableDone"));
      } else {
        await disable.mutateAsync(pending.ws.id);
        toast.success(t("workspaces.disableDone"));
      }
      setPending(null);
    } catch (err) {
      toast.error(getError(err));
    }
  }

  const columns: DataTableColumn<Workspace>[] = [
    {
      id: "name",
      header: t("workspaces.columns.name"),
      sortable: true,
      sortAccessor: (row) => row.name,
      cell: (ws) => (
        <div className="min-w-0">
          <p className="truncate text-[13px] font-semibold text-(--ink)">
            {ws.name}
          </p>
          <p className="mt-0.5 truncate text-[11px] text-(--faint)">{ws.id}</p>
        </div>
      ),
    },
    {
      id: "environment",
      header: t("workspaces.columns.environment"),
      sortable: true,
      sortAccessor: (row) => row.environment,
      cell: (ws) => <EnvBadge environment={ws.environment} />,
    },
    {
      id: "status",
      header: t("workspaces.columns.status"),
      sortable: true,
      sortAccessor: (row) => row.status,
      cell: (ws) => <StatusBadge status={ws.status} />,
    },
    {
      id: "created",
      header: t("workspaces.columns.created"),
      sortable: true,
      sortAccessor: (row) => row.created_at ?? "",
      headerClassName: "hidden sm:table-cell",
      className: "hidden text-(--muted-clr) sm:table-cell",
      cell: (ws) =>
        ws.created_at ? (
          <FormattedDate
            value={ws.created_at}
            mode="datetime"
            locale={locale}
          />
        ) : (
          "—"
        ),
    },
    {
      id: "actions",
      header: t("columns.actions"),
      className: "text-right",
      headerClassName: "text-right",
      cell: (ws) => {
        const canEnable = ws.status === "disabled";
        const canDisable = ws.status === "active";
        return (
          <div
            className="flex items-center justify-end gap-1"
            onClick={(e) => e.stopPropagation()}
          >
            {canEnable ? (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setPending({ type: "enable", ws })}
              >
                <Power size={13} />
                {t("workspaces.enable")}
              </Button>
            ) : null}
            {canDisable ? (
              <Button
                type="button"
                variant="danger-soft"
                size="sm"
                onClick={() => setPending({ type: "disable", ws })}
              >
                <PowerOff size={13} />
                {t("workspaces.disable")}
              </Button>
            ) : null}
          </div>
        );
      },
    },
  ];

  if (isError) {
    return (
      <ErrorState
        title={t("workspaces.loadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm text-muted-foreground">
          {t("workspaces.description")}
        </p>
        <Button
          type="button"
          size="sm"
          onClick={() => setShowCreate((v) => !v)}
        >
          <Plus size={13} />
          {t("workspaces.create")}
        </Button>
      </div>

      {showCreate ? (
        <Card>
          <CardContent className="py-5">
            <Form {...form}>
              <form
                className="grid gap-4 sm:grid-cols-[1fr_180px_auto] sm:items-end"
                onSubmit={form.handleSubmit((v) => void onCreate(v))}
              >
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("workspaces.form.name")}</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="environment"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>
                        {t("workspaces.form.environment")}
                      </FormLabel>
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {ENVIRONMENTS.map((env) => (
                            <SelectItem key={env} value={env}>
                              {t(`workspaces.env.${env}`)}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button type="submit" disabled={create.isPending}>
                  {t("workspaces.submit")}
                </Button>
              </form>
            </Form>
          </CardContent>
        </Card>
      ) : null}

      <DataTable
        columns={columns}
        rows={data ?? []}
        getRowId={(row) => row.id}
        loading={isLoading}
        mode="client"
        pageSize={10}
        emptyTitle={t("workspaces.empty")}
        emptyDescription={t("workspaces.emptyDescription")}
        prevLabel={tCommon("prev")}
        nextLabel={tCommon("next")}
        paginationShowingLabel={(from, to, total) =>
          tCommon("showing", { from, to, total })
        }
      />

      {pending ? (
        <ConfirmDialog
          open
          onOpenChange={(next) => {
            if (!next) setPending(null);
          }}
          title={
            pending.type === "enable"
              ? t("workspaces.enableTitle")
              : t("workspaces.disableTitle")
          }
          description={
            pending.type === "enable"
              ? t("workspaces.enableBody", { name: pending.ws.name })
              : t("workspaces.disableBody", { name: pending.ws.name })
          }
          confirmLabel={
            pending.type === "enable"
              ? t("workspaces.enable")
              : t("workspaces.disable")
          }
          cancelLabel={t("actions.cancel")}
          destructive={pending.type === "disable"}
          pending={enable.isPending || disable.isPending}
          onConfirm={() => confirmPending()}
        />
      ) : null}
    </div>
  );
}
