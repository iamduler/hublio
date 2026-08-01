"use client";

import { useTranslations } from "next-intl";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Power } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { Button } from "@hublio/ui/ui/button";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { EnvBadge } from "@hublio/ui/common/env-badge";
import { ConfirmDialog } from "@hublio/ui/common/confirm-dialog";
import { EmptyState } from "@hublio/ui/ui/empty-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useAuth } from "@/providers/auth-provider";
import { useWorkspace } from "@/providers/workspace-provider";
import { queryKeys } from "@/lib/query-keys";
import { workspacesApi } from "../api";

export function WorkspaceSettings() {
  const t = useTranslations("workspaces");
  const getError = useApiErrorMessage();
  const { user } = useAuth();
  const { activeWorkspace } = useWorkspace();
  const queryClient = useQueryClient();

  const toggle = useMutation({
    mutationFn: ({ id, enable }: { id: string; enable: boolean }) =>
      enable ? workspacesApi.enable(id) : workspacesApi.disable(id),
    onSuccess: () => {
      if (user?.organization_id) {
        void queryClient.invalidateQueries({
          queryKey: queryKeys.workspaces(user.organization_id),
        });
      }
    },
  });

  if (!activeWorkspace) {
    return <EmptyState title={t("settings.noWorkspace")} />;
  }

  const isDisabled = activeWorkspace.status === "disabled";

  async function onToggle() {
    try {
      await toggle.mutateAsync({
        id: activeWorkspace!.id,
        enable: isDisabled,
      });
      toast.success(isDisabled ? t("settings.enabled") : t("settings.disabled"));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("settings.general")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <Row label={t("settings.name")}>
          <span className="font-medium text-(--ink)">
            {activeWorkspace.name}
          </span>
        </Row>
        <Row label={t("settings.environment")}>
          <EnvBadge environment={activeWorkspace.environment} />
        </Row>
        <Row label={t("settings.status")}>
          <StatusBadge status={activeWorkspace.status} />
        </Row>
        <div className="border-t border-(--line) pt-4">
          <ConfirmDialog
            trigger={
              <Button variant={isDisabled ? "default" : "danger-soft"} size="sm">
                <Power size={14} />
                {isDisabled
                  ? t("settings.enable")
                  : t("settings.disable")}
              </Button>
            }
            title={
              isDisabled ? t("settings.enableTitle") : t("settings.disableTitle")
            }
            description={
              isDisabled ? t("settings.enableBody") : t("settings.disableBody")
            }
            confirmLabel={
              isDisabled ? t("settings.enable") : t("settings.disable")
            }
            cancelLabel={t("form.cancel")}
            destructive={!isDisabled}
            pending={toggle.isPending}
            onConfirm={onToggle}
          />
        </div>
      </CardContent>
    </Card>
  );
}

function Row({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-(--muted-clr)">{label}</span>
      {children}
    </div>
  );
}
