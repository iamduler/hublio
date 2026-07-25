import { getTranslations, setRequestLocale } from "next-intl/server";
import { PageHeader } from "@hublio/ui/common/page-header";
import { WorkspaceForm } from "@/features/workspaces/components/workspace-form";

export default async function NewWorkspacePage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("workspaces");

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <PageHeader title={t("createTitle")} description={t("createSubtitle")} />
      <WorkspaceForm />
    </div>
  );
}
