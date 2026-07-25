"use client";

import * as React from "react";
import { Check, ChevronDown } from "lucide-react";
import { cn } from "../lib/utils";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "./command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "./popover";

/** Prefer Combobox over Select when option count exceeds this. */
export const SELECT_SEARCH_THRESHOLD = 20;

export interface ComboboxOption {
  value: string;
  label: string;
  /** Closed-trigger text; defaults to `label`. Use for compact dialers (flag + code). */
  triggerLabel?: string;
  /** Leading visual (e.g. FlagEmoji with Noto Color Emoji). */
  leading?: React.ReactNode;
  /** Extra text matched by search (optional). */
  keywords?: string;
}

export interface ComboboxProps {
  options: ComboboxOption[];
  value?: string | null;
  onValueChange: (value: string | null) => void;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyText?: string;
  clearable?: boolean;
  clearLabel?: string;
  disabled?: boolean;
  className?: string;
  /** Trigger button className */
  triggerClassName?: string;
  /** Popover panel className (e.g. wider than trigger for dial pickers). */
  contentClassName?: string;
  /** Set true when used inside a Dialog (focus trap). */
  modal?: boolean;
}

export function Combobox({
  options,
  value,
  onValueChange,
  placeholder = "Select…",
  searchPlaceholder = "Search…",
  emptyText = "No results",
  clearable = false,
  clearLabel = "None",
  disabled = false,
  className,
  triggerClassName,
  contentClassName,
  modal = false,
}: ComboboxProps) {
  const [open, setOpen] = React.useState(false);

  const selected = options.find((o) => o.value === value);
  const displayText =
    selected?.triggerLabel ?? selected?.label ?? placeholder;
  const showPlaceholder = !selected;

  return (
    <Popover open={open} onOpenChange={setOpen} modal={modal}>
      <PopoverTrigger asChild>
        <button
          type="button"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn(
            "flex h-10 w-full items-center justify-between rounded-md border border-border bg-card px-3 py-2 text-sm outline-none hover:border-[var(--line-3)] focus:border-primary disabled:cursor-not-allowed disabled:opacity-50",
            triggerClassName,
            className,
          )}
        >
          <span
            className={cn(
              "flex min-w-0 items-center gap-1.5 text-left",
              showPlaceholder && "text-muted-foreground",
            )}
          >
            {!showPlaceholder ? selected?.leading : null}
            <span className="line-clamp-1">{displayText}</span>
          </span>
          <ChevronDown className="h-4 w-4 shrink-0 opacity-50" />
        </button>
      </PopoverTrigger>
      <PopoverContent
        align="start"
        className={cn(
          "w-[var(--radix-popover-trigger-width)] border-border bg-[var(--white)] p-0 shadow-[var(--shadow-pop)]",
          contentClassName,
        )}
        onWheel={(event) => event.stopPropagation()}
      >
        <Command
          className="max-h-[min(320px,var(--radix-popover-content-available-height,320px))]"
          filter={(itemValue, search) => {
            const q = search.trim().toLowerCase();
            if (!q) return 1;
            return itemValue.toLowerCase().includes(q) ? 1 : 0;
          }}
        >
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList className="max-h-[min(280px,var(--radix-popover-content-available-height,280px))] overscroll-contain">
            <CommandEmpty>{emptyText}</CommandEmpty>
            <CommandGroup className="overflow-visible">
              {clearable ? (
                <CommandItem
                  value={`__clear__ ${clearLabel}`}
                  onSelect={() => {
                    onValueChange(null);
                    setOpen(false);
                  }}
                >
                  <span className="text-muted-foreground">{clearLabel}</span>
                  <Check
                    className={cn(
                      "ml-auto h-4 w-4 text-primary",
                      value == null || value === "" ? "opacity-100" : "opacity-0",
                    )}
                  />
                </CommandItem>
              ) : null}
              {options.map((option) => (
                <CommandItem
                  key={option.value}
                  value={`${option.value} ${option.label} ${option.keywords ?? ""}`}
                  onSelect={() => {
                    onValueChange(option.value);
                    setOpen(false);
                  }}
                >
                  <span className="flex min-w-0 items-center gap-1.5">
                    {option.leading}
                    <span className="line-clamp-1">{option.label}</span>
                  </span>
                  <Check
                    className={cn(
                      "ml-auto h-4 w-4 shrink-0 text-primary",
                      value === option.value ? "opacity-100" : "opacity-0",
                    )}
                  />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
