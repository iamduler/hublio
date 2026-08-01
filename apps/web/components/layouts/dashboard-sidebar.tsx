"use client";

import { useTranslations } from "next-intl";
import {
  Activity,
  Blocks,
  Cable,
  GitBranch,
  KeyRound,
  LayoutDashboard,
  ListOrdered,
  PlayCircle,
  Settings,
  Users,
} from "lucide-react";
import {
  AppSidebar,
  type AppSidebarSection,
} from "@hublio/ui/common/app-sidebar";
import { Link, usePathname } from "@/i18n/navigation";
import { Logo } from "@/features/auth/auth-ui";

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
      {
        icon: ListOrdered,
        labelKey: "executions",
        href: "/dashboard/executions",
      },
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

  const sections: AppSidebarSection[] = SECTIONS.map((section) => ({
    label: section.labelKey ? t(section.labelKey) : undefined,
    items: section.items.map((item) => ({
      href: item.href,
      label: t(item.labelKey),
      icon: item.icon,
      active: isActive(pathname, item.href, item.exact),
    })),
  }));

  return (
    <AppSidebar
      logo={
        <Link href="/dashboard" className="no-underline">
          <Logo size="sm" />
        </Link>
      }
      sections={sections}
      renderLink={({ href, className, children, onClick }) => (
        <Link href={href} className={className} onClick={onClick}>
          {children}
        </Link>
      )}
      collapseLabel={t("collapse")}
      expandLabel={t("expand")}
      mobileTitle={t("menu")}
    />
  );
}
