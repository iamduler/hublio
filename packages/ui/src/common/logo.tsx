"use client";

import { cn } from "../lib/utils";

export type LogoSize = "sm" | "md" | "lg" | "xl" | "xxl";

const LOGO_HEIGHT: Record<LogoSize, number> = {
  sm: 28,
  md: 36,
  lg: 52,
  xl: 64,
  xxl: 105,
};

const LOGO_PADDING: Record<LogoSize, string> = {
  sm: "px-2 py-1.5",
  md: "px-4 py-3",
  lg: "px-6 py-4",
  xl: "px-8 py-6",
  xxl: "px-10 py-8",
};

const ASPECT = 564 / 182;

export type LogoProps = {
  size?: LogoSize;
  className?: string;
  /** Image path (defaults to public brand asset). */
  src?: string;
  alt?: string;
  priority?: boolean;
};

/** Full wordmark + tagline. Transparent PNG — always on a white plate. */
export function Logo({
  size = "md",
  className,
  src = "/logo/logo.png",
  alt = "Hublio",
  priority = false,
}: LogoProps) {
  const height = LOGO_HEIGHT[size];
  const width = Math.round(height * ASPECT);

  return (
    <div
      className={cn(
        "inline-flex items-center justify-center rounded-sm bg-white",
        LOGO_PADDING[size],
        className,
      )}
    >
      {/* eslint-disable-next-line @next/next/no-img-element -- package is framework-agnostic */}
      <img
        src={src}
        alt={alt}
        width={width}
        height={height}
        className="h-auto w-auto max-w-full"
        style={{ height, width: "auto" }}
        decoding="async"
        loading={priority ? "eager" : "lazy"}
        fetchPriority={priority ? "high" : undefined}
      />
    </div>
  );
}
