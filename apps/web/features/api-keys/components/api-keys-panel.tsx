"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { KeyRound, Plus } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@hublio/ui/ui/table";
import { Card } from "@hublio/ui/ui/card";
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
import { EmptyState } from "@hublio/ui/ui/empty-state";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  useApiKeys,
  useCreateApiKey,
  useDisableApiKey,
  useRotateApiKey,
} from "../hooks";

export function ApiKeysPanel() {
  const t = useTranslations("apiKeys");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useApiKeys();
  const createKey = useCreateApiKey();
  const disableKey = useDisableApiKey();
  const rotateKey = useRotateApiKey();

  const [open, setOpen] = useState(false);
  const [name, setName] = useState("");
  const [plaintext, setPlaintext] = useState<string | null>(null);

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

      {isLoading ? (
        <LoadingState rows={3} />
      ) : isError ? (
        <ErrorState
          title={t("loadError")}
          description={getError(error)}
          onRetry={() => void refetch()}
        />
      ) : !data || data.length === 0 ? (
        <EmptyState icon={KeyRound} title={t("empty")} />
      ) : (
        <Card className="overflow-hidden">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("columns.name")}</TableHead>
                <TableHead>{t("columns.prefix")}</TableHead>
                <TableHead>{t("columns.lastUsed")}</TableHead>
                <TableHead>{t("columns.status")}</TableHead>
                <TableHead className="text-right">
                  {t("columns.actions")}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.map((key) => (
                <TableRow key={key.id}>
                  <TableCell className="font-medium text-(--ink)">
                    {key.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-(--ink-2)">
                    {key.prefix ?? "—"}
                  </TableCell>
                  <TableCell className="text-(--ink-2)">
                    <FormattedDate value={key.last_used_at} relative />
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={key.status} />
                  </TableCell>
                  <TableCell className="text-right">
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
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Card>
      )}
    </div>
  );
}
