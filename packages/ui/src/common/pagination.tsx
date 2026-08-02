"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "../lib/utils";

export type PaginationProps = {
  page: number;
  total: number;
  perPage: number;
  onChange: (page: number) => void;
  /** Override “Showing X–Y of Z” — pass fully localized string. */
  showingLabel?: string;
  prevLabel?: string;
  nextLabel?: string;
  className?: string;
  /** When true, no top border (use for toolbar placement). */
  borderless?: boolean;
};

function pageWindow(current: number, totalPages: number, size = 7): number[] {
  if (totalPages <= size) {
    return Array.from({ length: totalPages }, (_, i) => i + 1);
  }
  const half = Math.floor(size / 2);
  let start = Math.max(1, current - half);
  const end = Math.min(totalPages, start + size - 1);
  start = Math.max(1, end - size + 1);
  return Array.from({ length: end - start + 1 }, (_, i) => start + i);
}

export function Pagination({
  page,
  total,
  perPage,
  onChange,
  showingLabel,
  prevLabel = "Prev",
  nextLabel = "Next",
  className,
  borderless,
}: PaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / perPage) || 1);
  const safePage = Math.min(Math.max(1, page), totalPages);
  const from = total === 0 ? 0 : (safePage - 1) * perPage + 1;
  const to = Math.min(safePage * perPage, total);
  const pages = pageWindow(safePage, totalPages);

  const defaultShowing =
    total === 0
      ? "Showing 0 of 0"
      : `Showing ${from}–${to} of ${total.toLocaleString()}`;

  const btnCls =
    "inline-flex h-7 min-w-7 items-center justify-center gap-1 rounded-lg px-2 text-xs font-medium transition-colors";

  return (
    <div
      className={cn(
        "flex flex-col gap-2 px-4 py-3 sm:flex-row sm:items-center sm:justify-between",
        !borderless && "border-t border-(--line)",
        className,
      )}
    >
      <p className="text-xs text-(--muted-clr)">
        {showingLabel ?? defaultShowing}
      </p>
      <div className="flex flex-wrap items-center gap-1">
        <button
          type="button"
          disabled={safePage <= 1}
          onClick={() => onChange(safePage - 1)}
          className={cn(
            btnCls,
            "text-(--ink-2) hover:bg-(--line-2) disabled:cursor-not-allowed disabled:opacity-40",
          )}
        >
          <ChevronLeft size={13} />
          {prevLabel}
        </button>
        {pages.map((p) => (
          <button
            key={p}
            type="button"
            onClick={() => onChange(p)}
            className={cn(
              btnCls,
              p === safePage
                ? "bg-primary text-primary-foreground"
                : "text-(--ink-2) hover:bg-(--line-2)",
            )}
          >
            {p}
          </button>
        ))}
        <button
          type="button"
          disabled={safePage >= totalPages}
          onClick={() => onChange(safePage + 1)}
          className={cn(
            btnCls,
            "text-(--ink-2) hover:bg-(--line-2) disabled:cursor-not-allowed disabled:opacity-40",
          )}
        >
          {nextLabel}
          <ChevronRight size={13} />
        </button>
      </div>
    </div>
  );
}
