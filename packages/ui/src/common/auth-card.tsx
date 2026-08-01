"use client";

import * as React from "react";
import { cn } from "../lib/utils";

export type AuthCardProps = {
  children: React.ReactNode;
  className?: string;
  contentClassName?: string;
  /** Footer below the card (e.g. copyright). Pass `null` to hide. */
  footer?: React.ReactNode;
};

export function AuthCard({
  children,
  className,
  contentClassName,
  footer = (
    <p className="mt-5 text-center text-[11px] text-(--faint)">
      © {new Date().getFullYear()} Hublio
    </p>
  ),
}: AuthCardProps) {
  return (
    <div className={cn("w-full max-w-105", className)}>
      <div
        className={cn(
          "w-full rounded-xl border border-(--line) bg-(--white) p-8",
          contentClassName,
        )}
        style={{
          boxShadow:
            "0 1px 3px rgba(0,0,0,0.07), 0 4px 20px rgba(0,0,0,0.05)",
        }}
      >
        {children}
      </div>
      {footer}
    </div>
  );
}

export type AuthFormHeaderProps = {
  logo?: React.ReactNode;
  title: React.ReactNode;
  subtitle?: React.ReactNode;
  className?: string;
};

/** Centered logo → title → subtitle stack for auth forms. */
export function AuthFormHeader({
  logo,
  title,
  subtitle,
  className,
}: AuthFormHeaderProps) {
  return (
    <div className={cn("mb-7 flex flex-col items-center text-center", className)}>
      {logo}
      <h1 className="mt-5 text-[17px] font-semibold tracking-tight text-(--ink)">
        {title}
      </h1>
      {subtitle ? (
        <p className="mt-1.5 text-[13px] text-muted-foreground">{subtitle}</p>
      ) : null}
    </div>
  );
}
