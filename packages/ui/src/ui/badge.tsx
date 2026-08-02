import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "../lib/utils";

/**
 * shadcn-style badge chips using design tokens (soft fill + ink).
 */
const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-md border px-2.5 py-0.5 text-[11.5px] font-semibold whitespace-nowrap transition-colors",
  {
    variants: {
      variant: {
        default: "border-transparent bg-primary text-primary-foreground",
        green:
          "border-transparent bg-(--success-soft) text-(--success)",
        amber: "border-transparent bg-(--amber-soft) text-(--amber)",
        sky: "border-transparent bg-(--sky-soft) text-(--sky)",
        violet: "border-transparent bg-(--violet-soft) text-(--violet)",
        gray: "border-(--line) bg-(--line-2) text-(--ink-2)",
        danger: "border-transparent bg-(--danger-soft) text-(--danger)",
        outline: "border-border bg-transparent text-muted-foreground",
      },
    },
    defaultVariants: { variant: "default" },
  },
);

export interface BadgeProps
  extends
    React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {
  dot?: boolean;
  icon?: React.ElementType;
}

function Badge({
  className,
  variant,
  dot,
  icon: Icon,
  children,
  ...props
}: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props}>
      {dot ? (
        <span
          className="h-1.5 w-1.5 shrink-0 rounded-full bg-current"
          aria-hidden
        />
      ) : null}
      {Icon ? <Icon size={13} /> : null}
      {children}
    </div>
  );
}

export { Badge, badgeVariants };
