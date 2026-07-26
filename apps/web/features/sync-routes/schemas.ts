import { z } from "zod";

type TFunc = (key: string) => string;

export function makeSyncRouteSchema(t: TFunc) {
  return z.object({
    name: z.string().min(1, t("required")),
    source_connection_id: z.string().min(1, t("required")),
    destination_connection_id: z.string().min(1, t("required")),
    capability: z.string().min(1, t("required")),
    trigger_type: z.enum(["webhook", "schedule", "both"]),
    resource_types: z.string().optional(),
  });
}

export type SyncRouteFormValues = z.infer<
  ReturnType<typeof makeSyncRouteSchema>
>;
