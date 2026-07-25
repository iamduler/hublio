"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { useRouter } from "@/i18n/navigation";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Label } from "@hublio/ui/ui/label";
import { Card, CardContent } from "@hublio/ui/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@hublio/ui/ui/select";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useConnections } from "@/features/connections/hooks";
import { useCreateSyncRoute } from "../hooks";
import type { SyncRouteTrigger } from "../types";

const TRIGGERS: SyncRouteTrigger[] = ["webhook", "schedule", "both"];

/**
 * Simplified single-activity sync route creator (source → one destination
 * step). Advanced multi-step fan-out can be edited via config later.
 */
export function SyncRouteForm() {
  const t = useTranslations("syncRoutes");
  const getError = useApiErrorMessage();
  const router = useRouter();
  const { data: connections } = useConnections();
  const createRoute = useCreateSyncRoute();

  const [name, setName] = useState("");
  const [source, setSource] = useState("");
  const [destination, setDestination] = useState("");
  const [capability, setCapability] = useState("");
  const [trigger, setTrigger] = useState<SyncRouteTrigger>("webhook");
  const [resourceTypes, setResourceTypes] = useState("");

  const valid =
    name.trim() && source && destination && capability.trim();

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!valid) return;
    try {
      const route = await createRoute.mutateAsync({
        source_connection_id: source,
        name: name.trim(),
        trigger_type: trigger,
        resource_types: resourceTypes
          .split(",")
          .map((r) => r.trim())
          .filter(Boolean),
        activities: [
          {
            group_mode: "sequential",
            steps: [
              {
                destination_connection_id: destination,
                capability: capability.trim(),
              },
            ],
          },
        ],
      });
      toast.success(t("created"));
      router.replace(`/dashboard/sync-routes/${route.id}`);
    } catch (err) {
      toast.error(getError(err));
    }
  }

  const options = connections ?? [];

  return (
    <Card>
      <CardContent className="p-6">
        <form className="space-y-5" onSubmit={(e) => void onSubmit(e)}>
          <div className="space-y-2">
            <Label htmlFor="sr-name">{t("form.name")}</Label>
            <Input
              id="sr-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          <div className="grid gap-5 sm:grid-cols-2">
            <div className="space-y-2">
              <Label>{t("form.source")}</Label>
              <Select value={source} onValueChange={setSource}>
                <SelectTrigger>
                  <SelectValue placeholder={t("form.selectConnection")} />
                </SelectTrigger>
                <SelectContent>
                  {options.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>{t("form.trigger")}</Label>
              <Select
                value={trigger}
                onValueChange={(v) => setTrigger(v as SyncRouteTrigger)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {TRIGGERS.map((tr) => (
                    <SelectItem key={tr} value={tr}>
                      {tr}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="sr-resources">{t("form.resourceTypes")}</Label>
            <Input
              id="sr-resources"
              placeholder={t("form.resourceTypesHint")}
              value={resourceTypes}
              onChange={(e) => setResourceTypes(e.target.value)}
            />
          </div>

          <div className="space-y-4 rounded-md border border-[var(--line)] bg-[var(--line-2)] p-4">
            <p className="text-sm font-medium text-[var(--ink)]">
              {t("form.destinationStep")}
            </p>
            <div className="grid gap-5 sm:grid-cols-2">
              <div className="space-y-2">
                <Label>{t("form.destination")}</Label>
                <Select value={destination} onValueChange={setDestination}>
                  <SelectTrigger>
                    <SelectValue placeholder={t("form.selectConnection")} />
                  </SelectTrigger>
                  <SelectContent>
                    {options.map((c) => (
                      <SelectItem key={c.id} value={c.id}>
                        {c.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="sr-capability">{t("form.capability")}</Label>
                <Input
                  id="sr-capability"
                  value={capability}
                  onChange={(e) => setCapability(e.target.value)}
                />
              </div>
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => router.back()}
            >
              {t("form.cancel")}
            </Button>
            <Button type="submit" disabled={!valid || createRoute.isPending}>
              {t("form.submit")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
