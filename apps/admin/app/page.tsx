"use client";

import { useTranslations } from "next-intl";
import { ShieldCheck } from "lucide-react";
import { Badge } from "@hublio/ui/ui/badge";
import { Button } from "@hublio/ui/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";

export default function AdminHomePage() {
  const t = useTranslations("admin");

  return (
    <main className="mx-auto flex min-h-screen max-w-3xl flex-col justify-center gap-6 px-6 py-16">
      <div className="flex flex-col gap-2">
        <span className="eyebrow">{t("eyebrow")}</span>
        <div className="flex items-center gap-3">
          <span className="grid size-11 place-items-center rounded-xl bg-accent text-accent-foreground">
            <ShieldCheck className="size-5" />
          </span>
          <h1 className="text-2xl font-semibold tracking-tight">{t("title")}</h1>
          <Badge variant="outline">{t("status")}</Badge>
        </div>
        <p className="text-muted-foreground">{t("subtitle")}</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <p className="text-sm text-muted-foreground">{t("placeholder")}</p>
          <div>
            <Button asChild>
              <a href="/docs">{t("cta")}</a>
            </Button>
          </div>
        </CardContent>
      </Card>
    </main>
  );
}
