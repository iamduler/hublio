import * as React from "react"
import { cn } from "../lib/utils"

export type TextareaProps = React.TextareaHTMLAttributes<HTMLTextAreaElement>;

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(({ className, ...props }, ref) => (
  <textarea
    className={cn(
      "flex min-h-[80px] w-full rounded-md border border-border bg-card px-3 py-2.5 text-sm leading-relaxed transition-all outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50 hover:border-[var(--line-3)] focus:border-primary focus:shadow-[0_0_0_3px_color-mix(in_srgb,var(--primary),transparent_87%)] resize-y",
      className
    )}
    ref={ref}
    {...props}
  />
))
Textarea.displayName = "Textarea"

export { Textarea }
