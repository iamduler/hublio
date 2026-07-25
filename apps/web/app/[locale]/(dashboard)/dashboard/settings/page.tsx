import { getTranslations, setRequestLocale } from "next-intl/server";
import { PageHeader } from "@hublio/ui/common/page-header";
import { WorkspaceSettings } from "@/features/workspaces/components/workspace-settings";

export default async function SettingsPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("workspaces");

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <PageHeader title={t("settings.title")} description={t("settings.subtitle")} />
      <WorkspaceSettings />
    </div>
  );
}
