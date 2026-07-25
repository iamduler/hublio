"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-keys";
import { intentsApi } from "./api";
import type { CreateIntentPayload } from "./types";

export function useIntent(intentId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.intent(intentId ?? "none"),
    queryFn: () => intentsApi.get(intentId!),
    enabled: Boolean(intentId),
  });
}

export function useCreateIntent() {
  return useMutation({
    mutationFn: (payload: CreateIntentPayload) =>
      intentsApi.create(payload, crypto.randomUUID()),
  });
}
