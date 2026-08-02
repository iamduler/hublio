"use client";

import { useTranslations, useLocale } from "next-intl";
import { Ban, CheckCircle2 } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { Card, CardContent } from "@hublio/ui/ui/card";
import { Button } from "@hublio/ui/ui/button";
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
  useOrganization,
  useSuspendOrganization,
} from "../hooks";

export function OrganizationDetail({
  organizationId,
}: {
  organizationId: string;
}) {
  const t = useTranslations("organizations");
  const locale = useLocale();
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } =
    useOrganization(organizationId);
  const suspend = useSuspendOrganization();
  const activate = useActivateOrganization();

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
  const pending = suspend.isPending || activate.isPending;

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
          </div>
        }
      />

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
