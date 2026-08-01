"use client";

import { Skeleton } from "./skeleton";
import { cn } from "../lib/utils";

export type LoadingStateProps = {
  /** Number of skeleton rows (default 3). */
  rows?: number;
  label?: string;
  className?: string;
};

/**
 * Prefer skeleton over blank screens (DESIGN_SYSTEM loading pattern).
 */
export function LoadingState({
  rows = 3,
  label,
  className,
}: LoadingStateProps) {
  return (
    <div
      className={cn("flex flex-col gap-3 py-6", className)}
      role="status"
      aria-busy="true"
      aria-label={typeof label === "string" ? label : undefined}
    >
      {label ? (
        <p className="text-sm text-(--ink-2)">{label}</p>
      ) : null}
      {Array.from({ length: rows }, (_, i) => (
        <Skeleton
          key={i}
          className="h-12 w-full bg-(--line-2)"
          style={{ maxWidth: i === rows - 1 ? "70%" : "100%" }}
        />
      ))}
    </div>
  );
}
