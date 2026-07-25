"use client";

import { format, formatDistanceToNow, parseISO } from "date-fns";

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

export function FormattedDate({
  value,
  relative,
  pattern = "MMM d, yyyy HH:mm",
  fallback = "—",
}: {
  value?: string | number | Date | null;
  relative?: boolean;
  pattern?: string;
  fallback?: string;
}) {
  if (!value) return <span>{fallback}</span>;
  const date = toDate(value);
  if (!date) return <span>{fallback}</span>;

  const title = format(date, "PPpp");
  const text = relative
    ? formatDistanceToNow(date, { addSuffix: true })
    : format(date, pattern);

  return (
    <time dateTime={date.toISOString()} title={title}>
      {text}
    </time>
  );
}
