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
  className,
}: {
  label: string;
  value: React.ReactNode;
  icon?: LucideIcon;
  hint?: React.ReactNode;
  className?: string;
}) {
  return (
    <Card className={cn("shadow-sm", className)}>
      <CardContent className="flex items-center gap-4 p-4">
        {Icon ? (
          <div className="grid h-10 w-10 place-items-center rounded-md bg-[var(--primary-soft)] text-[var(--primary-ink)]">
            <Icon size={18} />
          </div>
        ) : null}
        <div className="min-w-0">
          <p className="text-xs font-medium uppercase tracking-wide text-[var(--muted-clr)]">
            {label}
          </p>
          <p className="font-display text-2xl font-semibold text-[var(--ink)]">
            {value}
          </p>
          {hint ? (
            <p className="text-xs text-[var(--muted-clr)]">{hint}</p>
          ) : null}
        </div>
      </CardContent>
    </Card>
  );
}
