"use client";

import { ChevronDown } from "lucide-react";
import { cn } from "../lib/utils";

export type FilterSelectOption = {
  value: string;
  label: string;
};

export type FilterSelectProps = {
  placeholder: string;
  value: string;
  onChange: (value: string) => void;
  options: FilterSelectOption[] | string[];
  className?: string;
  id?: string;
  "aria-label"?: string;
};

function normalizeOptions(
  options: FilterSelectOption[] | string[],
): FilterSelectOption[] {
  return options.map((o) =>
    typeof o === "string" ? { value: o, label: o } : o,
  );
}

export function FilterSelect({
  placeholder,
  value,
  onChange,
  options,
  className,
  id,
  "aria-label": ariaLabel,
}: FilterSelectProps) {
  const opts = normalizeOptions(options);
  return (
    <div className={cn("relative", className)}>
      <select
        id={id}
        aria-label={ariaLabel ?? placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="h-8 appearance-none rounded-lg border border-(--line) bg-(--white) py-0 pl-3 pr-7 text-[13px] font-medium text-(--ink-2) transition-colors hover:border-(--line-3) focus:border-primary focus:outline-none"
      >
        <option value="">{placeholder}</option>
        {opts.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
      <ChevronDown
        size={12}
        className="pointer-events-none absolute right-2 top-1/2 -translate-y-1/2 text-(--faint)"
        aria-hidden
      />
    </div>
  );
}
