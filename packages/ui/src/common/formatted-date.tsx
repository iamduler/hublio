"use client";

import { format, formatDistanceToNow, parseISO } from "date-fns";
import {
  formatDate as formatDateIntl,
  formatDateTime as formatDateTimeIntl,
  formatTime as formatTimeIntl,
} from "../lib/format";

function toDate(value: string | number | Date): Date | null {
  if (value instanceof Date) return value;
  if (typeof value === "number") return new Date(value);
  try {
    return parseISO(value);
  } catch {
    const d = new Date(value);
    return Number.isNaN(d.getTime()) ? null : d;
  }
}

export type FormattedDateMode = "datetime" | "date" | "time" | "relative";

export function FormattedDate({
  value,
  relative,
  mode,
  pattern = "dd-MM-yyyy HH:mm",
  locale = "en",
  fallback = "—",
}: {
  value?: string | number | Date | null;
  /** @deprecated Prefer `mode="relative"`. */
  relative?: boolean;
  mode?: FormattedDateMode;
  pattern?: string;
  locale?: string;
  fallback?: string;
}) {
  if (!value) return <span>{fallback}</span>;
  const date = toDate(value);
  if (!date) return <span>{fallback}</span>;

  const resolved: FormattedDateMode =
    mode ?? (relative ? "relative" : "datetime");

  let text: string;
  switch (resolved) {
    case "date":
      text = formatDateIntl(date, locale, fallback);
      break;
    case "time":
      text = formatTimeIntl(date, locale, fallback);
      break;
    case "relative":
      text = formatDistanceToNow(date, { addSuffix: true });
      break;
    case "datetime":
    default:
      text =
        pattern === "dd-MM-yyyy HH:mm" || pattern === "MMM d, yyyy HH:mm"
          ? formatDateTimeIntl(date, locale, fallback)
          : format(date, pattern);
      break;
  }

  const title = format(date, "PPpp");

  return (
    <time dateTime={date.toISOString()} title={title}>
      {text}
    </time>
  );
}
