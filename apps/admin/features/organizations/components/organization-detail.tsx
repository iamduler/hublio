"use client";

import { useMemo, useState } from "react";
import { useTranslations, useLocale } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Archive, Ban, CheckCircle2, Pencil } from "lucide-react";
import { Link, useRouter } from "@/i18n/navigation";
import { Card, CardContent } from "@hublio/ui/ui/card";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@hublio/ui/ui/form";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@hublio/ui/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@hublio/ui/ui/tabs";
import { PageHeader } from "@hublio/ui/common/page-header";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { ConfirmDialog } from "@hublio/ui/common/confirm-dialog";
import { FormattedDate } from "@hublio/ui/common/formatted-date";
import { CopyValue } from "@hublio/ui/common/copy-value";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  useActivateOrganization,
  useArchiveOrganization,
  useOrganization,
  useSuspendOrganization,
  useUpdateOrganization,
} from "../hooks";
import {
  makeRenameOrganizationSchema,
  type RenameOrganizationValues,
} from "../schemas";
import { OrganizationWorkspaces } from "./organization-workspaces";
import { OrganizationUsers } from "./organization-users";

export function OrganizationDetail({
  organizationId,
}: {
  organizationId: string;
}) {
  const t = useTranslations("organizations");
  const tv = useTranslations("validation");
  const locale = useLocale();
  const router = useRouter();
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } =
    useOrganization(organizationId);
  const suspend = useSuspendOrganization();
  const activate = useActivateOrganization();
  const archive = useArchiveOrganization();
  const update = useUpdateOrganization();
  const [renameOpen, setRenameOpen] = useState(false);

  const schema = useMemo(() => makeRenameOrganizationSchema(tv), [tv]);
  const form = useForm<RenameOrganizationValues>({
    resolver: zodResolver(schema),
    values: { name: data?.name ?? "" },
  });

  if (isLoading) return <LoadingState rows={4} />;
  if (isError || !data) {
    return (
      <ErrorState
        title={t("loadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }

  const canSuspend = data.status === "active";
  const canActivate = data.status === "suspended";
  const canArchive =
    data.status === "active" || data.status === "suspended";
  const pending =
    suspend.isPending ||
    activate.isPending ||
    archive.isPending ||
    update.isPending;

  async function onSuspend() {
    try {
      await suspend.mutateAsync(organizationId);
      toast.success(t("actions.suspendDone"));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  async function onActivate() {
    try {
      await activate.mutateAsync(organizationId);
      toast.success(t("actions.activateDone"));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  async function onArchive() {
    try {
      await archive.mutateAsync(organizationId);
      toast.success(t("actions.archiveDone"));
      router.replace("/organizations");
    } catch (err) {
      toast.error(getError(err));
    }
  }

  async function onRename(values: RenameOrganizationValues) {
    try {
      await update.mutateAsync({
        organizationId,
        name: values.name,
      });
      toast.success(t("actions.renameDone"));
      setRenameOpen(false);
    } catch (err) {
      toast.error(getError(err));
    }
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={
          <span className="flex flex-wrap items-center gap-3">
            {data.name}
            <StatusBadge status={data.status} />
          </span>
        }
        description={
          <Link
            href="/organizations"
            className="text-sm text-muted-foreground no-underline hover:underline"
          >
            {t("backToList")}
          </Link>
        }
        actions={
          <div className="flex flex-wrap items-center gap-2">
            {canArchive ? (
              <Button
                variant="outline"
                size="sm"
                disabled={pending}
                onClick={() => setRenameOpen(true)}
              >
                <Pencil size={14} />
                {t("actions.rename")}
              </Button>
            ) : null}
            {canSuspend ? (
              <ConfirmDialog
                trigger={
                  <Button variant="danger-soft" size="sm" disabled={pending}>
                    <Ban size={14} />
                    {t("actions.suspend")}
                  </Button>
                }
                title={t("suspendTitle")}
                description={t("suspendBody", { name: data.name })}
                confirmLabel={t("actions.suspend")}
                cancelLabel={t("actions.cancel")}
                destructive
                pending={suspend.isPending}
                onConfirm={() => onSuspend()}
              />
            ) : null}
            {canActivate ? (
              <ConfirmDialog
                trigger={
                  <Button size="sm" disabled={pending}>
                    <CheckCircle2 size={14} />
                    {t("actions.activate")}
                  </Button>
                }
                title={t("activateTitle")}
                description={t("activateBody", { name: data.name })}
                confirmLabel={t("actions.activate")}
                cancelLabel={t("actions.cancel")}
                pending={activate.isPending}
                onConfirm={() => onActivate()}
              />
            ) : null}
            {canArchive ? (
              <ConfirmDialog
                trigger={
                  <Button variant="danger-soft" size="sm" disabled={pending}>
                    <Archive size={14} />
                    {t("actions.archive")}
                  </Button>
                }
                title={t("archiveTitle")}
                description={t("archiveBody", { name: data.name })}
                confirmLabel={t("actions.archive")}
                cancelLabel={t("actions.cancel")}
                destructive
                pending={archive.isPending}
                onConfirm={() => onArchive()}
              />
            ) : null}
          </div>
        }
      />

      <Tabs defaultValue="overview" className="space-y-4">
        <TabsList>
          <TabsTrigger value="overview">{t("tabs.overview")}</TabsTrigger>
          <TabsTrigger value="workspaces">{t("tabs.workspaces")}</TabsTrigger>
          <TabsTrigger value="users">{t("tabs.users")}</TabsTrigger>
        </TabsList>
        <TabsContent value="overview">
          <Card>
            <CardContent className="grid gap-4 py-6 sm:grid-cols-2">
              <Detail
                label={t("columns.id")}
                value={<CopyValue value={data.id} />}
              />
              <Detail
                label={t("columns.status")}
                value={<StatusBadge status={data.status} />}
              />
              <Detail
                label={t("columns.created")}
                value={
                  data.created_at ? (
                    <FormattedDate
                      value={data.created_at}
                      mode="datetime"
                      locale={locale}
                    />
                  ) : (
                    "—"
                  )
                }
              />
              <Detail
                label={t("columns.updated")}
                value={
                  data.updated_at ? (
                    <FormattedDate
                      value={data.updated_at}
                      mode="datetime"
                      locale={locale}
                    />
                  ) : (
                    "—"
                  )
                }
              />
            </CardContent>
          </Card>
        </TabsContent>
        <TabsContent value="workspaces">
          <OrganizationWorkspaces organizationId={organizationId} />
        </TabsContent>
        <TabsContent value="users">
          <OrganizationUsers organizationId={organizationId} />
        </TabsContent>
      </Tabs>

      <Dialog open={renameOpen} onOpenChange={setRenameOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("renameTitle")}</DialogTitle>
          </DialogHeader>
          <div className="px-5 pb-5 pt-2">
            <Form {...form}>
              <form
                className="space-y-4"
                onSubmit={form.handleSubmit((v) => void onRename(v))}
              >
                <FormField
                  control={form.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>{t("columns.name")}</FormLabel>
                      <FormControl>
                        <Input autoFocus {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <div className="flex justify-end gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => setRenameOpen(false)}
                  >
                    {t("actions.cancel")}
                  </Button>
                  <Button type="submit" disabled={update.isPending}>
                    {t("actions.save")}
                  </Button>
                </div>
              </form>
            </Form>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function Detail({
  label,
  value,
}: {
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="space-y-1">
      <div className="text-xs font-medium uppercase tracking-wide text-(--faint)">
        {label}
      </div>
      <div className="text-sm text-(--ink)">{value}</div>
    </div>
  );
}
