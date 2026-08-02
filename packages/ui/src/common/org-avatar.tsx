"use client";

import { cn } from "../lib/utils";

const ORG_COLORS = [
  "bg-blue-500",
  "bg-violet-500",
  "bg-emerald-500",
  "bg-amber-500",
  "bg-pink-500",
  "bg-teal-500",
  "bg-orange-500",
  "bg-indigo-500",
] as const;

const SIZE = {
  sm: "h-7 w-7 rounded-md text-[10px]",
  md: "h-9 w-9 rounded-lg text-[11px]",
  lg: "h-12 w-12 rounded-xl text-[14px]",
} as const;

/** Derive a short code (1–2 chars) from an org name or id. */
export function orgInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length >= 2) {
    return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
  }
  const single = parts[0] ?? name;
  return single.slice(0, 2).toUpperCase() || "?";
}

function colorFor(seed: string): string {
  let h = 0;
  for (let i = 0; i < seed.length; i++) {
    h = (h * 31 + seed.charCodeAt(i)) >>> 0;
  }
  return ORG_COLORS[h % ORG_COLORS.length]!;
}

export type OrgAvatarProps = {
  /** Display name — used for initials when `code` is omitted. */
  name: string;
  /** Explicit short label (Figma used `code`). */
  code?: string;
  size?: keyof typeof SIZE;
  className?: string;
};

export function OrgAvatar({
  name,
  code,
  size = "md",
  className,
}: OrgAvatarProps) {
  const label = (code ?? orgInitials(name)).slice(0, 2);
  return (
    <div
      className={cn(
        "flex shrink-0 select-none items-center justify-center font-bold text-white",
        SIZE[size],
        colorFor(code ?? name),
        className,
      )}
      aria-hidden
    >
      {label}
    </div>
  );
}
