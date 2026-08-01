"use client";

import * as React from "react";
import { Menu, PanelLeft, PanelLeftClose } from "lucide-react";
import { cn } from "../lib/utils";
import { Button } from "../ui/button";
import {
  Sheet,
  SheetContent,
  SheetTitle,
} from "../ui/sheet";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "../ui/tooltip";

const STORAGE_KEY = "hublio.sidebar.collapsed";

export type AppSidebarLinkProps = {
  href: string;
  className?: string;
  children: React.ReactNode;
  onClick?: () => void;
};

export type AppSidebarItem = {
  href: string;
  label: string;
  icon: React.ElementType;
  active?: boolean;
};

export type AppSidebarSection = {
  label?: string;
  items: AppSidebarItem[];
};

type AppSidebarContextValue = {
  collapsed: boolean;
  mobileOpen: boolean;
  toggleCollapsed: () => void;
  setMobileOpen: (open: boolean) => void;
};

const AppSidebarContext = React.createContext<AppSidebarContextValue | null>(
  null,
);

export function useAppSidebar(): AppSidebarContextValue {
  const ctx = React.useContext(AppSidebarContext);
  if (!ctx) {
    throw new Error("useAppSidebar must be used within AppSidebarProvider");
  }
  return ctx;
}

export function AppSidebarProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const [collapsed, setCollapsed] = React.useState(false);
  const [mobileOpen, setMobileOpen] = React.useState(false);
  const [hydrated, setHydrated] = React.useState(false);

  React.useEffect(() => {
    try {
      setCollapsed(localStorage.getItem(STORAGE_KEY) === "1");
    } catch {
      /* ignore */
    }
    setHydrated(true);
  }, []);

  React.useEffect(() => {
    if (!hydrated) return;
    try {
      localStorage.setItem(STORAGE_KEY, collapsed ? "1" : "0");
    } catch {
      /* ignore */
    }
  }, [collapsed, hydrated]);

  const toggleCollapsed = React.useCallback(() => {
    setCollapsed((prev) => !prev);
  }, []);

  const value = React.useMemo(
    () => ({ collapsed, mobileOpen, toggleCollapsed, setMobileOpen }),
    [collapsed, mobileOpen, toggleCollapsed],
  );

  return (
    <AppSidebarContext.Provider value={value}>
      {children}
    </AppSidebarContext.Provider>
  );
}

export type AppSidebarProps = {
  logo: React.ReactNode;
  /** Shown in the collapsed desktop rail; falls back to `logo`. */
  logoCollapsed?: React.ReactNode;
  sections: AppSidebarSection[];
  renderLink: (props: AppSidebarLinkProps) => React.ReactNode;
  collapseLabel: string;
  expandLabel: string;
  /** Accessible title for the mobile sheet. */
  mobileTitle?: string;
};

function navItemClass(active: boolean, collapsed: boolean) {
  return cn(
    "flex items-center rounded-md text-sm no-underline transition-colors",
    collapsed ? "justify-center px-0 py-2" : "gap-2.5 px-3 py-2",
    active
      ? "bg-(--line-2) font-medium text-(--ink)"
      : "text-(--ink-2) hover:bg-(--line-2) hover:text-(--ink)",
  );
}

