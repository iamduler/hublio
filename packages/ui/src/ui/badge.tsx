import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "../lib/utils";

const badgeVariants = cva(
  "inline-flex items-center gap-1 rounded-md text-[11.5px] font-medium px-2.5 py-1 whitespace-nowrap",
  {
    variants: {
      variant: {
        default: "badge-primary",
        green: "badge-green",
        amber: "badge-amber",
        sky: "badge-sky",
        violet: "badge-violet",
        gray: "badge-gray",
        danger: "badge-danger",
        outline: "border border-border text-muted-foreground",
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
      {dot && <span className="w-1.5 h-1.5 rounded-md bg-current" />}
      {Icon ? <Icon size={13} /> : null}
      {children}
    </div>
  );
}

export { Badge, badgeVariants };
