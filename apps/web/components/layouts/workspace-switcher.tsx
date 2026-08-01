"use client";

import { useTranslations } from "next-intl";
import { Check, ChevronsUpDown, Plus } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@hublio/ui/ui/dropdown-menu";
import { EnvBadge } from "@hublio/ui/common/env-badge";
import { useWorkspace } from "@/providers/workspace-provider";
import { useRouter } from "@/i18n/navigation";
import { cn } from "@/lib/utils";

export function WorkspaceSwitcher() {
  const t = useTranslations("workspaces");
  const router = useRouter();
  const { workspaces, activeWorkspace, isLoading, setActiveWorkspace } =
    useWorkspace();

  if (isLoading) {
    return (
      <div className="h-9 w-40 animate-pulse rounded-md bg-(--line-2)" />
    );
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        className="flex h-9 items-center gap-2 rounded-md border border-(--line) bg-(--white) px-3 text-sm outline-none hover:border-(--line-3) focus:border-primary"
        aria-label={t("switcher.label")}
      >
        <span className="max-w-[10rem] truncate font-medium text-(--ink)">
          {activeWorkspace?.name ?? t("switcher.none")}
        </span>
        {activeWorkspace ? (
          <EnvBadge environment={activeWorkspace.environment} />
        ) : null}
        <ChevronsUpDown size={14} className="text-(--muted-clr)" />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" className="min-w-[16rem]">
        <DropdownMenuLabel>{t("switcher.label")}</DropdownMenuLabel>
        {workspaces.length === 0 ? (
          <div className="px-3 py-2 text-sm text-(--muted-clr)">
            {t("switcher.empty")}
          </div>
        ) : (
          workspaces.map((workspace) => (
            <DropdownMenuItem
              key={workspace.id}
              onSelect={() => setActiveWorkspace(workspace.id)}
              className="justify-between"
            >
              <span className="flex items-center gap-2">
                <Check
                  size={14}
                  className={cn(
                    activeWorkspace?.id === workspace.id
                      ? "opacity-100 text-primary"
                      : "opacity-0",
                  )}
                />
                <span className="truncate">{workspace.name}</span>
              </span>
              <EnvBadge environment={workspace.environment} />
            </DropdownMenuItem>
          ))
        )}
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onSelect={() => router.push("/dashboard/workspaces/new")}
        >
          <Plus size={14} />
          {t("switcher.create")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