function SidebarChrome({
  logo,
  logoCollapsed,
  sections,
  renderLink,
  collapseLabel,
  expandLabel,
  collapsed,
  showCollapseToggle,
  onNavigate,
}: {
  logo: React.ReactNode;
  logoCollapsed?: React.ReactNode;
  sections: AppSidebarSection[];
  renderLink: (props: AppSidebarLinkProps) => React.ReactNode;
  collapseLabel: string;
  expandLabel: string;
  collapsed: boolean;
  showCollapseToggle: boolean;
  onNavigate?: () => void;
}) {
  const { toggleCollapsed } = useAppSidebar();

  return (
    <>
      <div
        className={cn(
          "flex h-(--nav-h) shrink-0 items-center border-b border-(--line)",
          collapsed ? "justify-center px-1" : "justify-between gap-1 px-3",
        )}
      >
        {collapsed ? (
          showCollapseToggle ? (
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              onClick={toggleCollapsed}
              aria-label={expandLabel}
            >
              {logoCollapsed ?? <PanelLeft className="size-4" />}
            </Button>
          ) : (
            (logoCollapsed ?? logo)
          )
        ) : (
          <>
            <div className="min-w-0">{logo}</div>
            {showCollapseToggle ? (
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="shrink-0"
                onClick={toggleCollapsed}
                aria-label={collapseLabel}
              >
                <PanelLeftClose className="size-4" />
              </Button>
            ) : null}
          </>
        )}
      </div>

      <nav
        className={cn(
          "flex flex-1 flex-col overflow-y-auto p-3",
          collapsed ? "gap-1" : "gap-4",
        )}
      >
        {sections.map((section, index) => (
          <div
            key={section.label ?? `section-${index}`}
            className="flex flex-col gap-0.5"
          >
            {section.label && !collapsed ? (
              <span className="px-3 pb-1 pt-2 text-[11px] font-semibold uppercase tracking-wide text-(--faint)">
                {section.label}
              </span>
            ) : null}
            {section.items.map((item) => {
              const Icon = item.icon;
              const link = renderLink({
                href: item.href,
                className: navItemClass(Boolean(item.active), collapsed),
                onClick: onNavigate,
                children: (
                  <>
                    <Icon className="size-4 shrink-0" aria-hidden />
                    {!collapsed ? <span className="truncate">{item.label}</span> : null}
                    {collapsed ? (
                      <span className="sr-only">{item.label}</span>
                    ) : null}
                  </>
                ),
              });

              if (!collapsed) return <React.Fragment key={item.href}>{link}</React.Fragment>;

              return (
                <Tooltip key={item.href}>
                  <TooltipTrigger asChild>
                    <span className="inline-flex w-full justify-center">
                      {link}
                    </span>
                  </TooltipTrigger>
                  <TooltipContent side="right" sideOffset={8}>
                    {item.label}
                  </TooltipContent>
                </Tooltip>
              );
            })}
          </div>
        ))}
      </nav>
    </>
  );
}

export function AppSidebar({
  logo,
  logoCollapsed,
  sections,
  renderLink,
  collapseLabel,
  expandLabel,
  mobileTitle = "Navigation",
}: AppSidebarProps) {
  const { collapsed, mobileOpen, setMobileOpen } = useAppSidebar();

  const chromeProps = {
    logo,
    logoCollapsed,
    sections,
    renderLink,
    collapseLabel,
    expandLabel,
  };

  return (
    <TooltipProvider delayDuration={0}>
      {/* Desktop rail */}
      <aside
        className={cn(
          "sticky top-0 hidden h-dvh shrink-0 flex-col border-r border-(--line) bg-(--white) transition-[width] duration-200 md:flex",
          collapsed ? "w-14" : "w-56",
        )}
      >
        <SidebarChrome
          {...chromeProps}
          collapsed={collapsed}
          showCollapseToggle
        />
      </aside>

      {/* Mobile drawer */}
      <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
        <SheetContent
          side="left"
          showCloseButton={false}
          className="w-56 max-w-[85vw] gap-0 bg-(--white) p-0"
        >
          <SheetTitle className="sr-only">{mobileTitle}</SheetTitle>
          <div className="flex h-full flex-col">
            <SidebarChrome
              {...chromeProps}
              collapsed={false}
              showCollapseToggle={false}
              onNavigate={() => setMobileOpen(false)}
            />
          </div>
        </SheetContent>
      </Sheet>
    </TooltipProvider>
  );
}

export { isActivePath } from "../lib/nav";

export type AppSidebarTriggerProps = {
  menuLabel: string;
  className?: string;
};

export function AppSidebarTrigger({
  menuLabel,
  className,
}: AppSidebarTriggerProps) {
  const { setMobileOpen } = useAppSidebar();

  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-sm"
      className={cn("md:hidden", className)}
      onClick={() => setMobileOpen(true)}
      aria-label={menuLabel}
    >
      <Menu className="size-4" />
    </Button>
  );
}
