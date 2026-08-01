"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@hublio/ui/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { useAuth } from "@/providers/auth-provider";
import { useOrg } from "@/providers/org-provider";
import { Logo } from "@/components/logo";

export default function AdminOverviewPage() {
  const t = useTranslations("admin");
  const { user } = useAuth();
  const { organization } = useOrg();

  return (
    <div className="mx-auto flex max-w-3xl flex-col gap-6">
      <div className="flex flex-col gap-2">
        <span className="eyebrow">{t("eyebrow")}</span>
        <div className="flex flex-wrap items-center gap-3">
          <Logo size="md" />
          <h1 className="text-2xl font-semibold tracking-tight">{t("title")}</h1>
          <Badge variant="outline">{t("status")}</Badge>
        </div>
        <p className="text-muted-foreground">{t("subtitle")}</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("overview.session")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <p>
            {t("overview.signedInAs")}{" "}
            <span className="font-mono text-(--ink)">{user?.email}</span>
          </p>
          {organization ? (
            <p>
              {t("overview.orgContext")}{" "}
              <span className="font-medium text-(--ink)">
                {organization.name}
              </span>{" "}
              ({organization.status})
            </p>
          ) : null}
          <p>{t("placeholder")}</p>
        </CardContent>
      </Card>
    </div>
  );
}
