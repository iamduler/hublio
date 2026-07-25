"use client";

import * as React from "react";
import { Search } from "lucide-react";
import { cn } from "../lib/utils";
import { Input, type InputProps } from "./input";

export type SearchFieldProps = Omit<InputProps, "size"> & {
  className?: string;
  inputClassName?: string;
  size?: "sm" | "md" | "lg";
};

const SIZE = {
  sm: {
    icon: 12,
    iconClass: "left-2.5",
    inputClass: "h-auto pl-7 pr-2 py-1.5 text-xs",
  },
  md: {
    icon: 14,
    iconClass: "left-3",
    inputClass: "h-auto pl-9 pr-3 py-2",
  },
  lg: {
    icon: 18,
    iconClass: "left-4",
    inputClass: "h-12 pl-11 pr-4 text-base",
  },
} as const;

const SearchField = React.forwardRef<HTMLInputElement, SearchFieldProps>(
  ({ className, inputClassName, size = "md", ...props }, ref) => {
    const s = SIZE[size];
    return (
      <div className={cn("relative", className)}>
        <Search
          size={s.icon}
          className={cn(
            "pointer-events-none absolute top-1/2 -translate-y-1/2 text-[var(--faint)]",
            s.iconClass,
          )}
          aria-hidden
        />
        <Input ref={ref} className={cn(s.inputClass, inputClassName)} {...props} />
      </div>
    );
  },
);
SearchField.displayName = "SearchField";

export { SearchField };
