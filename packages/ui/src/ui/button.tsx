"use client";

import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cn } from "../lib/utils";
import {
  buttonVariants,
  type ButtonVariantProps,
} from "./button-variants";

export interface ButtonProps
  extends
    React.ButtonHTMLAttributes<HTMLButtonElement>,
    ButtonVariantProps {
  asChild?: boolean;
}

/**
 * Radix Slot (asChild) requires exactly one React element child. JSX whitespace
 * between tags becomes text nodes and trips Slot under React 19 — strip those.
 */
function slotChild(children: React.ReactNode): React.ReactNode {
  const elements = React.Children.toArray(children).filter(React.isValidElement);
  return elements.length === 1 ? elements[0] : children;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, children, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    const content = asChild ? slotChild(children) : children;
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
        >
          {content}
        </Comp>
      );
    }
    return (
      <Comp
        ref={ref}
        className={cn(buttonVariants({ variant, size }), className)}
        {...props}
      >
        {content}
      </Comp>
    );
  },
);
Button.displayName = "Button";

export { Button, buttonVariants };
