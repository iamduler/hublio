"use client";

import { useTranslations } from "next-intl";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@hublio/ui/ui/tabs";
import { WorkspaceSettings } from "@/features/workspaces/components/workspace-settings";
import { MFASettingsPanel } from "@/features/settings/components/mfa-settings-panel";

export function SettingsTabs() {
  const t = useTranslations("workspaces.settings");

  return (
    <Tabs defaultValue="general" className="space-y-4">
      <TabsList>
        <TabsTrigger value="general">{t("tabGeneral")}</TabsTrigger>
        <TabsTrigger value="security">{t("tabSecurity")}</TabsTrigger>
      </TabsList>
      <TabsContent value="general">
        <WorkspaceSettings />
      </TabsContent>
      <TabsContent value="security">
        <MFASettingsPanel />
      </TabsContent>
    </Tabs>
  );
}
