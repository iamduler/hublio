"use client";

import { useTranslations } from "next-intl";
import {
  Activity,
  BarChart3,
  Blocks,
  Cable,
  GitBranch,
  KeyRound,
  LayoutDashboard,
  PlayCircle,
  Settings,
  Users,
} from "lucide-react";
import { Link, usePathname } from "@/i18n/navigation";
import { cn } from "@/lib/utils";

type NavLink = {
  icon: React.ElementType;
  labelKey: string;
  href: string;
  exact?: boolean;
};

type NavSection = {
  labelKey?: string;
  items: NavLink[];
};

const SECTIONS: NavSection[] = [
  {
    items: [
      {
        icon: LayoutDashboard,
        labelKey: "overview",
        href: "/dashboard",
        exact: true,
      },
    ],
  },
  {
    labelKey: "sections.integration",
    items: [
      { icon: Blocks, labelKey: "connectors", href: "/dashboard/connectors" },
      { icon: Cable, labelKey: "connections", href: "/dashboard/connections" },
      {
        icon: GitBranch,
        labelKey: "syncRoutes",
        href: "/dashboard/sync-routes",
      },
    ],
  },
  {
    labelKey: "sections.orchestration",
    items: [
      { icon: PlayCircle, labelKey: "intents", href: "/dashboard/intents" },
      { icon: Activity, labelKey: "events", href: "/dashboard/events" },
    ],
  },
  {
    labelKey: "sections.workspace",
    items: [
      { icon: KeyRound, labelKey: "apiKeys", href: "/dashboard/api-keys" },
      { icon: Users, labelKey: "team", href: "/dashboard/team" },
      { icon: Settings, labelKey: "settings", href: "/dashboard/settings" },
    ],
  },
];

function isActive(pathname: string, href: string, exact?: boolean): boolean {
  if (exact) return pathname === href;
  return pathname === href || pathname.startsWith(`${href}/`);
}

export function DashboardSidebar() {
  const t = useTranslations("dashboard.nav");
  const pathname = usePathname();

  return (
    <aside className="flex w-[var(--side-w)] shrink-0 flex-col border-r border-[var(--line)] bg-[var(--white)]">
      <div className="flex h-[var(--nav-h)] items-center gap-2 border-b border-[var(--line)] px-4">
        <BarChart3 className="text-primary" size={18} />
        <span className="font-display text-sm font-semibold text-[var(--ink)]">
          Hublio
        </span>
      </div>
      <nav className="flex flex-1 flex-col gap-4 overflow-y-auto p-3">
        {SECTIONS.map((section, index) => (
          <div key={section.labelKey ?? index} className="flex flex-col gap-0.5">
            {section.labelKey ? (
              <span className="px-3 pb-1 pt-2 text-[11px] font-semibold uppercase tracking-wide text-[var(--faint)]">
                {t(section.labelKey)}
              </span>
            ) : null}
            {section.items.map((item) => {
              const Icon = item.icon;
              const active = isActive(pathname, item.href, item.exact);
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "flex items-center gap-2.5 rounded-md px-3 py-2 text-sm no-underline transition-colors",
                    active
                      ? "bg-[var(--primary-soft)] font-medium text-[var(--primary-ink)]"
                      : "text-[var(--ink-2)] hover:bg-[var(--line-2)] hover:text-[var(--ink)]",
                  )}
                >
                  <Icon size={16} />
                  {t(item.labelKey)}
                </Link>
              );
            })}
          </div>
        ))}
      </nav>
    </aside>
  );
}
