"use client";

import { useTranslations } from "next-intl";
import { Building2 } from "lucide-react";
import { EmptyState } from "@hublio/ui/ui/empty-state";

export default function OrganizationsStubPage() {
  const t = useTranslations("organizations");
  return (
    <EmptyState icon={Building2} title={t("stubTitle")} description={t("stubDescription")} />
  );
}
