"use client";

import * as React from "react";
import { MoreHorizontal } from "lucide-react";
import { cn } from "../lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";

export type ActionMenuItem = {
  label: string;
  icon?: React.ReactNode;
  onSelect: () => void;
  danger?: boolean;
  /** Render a separator before this item. */
  separator?: boolean;
  disabled?: boolean;
};

export type ActionMenuProps = {
  items: ActionMenuItem[];
  /** Accessible label for the trigger. */
  label?: string;
  /**
   * When true, trigger is always visible.
   * Default: visible on `group-hover` / focus (Figma row chrome).
   */
  alwaysVisible?: boolean;
  className?: string;
  align?: "start" | "center" | "end";
};

export function ActionMenu({
  items,
  label = "Actions",
  alwaysVisible,
  className,
  align = "end",
}: ActionMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          aria-label={label}
          className={cn(
            "inline-flex h-7 w-7 items-center justify-center rounded-lg text-(--faint) transition-colors hover:bg-(--line-2) hover:text-(--ink-2)",
            !alwaysVisible &&
              "opacity-0 focus-visible:opacity-100 group-hover:opacity-100 data-[state=open]:opacity-100",
            className,
          )}
          onClick={(e) => e.stopPropagation()}
        >
          <MoreHorizontal size={14} />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align={align}
        className="min-w-48 rounded-xl p-1"
        onClick={(e) => e.stopPropagation()}
      >
        {items.map((item, i) => (
          <React.Fragment key={`${item.label}-${i}`}>
            {item.separator ? <DropdownMenuSeparator className="my-1" /> : null}
            <DropdownMenuItem
              disabled={item.disabled}
              onSelect={(e) => {
                e.preventDefault();
                item.onSelect();
              }}
              className={cn(
                "gap-2.5 px-3 py-2 text-[13px]",
                item.danger &&
                  "text-destructive focus:bg-destructive/10 focus:text-destructive",
              )}
            >
              {item.icon ? (
                <span className="shrink-0 text-current">{item.icon}</span>
              ) : null}
              {item.label}
            </DropdownMenuItem>
          </React.Fragment>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
