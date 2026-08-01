"use client";

import { useTranslations } from "next-intl";
import { Blocks } from "lucide-react";
import { EmptyState } from "@hublio/ui/ui/empty-state";

export default function ConnectorsStubPage() {
  const t = useTranslations("connectors");
  return (
    <EmptyState icon={Blocks} title={t("stubTitle")} description={t("stubDescription")} />
  );
}
