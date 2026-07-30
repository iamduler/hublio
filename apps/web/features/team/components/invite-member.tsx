"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@hublio/ui/ui/select";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@hublio/ui/ui/form";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useAddMember } from "../hooks";
import { makeInviteSchema, type InviteValues } from "../schemas";
import type { WorkspaceRole } from "../types";

const ROLES: WorkspaceRole[] = ["owner", "admin", "member"];

export function InviteMember() {
  const t = useTranslations("team");
  const tv = useTranslations("validation");
  const getError = useApiErrorMessage();
  const addMember = useAddMember();

  const schema = useMemo(() => makeInviteSchema(tv), [tv]);
  const form = useForm<InviteValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", role: "member" },
  });

  async function onSubmit(values: InviteValues) {
    try {
      await addMember.mutateAsync(values);
      toast.success(t("invited"));
      form.reset({ email: "", role: values.role });
    } catch (err) {
      toast.error(getError(err));
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("inviteTitle")}</CardTitle>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form
            className="flex flex-col gap-4 sm:flex-row sm:items-end"
            onSubmit={form.handleSubmit(onSubmit)}
            noValidate
          >
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem className="flex-1">
                  <FormLabel>{t("email")}</FormLabel>
                  <FormControl>
                    <Input type="email" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="role"
              render={({ field }) => (
                <FormItem className="w-full sm:w-40">
                  <FormLabel>{t("role")}</FormLabel>
                  <Select value={field.value} onValueChange={field.onChange}>
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {ROLES.map((role) => (
                        <SelectItem key={role} value={role}>
                          {role}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button type="submit" disabled={form.formState.isSubmitting}>
              {t("invite")}
            </Button>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
