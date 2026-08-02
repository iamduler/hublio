"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Play } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Label } from "@hublio/ui/ui/label";
import { Textarea } from "@hublio/ui/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@hublio/ui/ui/select";
import { StatusBadge } from "@hublio/ui/common/status-badge";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useConnections } from "@/features/connections/hooks";
import { useCreateIntent } from "../hooks";
import type { CreateIntentResult } from "../types";

export function RunIntent() {
  const t = useTranslations("intents");
  const getError = useApiErrorMessage();
  const { data: connections } = useConnections();
  const createIntent = useCreateIntent();

  const [connectionId, setConnectionId] = useState("");
  const [capability, setCapability] = useState("");
  const [payload, setPayload] = useState("{}");
  const [payloadError, setPayloadError] = useState<string | null>(null);
  const [result, setResult] = useState<CreateIntentResult | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    let parsed: Record<string, unknown>;
    try {
      parsed = JSON.parse(payload || "{}") as Record<string, unknown>;
      setPayloadError(null);
    } catch {
      setPayloadError(t("invalidJson"));
      return;
    }
    try {
      const res = await createIntent.mutateAsync({
        connection_id: connectionId,
        capability: capability.trim(),
        payload: parsed,
      });
      setResult(res);
      toast.success(t("submitted"));
    } catch (err) {
      toast.error(getError(err));
    }
  }

  const executions =
    result?.executions ?? (result?.execution ? [result.execution] : []);

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>{t("runTitle")}</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={(e) => void onSubmit(e)}>
            <div className="space-y-2">
              <Label>{t("form.connection")}</Label>
              <Select value={connectionId} onValueChange={setConnectionId}>
                <SelectTrigger>
                  <SelectValue placeholder={t("form.selectConnection")} />
                </SelectTrigger>
                <SelectContent>
                  {(connections ?? []).map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="capability">{t("form.capability")}</Label>
              <Input
                id="capability"
                value={capability}
                onChange={(e) => setCapability(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="payload">{t("form.payload")}</Label>
              <Textarea
                id="payload"
                rows={8}
                className="text-xs"
                value={payload}
                onChange={(e) => setPayload(e.target.value)}
              />
              {payloadError ? (
                <p className="text-xs font-medium text-border-(--danger)">
                  {payloadError}
                </p>
              ) : null}
            </div>
            <Button
              type="submit"
              disabled={
                !connectionId || !capability.trim() || createIntent.isPending
              }
            >
              <Play size={14} />
              {t("form.submit")}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t("resultTitle")}</CardTitle>
        </CardHeader>
        <CardContent>
          {result ? (
            <div className="space-y-4 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-(--muted-clr)">
                  {t("result.intent")}
                </span>
                <Link
                  href={`/dashboard/intents/${result.intent.id}`}
                  className="text-xs text-primary no-underline hover:underline"
                >
                  {result.intent.id}
                </Link>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-(--muted-clr)">
                  {t("result.status")}
                </span>
                <StatusBadge status={result.intent.status} />
              </div>
              <div className="space-y-2">
                <span className="text-(--muted-clr)">
                  {t("result.executions")}
                </span>
                <ul className="space-y-1.5">
                  {executions.map((exec) => (
                    <li key={exec.id}>
                      <Link
                        href={`/dashboard/executions/${exec.id}`}
                        className="flex items-center justify-between rounded-md border border-(--line) px-3 py-2 no-underline hover:border-primary"
                      >
                        <span className="text-xs text-(--ink-2)">
                          {exec.id}
                        </span>
                        <StatusBadge status={exec.status} />
                      </Link>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          ) : (
            <p className="text-sm text-(--muted-clr)">{t("noResult")}</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
