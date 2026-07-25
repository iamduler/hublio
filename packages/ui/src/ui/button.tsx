"use client";

import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "../lib/utils";

/**
 * Core roles (reuse these):
 * - default  — primary CTA (create / submit / save)
 * - outline  — secondary toolbar (export / filter)
 * - navy     — strong dark CTA
 * - ghost    — quiet bordered chrome
 *
 * Special: soft | danger-soft | link
 * Icon-only: size icon | icon-sm (row actions)
 */
const buttonVariants = cva(
  // `no-underline` / `text-inherit` keep asChild <Link>/<a> visually identical to <button>.
  "inline-flex items-center justify-center gap-1.5 whitespace-nowrap rounded-md text-sm font-medium transition-all cursor-pointer no-underline text-inherit box-border disabled:opacity-50 disabled:pointer-events-none",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground shadow-sm hover:opacity-90 active:scale-[.98]",
        outline:
          "border border-[var(--line)] bg-[var(--white)] text-[var(--ink-2)] hover:bg-[var(--line-2)]",
        navy: "bg-[var(--navy-900)] text-white hover:bg-[#11111f]",
        ghost:
          "border border-border bg-[var(--white)] text-foreground hover:bg-secondary",
        soft: "text-primary hover:opacity-90",
        "danger-soft":
          "bg-[var(--danger-soft)] text-[var(--danger)] hover:opacity-90",
        link: "rounded-none underline-offset-4 hover:underline text-primary gap-0",
      },
      size: {
        default: "h-10 px-4",
        sm: "h-10 px-3 text-sm",
        lg: "h-11 px-6",
        icon: "h-10 w-9 gap-0",
        "icon-sm": "h-7 w-7 gap-0 p-0",
      },
    },
    compoundVariants: [
      {
        size: "icon-sm",
        className:
          "border-0 bg-transparent text-[var(--faint)] shadow-none hover:bg-[var(--line-2)] hover:text-[var(--ink-2)]",
      },
    ],
    defaultVariants: { variant: "default", size: "default" },
  },
);

export interface ButtonProps
  extends
    React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    if (variant === "soft") {
      return (
        <Comp
          ref={ref}
          className={cn(
            "inline-flex items-center justify-center gap-1.5 whitespace-nowrap rounded-md text-sm font-medium transition-all cursor-pointer no-underline box-border disabled:opacity-50 disabled:pointer-events-none",
            className,
          )}
          style={{
            background: "color-mix(in srgb, var(--primary), #fff 88%)",
            color: "color-mix(in srgb, var(--primary), #000 18%)",
            height: size === "sm" ? 32 : 40,
            padding: size === "sm" ? "0 12px" : "0 16px",
            fontSize: size === "sm" ? 12.5 : 13.5,
          }}
          {...props}
        />
      );
    }
    return (
      <Comp
        ref={ref}
        className={cn(buttonVariants({ variant, size }), className)}
        {...props}
      />
    );
  },
);
Button.displayName = "Button";

export { Button, buttonVariants };
