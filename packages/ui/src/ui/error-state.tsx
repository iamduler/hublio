"use client";

import type { LucideIcon } from "lucide-react";
import { AlertCircle } from "lucide-react";
import { EmptyState } from "./empty-state";
import { Button } from "./button";

export type ErrorStateProps = {
  title: string;
  description?: string;
  onRetry?: () => void;
  retryLabel?: string;
  icon?: LucideIcon;
  size?: "sm" | "md";
  className?: string;
};

/**
 * Inline error with retry — never show raw backend exception text.
 * Prefer `useApiErrorMessage()` for `description`.
 */
export function ErrorState({
  title,
  description,
  onRetry,
  retryLabel = "Retry",
  icon = AlertCircle,
  size = "md",
  className,
}: ErrorStateProps) {
  return (
    <EmptyState
      icon={icon}
      title={title}
      description={description}
      size={size}
      className={className}
      action={
        onRetry ? (
          <Button variant="outline" size="sm" onClick={onRetry}>
            {retryLabel}
          </Button>
        ) : undefined
      }
    />
  );
}
