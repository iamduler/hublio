"use client";

import { useCallback } from "react";
import { useTranslations } from "next-intl";
import { getApiErrorMessage } from "@/lib/api/errors";

export function useApiErrorMessage() {
  const t = useTranslations("errors");
  return useCallback((err: unknown) => getApiErrorMessage(err, t), [t]);
}
