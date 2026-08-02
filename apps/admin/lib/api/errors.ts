import { ApiError } from "./client";

export type ErrorTranslator = (key: string) => string;

export function getApiErrorMessage(
  err: unknown,
  t: ErrorTranslator,
): string {
  if (err instanceof ApiError) {
    if (err.status === 401 || err.code === "UNAUTHORIZED") {
      return t("unauthorized");
    }
    if (err.status >= 500 || err.code === "INTERNAL") {
      return t("generic");
    }
    if (err.message && err.message !== "UNKNOWN") {
      return err.message;
    }
    return t("generic");
  }

  if (err instanceof TypeError) {
    return t("network");
  }

  return t("generic");
}
