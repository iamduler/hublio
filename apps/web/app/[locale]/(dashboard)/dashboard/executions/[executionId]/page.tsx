import { setRequestLocale } from "next-intl/server";
import { ExecutionDetail } from "@/features/executions/components/execution-detail";

export default async function ExecutionDetailPage({
  params,
}: {
  params: Promise<{ locale: string; executionId: string }>;
}) {
  const { locale, executionId } = await params;
  setRequestLocale(locale);

  return <ExecutionDetail executionId={executionId} />;
}
