"use client";
import * as React from "react";
import * as AvatarPrimitive from "@radix-ui/react-avatar";
import { cn } from "../lib/utils";

const COLORS = [
  "#F7462E",
  "#2563EB",
  "#7C3AED",
  "#0F9D58",
  "#D97706",
  "#0284C7",
  "#DB2777",
];
function avColor(s: string) {
  let h = 0;
  for (const c of s) h = (h * 31 + c.charCodeAt(0)) >>> 0;
  return COLORS[h % COLORS.length];
}
function initials(name: string) {
  return (
    name
      .split(" ")
      .filter(Boolean)
      .slice(-2)
      .map((w) => w[0])
      .join("")
      .toUpperCase()
      .slice(0, 2) || "?"
  );
}

const Avatar = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Root>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Root
    ref={ref}
    className={cn(
      "relative flex shrink-0 overflow-hidden rounded-md",
      className,
    )}
    {...props}
  />
));
Avatar.displayName = AvatarPrimitive.Root.displayName;

const AvatarImage = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Image>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Image>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Image
    ref={ref}
    className={cn("aspect-square h-full w-full", className)}
    {...props}
  />
));
AvatarImage.displayName = AvatarPrimitive.Image.displayName;

const AvatarFallback = React.forwardRef<
  React.ElementRef<typeof AvatarPrimitive.Fallback>,
  React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Fallback>
>(({ className, ...props }, ref) => (
  <AvatarPrimitive.Fallback
    ref={ref}
    className={cn(
      "flex h-full w-full items-center justify-center rounded-md text-white text-sm font-medium",
      className,
    )}
    {...props}
  />
));
AvatarFallback.displayName = AvatarPrimitive.Fallback.displayName;

interface UserAvatarProps {
  name: string;
  size?: number;
  src?: string;
  className?: string;
  /** hash = colorful; navy = brand navy square */
  tone?: "hash" | "navy" | "primary";
  /** Override generated initials */
  initialsText?: string;
}

function UserAvatar({
  name,
  size = 36,
  src,
  className,
  tone = "hash",
  initialsText,
}: UserAvatarProps) {
  const label = initialsText ?? initials(name);
  return (
    <Avatar
      className={cn("flex-none", className)}
      style={{ width: size, height: size }}
    >
      {src && <AvatarImage src={src} alt={name} />}
      <AvatarFallback
        style={{
          background:
            tone === "navy"
              ? "var(--navy-900)"
              : tone === "primary"
                ? "var(--primary)"
                : avColor(name),
          fontSize: size * 0.38,
        }}
      >
        {label}
      </AvatarFallback>
    </Avatar>
  );
}

export { Avatar, AvatarImage, AvatarFallback, UserAvatar };
