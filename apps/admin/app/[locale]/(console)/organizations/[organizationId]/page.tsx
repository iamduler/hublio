"use client";

import { use } from "react";
import { OrganizationDetail } from "@/features/organizations/components/organization-detail";

export default function OrganizationDetailPage({
  params,
}: {
  params: Promise<{ organizationId: string }>;
}) {
  const { organizationId } = use(params);
  return <OrganizationDetail organizationId={organizationId} />;
}
