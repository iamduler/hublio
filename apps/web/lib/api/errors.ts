import { ApiError } from "./client";

export type ErrorTranslator = (key: string) => string;

/**
 * Convert any thrown value into a user-safe message.
 *
 * Handles both Go error shapes surfaced through {@link ApiError}:
 *  - `{ error, code, detail }`  (domain errors)
 *  - `{ error, message }`       (auth/middleware errors)
 *
 * Never returns raw backend exception text for 5xx; falls back to
 * the localized generic message instead.
 */
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
    // fetch network failure
    return t("network");
  }

  return t("generic");
}
