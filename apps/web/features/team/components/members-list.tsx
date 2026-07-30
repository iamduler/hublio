"use client";

import { useTranslations } from "next-intl";
import { Users } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@hublio/ui/ui/table";
import { Card } from "@hublio/ui/ui/card";
import { EmptyState } from "@hublio/ui/ui/empty-state";
import { ErrorState } from "@hublio/ui/ui/error-state";
import { LoadingState } from "@hublio/ui/ui/loading-state";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useMembers } from "../hooks";

export function MembersList() {
  const t = useTranslations("team");
  const getError = useApiErrorMessage();
  const { data, isLoading, isError, error, refetch } = useMembers();

  if (isLoading) return <LoadingState rows={4} />;
  if (isError) {
    return (
      <ErrorState
        title={t("loadError")}
        description={getError(error)}
        onRetry={() => void refetch()}
      />
    );
  }
  if (!data || data.length === 0) {
    return (
      <EmptyState
        icon={Users}
        title={t("empty")}
        description={t("emptyBody")}
      />
    );
  }

  return (
    <Card className="overflow-hidden">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("columns.name")}</TableHead>
            <TableHead>{t("columns.email")}</TableHead>
            <TableHead>{t("columns.role")}</TableHead>
            <TableHead>{t("columns.joined")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((member) => (
            <TableRow key={member.user_id}>
              <TableCell className="font-medium text-[var(--ink)]">
                {member.full_name || "—"}
              </TableCell>
              <TableCell className="text-[var(--ink-2)]">{member.email}</TableCell>
              <TableCell className="capitalize text-[var(--ink-2)]">
                {member.role}
              </TableCell>
              <TableCell className="text-[var(--muted-clr)]">
                {member.created_at
                  ? new Date(member.created_at).toLocaleString()
                  : "—"}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Card>
  );
}
