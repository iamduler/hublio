"use client";

import { useTranslations } from "next-intl";
import { Blocks, Building2, LayoutDashboard } from "lucide-react";
import {
  AppSidebar,
  type AppSidebarSection,
} from "@hublio/ui/common/app-sidebar";
import { Link, usePathname } from "@/i18n/navigation";
import { Logo } from "@/components/logo";

type NavLink = {
  icon: React.ElementType;
  labelKey: string;
  href: string;
  exact?: boolean;
};

const NAV: NavLink[] = [
  { icon: LayoutDashboard, labelKey: "overview", href: "/", exact: true },
  { icon: Building2, labelKey: "organizations", href: "/organizations" },
  { icon: Blocks, labelKey: "connectors", href: "/connectors" },
];

function isActive(pathname: string, href: string, exact?: boolean): boolean {
  if (exact) return pathname === href;
  return pathname === href || pathname.startsWith(`${href}/`);
}

export function AdminSidebar() {
  const t = useTranslations("shell.nav");
  const pathname = usePathname();

  const sections: AppSidebarSection[] = [
    {
      items: NAV.map((item) => ({
        href: item.href,
        label: t(item.labelKey),
        icon: item.icon,
        active: isActive(pathname, item.href, item.exact),
      })),
    },
  ];

  return (
    <AppSidebar
      logo={
        <Link href="/" className="no-underline">
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
