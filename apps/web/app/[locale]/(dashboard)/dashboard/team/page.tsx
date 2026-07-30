import { getTranslations, setRequestLocale } from "next-intl/server";
import { PageHeader } from "@hublio/ui/common/page-header";
import { InviteMember } from "@/features/team/components/invite-member";
import { MembersList } from "@/features/team/components/members-list";

export default async function TeamPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("team");

  return (
    <div className="space-y-6">
      <PageHeader title={t("title")} description={t("subtitle")} />
      <MembersList />
      <InviteMember />
    </div>
  );
}
