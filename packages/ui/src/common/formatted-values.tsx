"use client";

import {
  formatCurrency,
  formatDate,
  formatDateTime,
  formatTime,
} from "../lib/format";

type CommonProps = {
  value?: string | number | Date | null;
  locale?: string;
  fallback?: string;
  className?: string;
};

export function FormattedTime({
  value,
  locale = "en",
  fallback = "—",
  className,
}: CommonProps) {
  const text = formatTime(value, locale, fallback);
  if (!value || text === fallback) {
    return <span className={className}>{fallback}</span>;
  }
  const date =
    value instanceof Date
      ? value
      : new Date(typeof value === "number" ? value : Date.parse(String(value)));
  return (
    <time
      className={className}
      dateTime={Number.isNaN(date.getTime()) ? undefined : date.toISOString()}
    >
      {text}
    </time>
  );
}

export function FormattedDateTime({
  value,
  locale = "en",
  fallback = "—",
  className,
}: CommonProps) {
  const text = formatDateTime(value, locale, fallback);
  if (!value || text === fallback) {
    return <span className={className}>{fallback}</span>;
  }
  const date =
    value instanceof Date
      ? value
      : new Date(typeof value === "number" ? value : Date.parse(String(value)));
  return (
    <time
      className={className}
      dateTime={Number.isNaN(date.getTime()) ? undefined : date.toISOString()}
      title={text}
    >
      {text}
    </time>
  );
}

export function FormattedCurrency({
  value,
  currency = "USD",
  locale = "en",
  fallback = "—",
  className,
}: {
  value?: number | null;
  currency?: string;
  locale?: string;
  fallback?: string;
  className?: string;
}) {
  return (
    <span className={className}>
      {formatCurrency(value, currency, locale, fallback)}
    </span>
  );
}

/** Convenience re-export of medium date-only via Intl. */
export function FormattedDateOnly({
  value,
  locale = "en",
  fallback = "—",
  className,
}: CommonProps) {
  const text = formatDate(value, locale, fallback);
  if (!value || text === fallback) {
    return <span className={className}>{fallback}</span>;
  }
  const date =
    value instanceof Date
      ? value
      : new Date(typeof value === "number" ? value : Date.parse(String(value)));
  return (
    <time
      className={className}
      dateTime={Number.isNaN(date.getTime()) ? undefined : date.toISOString()}
    >
      {text}
    </time>
  );
}
