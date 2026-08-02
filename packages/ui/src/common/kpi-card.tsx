import * as React from "react";
import type { LucideIcon } from "lucide-react";
import { Card, CardContent } from "../ui/card";
import { cn } from "../lib/utils";

/** Compact metric card ported from the Figma `KPICard`. */
export function KpiCard({
  label,
  value,
  icon: Icon,
  hint,
  unavailable,
  className,
}: {
  label: string;
  value: React.ReactNode;
  icon?: LucideIcon;
  hint?: React.ReactNode;
  /** Muted placeholder when backend metrics are not wired yet. */
  unavailable?: boolean;
  className?: string;
}) {
  return (
    <Card
      className={cn(
        "shadow-sm",
        unavailable && "border-dashed opacity-80",
        className,
      )}
    >
      <CardContent className="flex items-center gap-4 p-4">
        {Icon ? (
          <div className="grid h-10 w-10 place-items-center rounded-md bg-[var(--primary-soft)] text-[var(--primary-ink)]">
            <Icon size={18} />
          </div>
        ) : null}
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wide text-(--muted-clr)">
            {label}
          </p>
          <p
            className={cn(
              "font-display text-2xl font-semibold",
              unavailable ? "text-(--faint)" : "text-(--ink)",
            )}
          >
            {value}
          </p>
          {hint ? (
            <p className="text-xs text-(--muted-clr)">{hint}</p>
          ) : null}
        </div>
      </CardContent>
    </Card>
  );
}
