"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import { Pencil, Play, Power, RefreshCw, Trash2 } from "lucide-react";
import { Link, useRouter } from "@/i18n/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { Button } from "@hublio/ui/ui/button";
import { Badge } from "@hublio/ui/ui/badge";
import { Input } from "@hublio/ui/ui/input";
import { Label } from "@hublio/ui/ui/label";
import { Textarea } from "@hublio/ui/ui/textarea";
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
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@hublio/ui/ui/dialog";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { PageHeader } from "@hublio/ui/common/page-header";
import { CopyValue } from "@hublio/ui/common/copy-value";
import { ConfirmDialog } from "@hublio/ui/common/confirm-dialog";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useActiveWorkspaceId } from "@/providers/workspace-provider";
import { queryKeys } from "@/lib/query-keys";
import { syncRoutesApi } from "../api";
import {
  usePollSyncRoute,
  useSyncRoute,
  useSyncRouteAction,
  useUpsertSyncRouteWatermark,
} from "../hooks";
import type { SyncRouteWatermark } from "../types";

export function SyncRouteDetail({ syncRouteId }: { syncRouteId: string }) {
  const t = useTranslations("syncRoutes");
  const getError = useApiErrorMessage();
  const router = useRouter();
  const workspaceId = useActiveWorkspaceId();
  const { data, isLoading, isError, error, refetch } =
    useSyncRoute(syncRouteId);
  const action = useSyncRouteAction();
  const poll = usePollSyncRoute(syncRouteId);
  const upsertWm = useUpsertSyncRouteWatermark(syncRouteId);
  const [secret, setSecret] = useState<string | null>(null);
  const [pollResource, setPollResource] = useState("");
  const [editingWm, setEditingWm] = useState<SyncRouteWatermark | null>(null);
  const [wmCursorText, setWmCursorText] = useState("{}");
  const [wmResourceType, setWmResourceType] = useState("");
  const [addingWm, setAddingWm] = useState(false);

  const watermarks = useQuery({
    queryKey: [
      ...queryKeys.syncRoute(workspaceId ?? "none", syncRouteId),
      "watermarks",
    ],
    queryFn: () => syncRoutesApi.watermarks(workspaceId!, syncRouteId),
    enabled: Boolean(workspaceId && data),
  });

  const resourceOptions = useMemo(() => {
    const fromRoute = data?.resource_types ?? [];
    const fromWm = (watermarks.data ?? []).map((w) => w.resource_type);
    return Array.from(new Set([...fromRoute, ...fromWm].filter(Boolean)));
  }, [data?.resource_types, watermarks.data]);

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

  async function run(kind: "enable" | "disable" | "remove" | "rotate") {
    try {
      const result = await action.mutateAsync({ id: syncRouteId, action: kind });
      if (kind === "remove") {
        toast.success(t("deleted"));
        router.replace("/dashboard/sync-routes");
        return;
      }
      if (kind === "rotate" && result && "webhook_secret" in result) {
        setSecret(
          (result as { webhook_secret?: string }).webhook_secret ?? null,
        );
      }
      toast.success(t(`actions.${kind}Done`));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  async function runPoll() {
    const resourceType =
      pollResource || resourceOptions[0] || data?.resource_types?.[0];
    if (!resourceType) {
      toast.error(t("pollNeedResource"));
      return;
    }
    try {
      await poll.mutateAsync(resourceType);
      toast.success(t("actions.pollDone"));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  function openEditWatermark(wm: SyncRouteWatermark) {
    setAddingWm(false);
    setEditingWm(wm);
    setWmResourceType(wm.resource_type);
    setWmCursorText(JSON.stringify(wm.cursor ?? {}, null, 2));
  }

  function openAddWatermark() {
    setEditingWm(null);
    setAddingWm(true);
    setWmResourceType(resourceOptions[0] ?? "");
    setWmCursorText("{}");
  }

  async function saveWatermark() {
    let cursor: Record<string, unknown>;
    try {
      cursor = JSON.parse(wmCursorText || "{}") as Record<string, unknown>;
    } catch {
      toast.error(t("watermarkCursorInvalid"));
      return;
    }
    const resourceType = wmResourceType.trim();
    if (!resourceType) {
      toast.error(t("pollNeedResource"));
      return;
    }
    try {
      await upsertWm.mutateAsync({ resourceType, cursor });
      toast.success(t("watermarkSaved"));
      setEditingWm(null);
      setAddingWm(false);
    } catch (err) {
      toast.error(getError(err));
    }
  }

  const isDisabled = data.status === "disabled";
  const canEdit = data.status === "draft" || data.status === "disabled";
  const defaultPollResource = pollResource || resourceOptions[0] || "";

  return (
    <div className="space-y-6">
      <PageHeader
        title={
          <span className="flex items-center gap-3">
            {data.name}
            <StatusBadge status={data.status} />
          </span>
        }
        description={
          <span className="flex items-center gap-2">
            <Badge variant="sky">{data.trigger_type}</Badge>
            <span>{(data.resource_types ?? []).join(", ")}</span>
          </span>
        }
        actions={
          <div className="flex flex-wrap items-center gap-2">
            {canEdit ? (
              <Button variant="outline" size="sm" asChild>
                <Link href={`/dashboard/sync-routes/${syncRouteId}/edit`}>
                  <Pencil size={14} />
                  {t("actions.edit")}
                </Link>
              </Button>
            ) : null}
            <div className="flex items-center gap-1">
              {resourceOptions.length > 0 ? (
                <Select
                  value={defaultPollResource}
                  onValueChange={setPollResource}
                >
                  <SelectTrigger className="h-8 w-[140px]">
                    <SelectValue placeholder={t("form.resourceType")} />
                  </SelectTrigger>
                  <SelectContent>
                    {resourceOptions.map((rt) => (
                      <SelectItem key={rt} value={rt}>
                        {rt}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : null}
              <Button
                variant="outline"
                size="sm"
                disabled={poll.isPending}
                onClick={() => void runPoll()}
              >
                <Play size={14} />
                {t("actions.poll")}
              </Button>
            </div>
            <Button
              variant="outline"
              size="sm"
              disabled={action.isPending}
              onClick={() => void run("rotate")}
            >
              <RefreshCw size={14} />
              {t("actions.rotateSecret")}
            </Button>
            <Button
              variant={isDisabled ? "default" : "danger-soft"}
              size="sm"
              disabled={action.isPending}
              onClick={() => void run(isDisabled ? "enable" : "disable")}
            >
              <Power size={14} />
              {isDisabled ? t("actions.enable") : t("actions.disable")}
            </Button>
            <ConfirmDialog
              trigger={
                <Button variant="danger-soft" size="sm">
                  <Trash2 size={14} />
                  {t("actions.delete")}
                </Button>
              }
              title={t("deleteTitle")}
              description={t("deleteBody")}
              confirmLabel={t("actions.delete")}
              cancelLabel={t("form.cancel")}
              destructive
              pending={action.isPending}
              onConfirm={() => run("remove")}
            />
          </div>
        }
      />

      {secret ? (
        <Card>
          <CardHeader>
            <CardTitle>{t("webhookSecret")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <p className="text-sm text-(--ink-2)">{t("secretOnce")}</p>
            <CopyValue value={secret} />
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <CardTitle>{t("watermarks")}</CardTitle>
          <Button variant="outline" size="sm" onClick={openAddWatermark}>
            {t("watermarkAdd")}
          </Button>
        </CardHeader>
        <CardContent>
          {watermarks.isLoading ? (
            <LoadingState rows={2} />
          ) : !watermarks.data || watermarks.data.length === 0 ? (
            <p className="text-sm text-(--muted-clr)">
              {t("noWatermarks")}
            </p>
          ) : (
            <ul className="space-y-2">
              {watermarks.data.map((wm) => (
                <li
                  key={wm.resource_type}
                  className="flex items-center justify-between gap-3 rounded-md border border-(--line) px-3 py-2 text-sm"
                >
                  <div className="min-w-0 flex-1">
                    <span className="font-medium text-(--ink)">
                      {wm.resource_type}
                    </span>
                    <code className="mt-1 block truncate text-xs text-(--muted-clr)">
                      {JSON.stringify(wm.cursor ?? {})}
                    </code>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => openEditWatermark(wm)}
                  >
                    <Pencil size={14} />
                    {t("watermarkEdit")}
                  </Button>
                </li>
              ))}
            </ul>
          )}
        </CardContent>
      </Card>

      <Dialog
        open={Boolean(editingWm) || addingWm}
        onOpenChange={(open) => {
          if (!open) {
            setEditingWm(null);
            setAddingWm(false);
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {addingWm ? t("watermarkAdd") : t("watermarkEdit")}
            </DialogTitle>
          </DialogHeader>
          <DialogBody className="space-y-4">
            <div className="space-y-2">
              <Label>{t("form.resourceType")}</Label>
              {addingWm ? (
                <Input
                  value={wmResourceType}
                  onChange={(e) => setWmResourceType(e.target.value)}
                  placeholder="invoice"
                />
              ) : (
                <Input value={wmResourceType} disabled />
              )}
            </div>
            <div className="space-y-2">
              <Label>{t("watermarkCursor")}</Label>
              <Textarea
                rows={8}
                className="text-xs"
                value={wmCursorText}
                onChange={(e) => setWmCursorText(e.target.value)}
              />
            </div>
          </DialogBody>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setEditingWm(null);
                setAddingWm(false);
              }}
            >
              {t("form.cancel")}
            </Button>
            <Button
              disabled={upsertWm.isPending}
              onClick={() => void saveWatermark()}
            >
              {t("watermarkSave")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
