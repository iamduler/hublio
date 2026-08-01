"use client";

import { useEffect, useRef, useState } from "react";
import { Check, ChevronDown, Globe } from "lucide-react";

export type LocaleSwitcherVariant = "topbar" | "default" | "auth" | "shell";

export type LocaleSwitcherProps<T extends string = string> = {
  locale: T;
  locales: readonly T[];
  labels: Record<T, string>;
  onSelect: (next: T) => void;
  variant?: LocaleSwitcherVariant;
  ariaLabel?: string;
};

export function LocaleSwitcher<T extends string = string>({
  locale,
  locales,
  labels,
  onSelect,
  variant = "default",
  ariaLabel = "Change language",
}: LocaleSwitcherProps<T>) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  function selectLocale(next: T) {
    setOpen(false);
    if (next === locale) return;
    onSelect(next);
  }

  const isDark = variant === "topbar";
  const btnClass =
    variant === "topbar" || variant === "auth" || variant === "shell"
      ? isDark
        ? "flex items-center gap-1.5 text-xs text-(--faint) hover:text-white py-1 px-2 rounded transition-colors"
        : "flex items-center gap-1.5 text-xs text-(--ink-2) hover:text-(--ink) py-1.5 px-2.5 rounded-md border border-(--line) bg-(--white) transition-colors"
      : "flex items-center gap-1.5 text-sm text-(--ink-2) hover:text-(--ink)";

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className={btnClass}
        aria-label={ariaLabel}
        aria-expanded={open}
        aria-haspopup="listbox"
      >
        <Globe size={13} />
        <span className="font-medium">{labels[locale]}</span>
        <ChevronDown
          size={13}
          className={`transition-transform ${open ? "rotate-180" : ""}`}
        />
      </button>
      {open ? (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div
            role="listbox"
            className="absolute top-full right-0 z-50 mt-1 min-w-36 rounded-md border border-(--line) bg-(--white) py-1 shadow-md"
          >
            {locales.map((item) => (
              <button
                key={item}
                type="button"
                role="option"
                aria-selected={item === locale}
                onClick={() => selectLocale(item)}
                className={`flex w-full items-center justify-between px-3 py-1.5 text-left text-xs hover:bg-(--line-2) ${item === locale
                  ? "font-medium text-(--ink)"
                  : "text-(--ink-2)"
                  }`}
              >
                {labels[item]}
                {item === locale ? <Check size={13} /> : null}
              </button>
            ))}
          </div>
        </>
      ) : null}
    </div>
  );
}
