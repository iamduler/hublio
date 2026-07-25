"use client";

import { useLocale } from "next-intl";
import { Check, ChevronDown, Globe } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { usePathname, useRouter } from "@/i18n/navigation";
import { locales, type Locale } from "@/lib/i18n/config";

const LOCALE_LABELS: Record<Locale, string> = {
  en: "English",
  vi: "Tiếng Việt",
};

type Variant = "topbar" | "default" | "auth" | "shell";

export function LocaleSwitcher({ variant = "default" }: { variant?: Variant }) {
  const locale = useLocale() as Locale;
  const router = useRouter();
  const pathname = usePathname();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  function selectLocale(next: Locale) {
    setOpen(false);
    if (next === locale) return;

    const qs =
      typeof window !== "undefined"
        ? window.location.search.replace(/^\?/, "")
        : "";
    if (!qs) {
      router.replace(pathname, { locale: next });
      return;
    }

    const query: Record<string, string | string[]> = {};
    new URLSearchParams(qs).forEach((value, key) => {
      const existing = query[key];
      if (existing === undefined) {
        query[key] = value;
      } else if (Array.isArray(existing)) {
        existing.push(value);
      } else {
        query[key] = [existing, value];
      }
    });

    router.replace({ pathname, query }, { locale: next });
  }

  const isDark = variant === "topbar";
  const btnClass =
    variant === "topbar" || variant === "auth" || variant === "shell"
      ? isDark
        ? "flex items-center gap-1.5 text-xs text-[var(--faint)] hover:text-white py-1 px-2 rounded transition-colors"
        : "flex items-center gap-1.5 text-xs text-[var(--ink-2)] hover:text-[var(--ink)] py-1.5 px-2.5 rounded-md border border-[var(--line)] bg-[var(--white)] transition-colors"
      : "flex items-center gap-1.5 text-sm text-[var(--ink-2)] hover:text-[var(--ink)]";

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className={btnClass}
        aria-label="Change language"
      >
        <Globe size={13} />
        <span className="font-medium">{LOCALE_LABELS[locale]}</span>
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
            className="absolute top-full right-0 mt-1 bg-[var(--white)] border border-[var(--line)] rounded-md shadow-md py-1 z-50 min-w-36"
          >
            {locales.map((item) => (
              <button
                key={item}
                type="button"
                role="option"
                aria-selected={item === locale}
                onClick={() => selectLocale(item)}
                className={`w-full text-left px-3 py-1.5 text-xs hover:bg-[var(--line-2)] flex items-center justify-between ${
                  item === locale
                    ? "text-[var(--ink)] font-medium"
                    : "text-[var(--ink-2)]"
                }`}
              >
                {LOCALE_LABELS[item]}
                {item === locale && <Check size={13} />}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
