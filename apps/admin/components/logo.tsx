"use client";

import Image from "next/image";
import { cn } from "@/lib/utils";

const LOGO_HEIGHT: Record<"sm" | "md" | "lg", number> = {
  sm: 28,
  md: 36,
  lg: 52,
};

/** Full wordmark + tagline. Transparent PNG — always on a white plate. */
export function Logo({
  size = "md",
  className,
  priority = false,
}: {
  size?: "sm" | "md" | "lg";
  className?: string;
  priority?: boolean;
}) {
  const height = LOGO_HEIGHT[size];
  const width = Math.round(height * (564 / 182));

  return (
    <div
      className={cn(
        "inline-flex items-center justify-center rounded-sm bg-white",
        size === "lg" ? "px-3 py-2" : "px-2 py-1.5",
        className,
      )}
    >
      <Image
        src="/logo/logo.png"
        alt="Hublio"
        width={width}
        height={height}
        className="h-auto w-auto max-w-full"
        style={{ height, width: "auto" }}
        priority={priority}
      />
    </div>
  );
}
