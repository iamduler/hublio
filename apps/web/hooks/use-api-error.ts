"use client";

import { useCallback } from "react";
import { useTranslations } from "next-intl";
import { getApiErrorMessage } from "@/lib/api/errors";

/**
 * Returns a memoized translator that maps any thrown value to a
 * localized, user-safe message using the `errors` namespace.
 */
export function useApiErrorMessage() {
  const t = useTranslations("errors");
  return useCallback((err: unknown) => getApiErrorMessage(err, t), [t]);
}
