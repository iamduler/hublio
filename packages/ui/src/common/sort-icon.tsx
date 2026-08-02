"use client";

import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { cn } from "../lib/utils";

export type SortDirection = "asc" | "desc";

export type SortState = {
  column: string;
  direction: SortDirection;
} | null;

export function SortIcon({
  active,
  direction,
  className,
}: {
  active: boolean;
  direction?: SortDirection;
  className?: string;
}) {
  if (!active) {
    return (
      <ArrowUpDown
        size={11}
        className={cn("text-(--faint)", className)}
        aria-hidden
      />
    );
  }
  if (direction === "asc") {
    return (
      <ArrowUp size={11} className={cn("text-primary", className)} aria-hidden />
    );
  }
  return (
    <ArrowDown
      size={11}
      className={cn("text-primary", className)}
      aria-hidden
    />
  );
}

export function toggleSort(
  current: SortState,
  column: string,
): SortState {
  if (!current || current.column !== column) {
    return { column, direction: "asc" };
  }
  if (current.direction === "asc") {
    return { column, direction: "desc" };
  }
  return null;
}
