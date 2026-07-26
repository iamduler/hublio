import { z } from "zod";

type TFunc = (key: string) => string;

export function makeCreateApiKeySchema(t: TFunc) {
  return z.object({
    name: z.string().min(1, t("required")),
  });
}

export type CreateApiKeyValues = z.infer<
  ReturnType<typeof makeCreateApiKeySchema>
>;
