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
import { useCreateSyncRoute, useUpdateSyncRoute } from "../hooks";
import type { SyncRoute, SyncRouteTrigger } from "../types";

const TRIGGERS: SyncRouteTrigger[] = ["webhook", "schedule", "both"];

function firstStep(route: SyncRoute) {
  const step = route.activities?.[0]?.steps?.[0];
  return {
    destination: step?.destination_connection_id ?? "",
    capability: step?.capability ?? "",
  };
}

/**
 * Simplified single-activity sync route form (source → one destination
 * step). Supports create and edit (PATCH when draft/disabled).
 */
export function SyncRouteForm({
  mode = "create",
  initial,
}: {
  mode?: "create" | "edit";
  initial?: SyncRoute;
}) {
  const t = useTranslations("syncRoutes");
  const getError = useApiErrorMessage();
  const router = useRouter();
  const { data: connections } = useConnections();
  const createRoute = useCreateSyncRoute();
  const updateRoute = useUpdateSyncRoute(initial?.id ?? "");

  const initialStep = initial ? firstStep(initial) : null;

  const [name, setName] = useState(initial?.name ?? "");
  const [source, setSource] = useState(initial?.source_connection_id ?? "");
  const [destination, setDestination] = useState(
    initialStep?.destination ?? "",
  );
  const [capability, setCapability] = useState(initialStep?.capability ?? "");
  const [trigger, setTrigger] = useState<SyncRouteTrigger>(
    initial?.trigger_type ?? "webhook",
  );
  const [resourceTypes, setResourceTypes] = useState(
    (initial?.resource_types ?? []).join(", "),
  );

  const pending = createRoute.isPending || updateRoute.isPending;
  const valid =
    name.trim() && source && destination && capability.trim();

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!valid) return;

    const resource_types = resourceTypes
      .split(",")
      .map((r) => r.trim())
      .filter(Boolean);
    const activities = [
      {
        group_mode: "sequential" as const,
        steps: [
          {
            destination_connection_id: destination,
            capability: capability.trim(),
          },
        ],
      },
    ];

    try {
      if (mode === "edit" && initial) {
        const route = await updateRoute.mutateAsync({
          name: name.trim(),
          source_connection_id: source,
          trigger_type: trigger,
          resource_types,
          activities,
        });
        toast.success(t("updated"));
        router.replace(`/dashboard/sync-routes/${route.id}`);
        return;
      }

      const route = await createRoute.mutateAsync({
        source_connection_id: source,
        name: name.trim(),
        trigger_type: trigger,
        resource_types,
        activities,
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
            <Button type="submit" disabled={!valid || pending}>
              {mode === "edit" ? t("form.submitEdit") : t("form.submit")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
