"use client";

import { useEffect, useRef, useState } from "react";
import { Check, ChevronDown, SunMoon } from "lucide-react";
import type { UserTheme } from "../lib/theme";

export type ThemeSwitcherVariant = "topbar" | "auth" | "shell";

export type ThemeOption = {
  value: UserTheme;
  label: string;
};

export type ThemeSwitcherProps = {
  theme: UserTheme;
  options: ThemeOption[];
  onSelect: (next: UserTheme) => void | Promise<void>;
  isReady?: boolean;
  variant?: ThemeSwitcherVariant;
  ariaLabel?: string;
  /** Label shown while preferences are loading */
  fallbackLabel?: string;
};

export function ThemeSwitcher({
  theme,
  options,
  onSelect,
  isReady = true,
  variant = "shell",
  ariaLabel = "Change appearance",
  fallbackLabel,
}: ThemeSwitcherProps) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  async function selectTheme(next: UserTheme) {
    setOpen(false);
    if (next === theme) return;
    await onSelect(next);
  }

  const isDark = variant === "topbar";
  const btnClass = isDark
    ? "flex items-center gap-1.5 text-xs text-(--faint) hover:text-white py-1 px-2 rounded transition-colors"
    : "flex items-center gap-1.5 text-xs text-(--ink-2) hover:text-(--ink) py-1.5 px-2.5 rounded-md border border-(--line) bg-(--white) transition-colors";

  const currentLabel =
    options.find((o) => o.value === theme)?.label ??
    fallbackLabel ??
    options[0]?.label ??
    theme;
  const label = isReady ? currentLabel : (fallbackLabel ?? currentLabel);

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
        <SunMoon size={13} />
        <span className="font-medium">{label}</span>
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
            className="absolute top-full right-0 z-50 mt-1 min-w-36 rounded-md border border-(--line) bg-(--white) py-1 shadow-lg"
          >
            {options.map((option) => (
              <button
                key={option.value}
                type="button"
                role="option"
                aria-selected={option.value === theme}
                onClick={() => void selectTheme(option.value)}
                className={`flex w-full items-center justify-between px-3 py-1.5 text-left text-xs hover:bg-(--line-2) ${option.value === theme
                  ? "font-medium text-(--ink)"
                  : "text-(--ink-2)"
                  }`}
              >
                {option.label}
                {option.value === theme ? <Check size={13} /> : null}
              </button>
            ))}
          </div>
        </>
      ) : null}
    </div>
  );
}
