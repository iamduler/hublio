"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Card, CardContent } from "@hublio/ui/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@hublio/ui/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@hublio/ui/ui/select";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useAuth } from "@/providers/auth-provider";
import { useWorkspace } from "@/providers/workspace-provider";
import { useRouter } from "@/i18n/navigation";
import { useCreateWorkspace } from "../hooks";
import {
  makeWorkspaceSchema,
  type WorkspaceFormValues,
} from "../schemas";

const ENVIRONMENTS = ["production", "staging", "development"] as const;

export function WorkspaceForm() {
  const t = useTranslations("workspaces");
  const tv = useTranslations("validation");
  const getError = useApiErrorMessage();
  const router = useRouter();
  const { user } = useAuth();
  const { setActiveWorkspace } = useWorkspace();
  const createWorkspace = useCreateWorkspace(user?.organization_id);

  const schema = useMemo(() => makeWorkspaceSchema(tv), [tv]);
  const form = useForm<WorkspaceFormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", environment: "production" },
  });

  async function onSubmit(values: WorkspaceFormValues) {
    try {
      const workspace = await createWorkspace.mutateAsync(values);
      setActiveWorkspace(workspace.id);
      toast.success(t("created"));
      router.replace("/dashboard");
    } catch (err) {
      toast.error(getError(err));
    }
  }

  return (
    <Card>
      <CardContent className="p-6">
        <Form {...form}>
          <form
            className="space-y-5"
            onSubmit={form.handleSubmit((v) => void onSubmit(v))}
          >
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("form.name")}</FormLabel>
                  <FormControl>
                    <Input id="ws-name" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="environment"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("form.environment")}</FormLabel>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {ENVIRONMENTS.map((env) => (
                        <SelectItem key={env} value={env}>
                          {env}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => router.back()}
              >
                {t("form.cancel")}
              </Button>
              <Button type="submit" disabled={createWorkspace.isPending}>
                {t("form.submit")}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
