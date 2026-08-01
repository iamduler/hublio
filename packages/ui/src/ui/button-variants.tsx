import { cva, type VariantProps } from "class-variance-authority";

/**
 * Shared button class helper — safe to call from Server Components.
 * Keep this module free of `"use client"`.
 *
 * Core roles (reuse these):
 * - default  — primary CTA (create / submit / save)
 * - outline  — secondary toolbar (export / filter)
 * - navy     — strong dark CTA
 * - ghost    — quiet bordered chrome
 *
 * Special: soft | danger-soft | link
 * Icon-only: size icon | icon-sm (row actions)
 */
export const buttonVariants = cva(
  // `no-underline` / `text-inherit` keep <Link>/<a> visually identical to <button>.
  "inline-flex items-center justify-center gap-1.5 whitespace-nowrap rounded-md text-sm font-medium transition-all cursor-pointer no-underline text-inherit box-border disabled:opacity-50 disabled:pointer-events-none",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground shadow-sm hover:opacity-90 active:scale-[.98]",
        outline:
          "border border-(--line) bg-(--white) text-(--ink-2) hover:bg-(--line-2)",
        navy: "bg-[var(--navy-900)] text-white hover:bg-[#11111f]",
        ghost:
          "border border-border bg-(--white) text-foreground hover:bg-secondary",
        soft: "text-primary hover:opacity-90",
        "danger-soft":
          "bg-border-(--danger-soft) text-border-(--danger) hover:opacity-90",
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
          "border-0 bg-transparent text-(--faint) shadow-none hover:bg-(--line-2) hover:text-(--ink-2)",
      },
    ],
    defaultVariants: { variant: "default", size: "default" },
  },
);

export type ButtonVariantProps = VariantProps<typeof buttonVariants>;
