import * as React from "react";
import { ChevronRight } from "lucide-react";
import { cn } from "../lib/utils";

export type BreadcrumbItem = {
  label: React.ReactNode;
  href?: string;
};

export type BreadcrumbProps = {
  items: BreadcrumbItem[];
  /** Optional link renderer (Next.js Link, etc.). */
  renderLink?: (props: {
    href: string;
    className: string;
    children: React.ReactNode;
  }) => React.ReactNode;
  className?: string;
};

/**
 * Thin trail: Admin Console › Organizations
 */
export function Breadcrumb({ items, renderLink, className }: BreadcrumbProps) {
  if (items.length === 0) return null;

  return (
    <nav
      aria-label="Breadcrumb"
      className={cn(
        "mb-1.5 flex flex-wrap items-center gap-1.5 text-xs text-(--faint)",
        className,
      )}
    >
      {items.map((item, index) => {
        const isLast = index === items.length - 1;
        return (
          <React.Fragment key={index}>
            {index > 0 ? (
              <ChevronRight size={12} className="shrink-0 text-(--faint)" />
            ) : null}
            {item.href && !isLast && renderLink ? (
              renderLink({
                href: item.href,
                className: "text-(--faint) no-underline hover:text-(--ink-2)",
                children: item.label,
              })
            ) : item.href && !isLast ? (
              <a
                href={item.href}
                className="text-(--faint) no-underline hover:text-(--ink-2)"
              >
                {item.label}
              </a>
            ) : (
              <span
                className={cn(
                  isLast && "font-medium text-(--ink-2)",
                  !isLast && "text-(--faint)",
                )}
                aria-current={isLast ? "page" : undefined}
              >
                {item.label}
              </span>
            )}
          </React.Fragment>
        );
      })}
    </nav>
  );
}
