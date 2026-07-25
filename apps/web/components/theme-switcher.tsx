"use client";

import { useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { Check, ChevronDown, SunMoon } from "lucide-react";
import { usePreferredTheme } from "@/hooks/use-preferred-theme";
import { THEME_OPTIONS, type UserTheme } from "@/lib/preferences/types";

type Variant = "topbar" | "auth" | "shell";

export function ThemeSwitcher({ variant = "shell" }: { variant?: Variant }) {
  const t = useTranslations("common.settings.appearance.themes");
  const { theme, setTheme, isReady } = usePreferredTheme();
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
    await setTheme(next);
  }

  const isDark = variant === "topbar";
  const btnClass = isDark
    ? "flex items-center gap-1.5 text-xs text-[var(--faint)] hover:text-white py-1 px-2 rounded transition-colors"
    : "flex items-center gap-1.5 text-xs text-[var(--ink-2)] hover:text-[var(--ink)] py-1.5 px-2.5 rounded-md border border-[var(--line)] bg-[var(--white)] transition-colors";

  const label = isReady ? t(theme) : t("system");

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className={btnClass}
        aria-label="Change appearance"
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
      {open && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setOpen(false)} />
          <div
            role="listbox"
            className="absolute top-full right-0 mt-1 bg-[var(--white)] border border-[var(--line)] rounded-md shadow-lg py-1 z-50 min-w-36"
          >
            {THEME_OPTIONS.map((option) => (
              <button
                key={option}
                type="button"
                role="option"
                aria-selected={option === theme}
                onClick={() => void selectTheme(option)}
                className={`w-full text-left px-3 py-1.5 text-xs hover:bg-[var(--line-2)] flex items-center justify-between ${
                  option === theme
                    ? "text-[var(--ink)] font-medium"
                    : "text-[var(--ink-2)]"
                }`}
              >
                {t(option)}
                {option === theme && <Check size={13} />}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
