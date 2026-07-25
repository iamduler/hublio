import { getTranslations, setRequestLocale } from "next-intl/server";
import { MailCheck } from "lucide-react";
import { Link } from "@/i18n/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { EmptyState } from "@hublio/ui/ui/empty-state";

export default async function VerifyEmailPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("auth.verify");

  return (
    <Card className="w-full max-w-md shadow-md">
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
        <p className="text-sm text-[var(--muted-clr)]">{t("subtitle")}</p>
      </CardHeader>
      <CardContent>
        <EmptyState
          icon={MailCheck}
          title={t("unavailableTitle")}
          description={t("unavailableBody")}
          size="sm"
        />
        <p className="mt-4 text-center text-sm text-[var(--ink-2)]">
          <Link
            href="/login"
            className="text-primary no-underline hover:underline"
          >
            {t("backToLogin")}
          </Link>
        </p>
      </CardContent>
    </Card>
  );
}
