import { z } from "zod";

type TFunc = (key: string) => string;

export function makeRunIntentSchema(t: TFunc) {
  return z.object({
    connection_id: z.string().min(1, t("required")),
    capability: z.string().min(1, t("required")),
    payload: z.string().min(1, t("required")),
  });
}

export type RunIntentFormValues = z.infer<
  ReturnType<typeof makeRunIntentSchema>
>;
