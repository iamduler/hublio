"use client";

import * as React from "react";
import { cn } from "../lib/utils";
import { AppSidebarTrigger } from "./app-sidebar";

export type AppShellProps = {
  sidebar: React.ReactNode;
  header: React.ReactNode;
  children: React.ReactNode;
  className?: string;
  mainClassName?: string;
};

/** Sidebar + header + main frame. Wrap with `AppSidebarProvider` in the app layout. */
export function AppShell({
  sidebar,
  header,
  children,
  className,
  mainClassName,
}: AppShellProps) {
  return (
    <div className={cn("flex min-h-dvh", className)}>
      {sidebar}
      <div className="flex min-w-0 flex-1 flex-col">
        {header}
        <main className={cn("flex-1 p-4 md:p-6", mainClassName)}>
          {children}
        </main>
      </div>
    </div>
  );
}

export type AppHeaderProps = {
  /** Mobile menu label for `AppSidebarTrigger`. */
  menuLabel: string;
  /** Content after the menu trigger (workspace switcher, org name, title, …). */
  leading?: React.ReactNode;
  /** Right-side actions (theme, locale, logout, …). */
  trailing?: React.ReactNode;
  /** Optional email shown between leading and trailing on `sm+`. */
  email?: string | null;
  className?: string;
};

export function AppHeader({
  menuLabel,
  leading,
  trailing,
  email,
  className,
}: AppHeaderProps) {
  return (
    <header
      className={cn(
        "flex h-(--nav-h) items-center justify-between gap-3 border-b border-(--line) bg-(--white) px-3 md:gap-4 md:px-5",
        className,
      )}
    >
      <div className="flex min-w-0 items-center gap-2 md:gap-4">
        <AppSidebarTrigger menuLabel={menuLabel} />
        {leading}
      </div>
      <div className="flex shrink-0 items-center gap-2">
        {email ? (
          <span className="hidden text-xs text-(--muted-clr) sm:inline">
            {email}
          </span>
        ) : null}
        {trailing}
      </div>
    </header>
  );
}
