"use client";

import * as React from "react";
import type { LucideIcon } from "lucide-react";
import { cn } from "../lib/utils";

export type EmptyStateProps = {
  icon?: LucideIcon;
  title: React.ReactNode;
  description?: React.ReactNode;
  action?: React.ReactNode;
  size?: "sm" | "md";
  className?: string;
};

function EmptyState({
  icon: Icon,
  title,
  description,
  action,
  size = "md",
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        "empty-state",
        size === "sm" && "px-4 py-8",
        className,
      )}
    >
      {Icon ? (
        <div className={cn("ei", size === "sm" && "!h-12 !w-12 !rounded-xl")}>
          <Icon size={size === "sm" ? 22 : 28} />
        </div>
      ) : null}
      <h2 className={cn(size === "sm" && "!text-base")}>{title}</h2>
      {description ? <p>{description}</p> : null}
      {action}
    </div>
  );
}

export { EmptyState };
