/** Pure format helpers — safe for Server Components. */

function toDate(value: string | number | Date): Date | null {
  if (value instanceof Date) return Number.isNaN(value.getTime()) ? null : value;
  if (typeof value === "number") {
    const d = new Date(value);
    return Number.isNaN(d.getTime()) ? null : d;
  }
  const parsed = Date.parse(value);
  if (Number.isNaN(parsed)) return null;
  return new Date(parsed);
}

function pad2(n: number): string {
  return String(n).padStart(2, "0");
}

/** DD-MM-YYYY */
export function formatDate(
  value?: string | number | Date | null,
  _locale = "en",
  fallback = "—",
): string {
  const date = value == null ? null : toDate(value);
  if (!date) return fallback;
  return `${pad2(date.getDate())}-${pad2(date.getMonth() + 1)}-${date.getFullYear()}`;
}

/** H:i 24h → HH:mm */
export function formatTime(
  value?: string | number | Date | null,
  _locale = "en",
  fallback = "—",
): string {
  const date = value == null ? null : toDate(value);
  if (!date) return fallback;
  return `${pad2(date.getHours())}:${pad2(date.getMinutes())}`;
}

/** DD-MM-YYYY HH:mm (24h) */
export function formatDateTime(
  value?: string | number | Date | null,
  _locale = "en",
  fallback = "—",
): string {
  const date = value == null ? null : toDate(value);
  if (!date) return fallback;
  return `${formatDate(date, _locale, fallback)} ${formatTime(date, _locale, fallback)}`;
}

export function formatCurrency(
  value?: number | null,
  currency = "USD",
  locale = "en",
  fallback = "—",
): string {
  if (value == null || Number.isNaN(value)) return fallback;
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
  }).format(value);
}

export function formatNumber(
  value?: number | null,
  locale = "en",
  fallback = "—",
): string {
  if (value == null || Number.isNaN(value)) return fallback;
  return new Intl.NumberFormat(locale).format(value);
}
