"use client";

import { useTranslations } from "next-intl";
import { ShieldAlert } from "lucide-react";
import { Button } from "@hublio/ui/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { useAuth } from "@/providers/auth-provider";
import { useRouter } from "@/i18n/navigation";

export default function ForbiddenPage() {
  const t = useTranslations("auth.forbidden");
  const { logout, isAuthenticated } = useAuth();
  const router = useRouter();

  async function onSignOut() {
    await logout();
    router.replace("/login");
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="space-y-3 text-center">
        <div className="mx-auto grid size-11 place-items-center rounded-xl bg-destructive/10 text-destructive">
          <ShieldAlert className="size-5" />
        </div>
        <CardTitle className="text-lg">{t("title")}</CardTitle>
        <p className="text-sm text-muted-foreground">{t("subtitle")}</p>
      </CardHeader>
      <CardContent className="flex flex-col gap-2">
        {isAuthenticated ? (
          <Button type="button" onClick={() => void onSignOut()}>
            {t("signOut")}
          </Button>
        ) : (
          <Button type="button" onClick={() => router.replace("/login")}>
            {t("goLogin")}
          </Button>
        )}
      </CardContent>
    </Card>
  );
}
