"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Link, useRouter } from "@/i18n/navigation";
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
import { Breadcrumb } from "@hublio/ui/common/breadcrumb";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useCreateOrganization } from "../hooks";
import {
  makeCreateOrganizationSchema,
  type CreateOrganizationValues,
} from "../schemas";

export function OrganizationCreateForm() {
  const t = useTranslations("organizations");
  const tv = useTranslations("validation");
  const getError = useApiErrorMessage();
  const router = useRouter();
  const create = useCreateOrganization();

  const schema = useMemo(() => makeCreateOrganizationSchema(tv), [tv]);
  const form = useForm<CreateOrganizationValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "" },
  });

  async function onSubmit(values: CreateOrganizationValues) {
    try {
      const result = await create.mutateAsync({ name: values.name });
      toast.success(t("create.done"));
      router.replace(`/organizations/${result.organization.id}`);
    } catch (err) {
      toast.error(getError(err));
    }
  }

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <div>
        <Breadcrumb
          items={[
            { label: t("breadcrumb.admin"), href: "/" },
            {
              label: t("breadcrumb.organizations"),
              href: "/organizations",
            },
            { label: t("create.title") },
          ]}
          renderLink={({ href, className, children }) => (
            <Link href={href} className={className}>
              {children}
            </Link>
          )}
        />
        <h1 className="mt-2 font-display text-xl font-semibold tracking-tight text-(--ink)">
          {t("create.title")}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {t("create.description")}
        </p>
      </div>

      <Card>
        <CardContent className="py-6">
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
                    <FormLabel>{t("create.form.name")}</FormLabel>
                    <FormControl>
                      <Input
                        autoFocus
                        placeholder={t("create.form.namePlaceholder")}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <p className="text-xs text-muted-foreground">
                {t("create.hint")}
              </p>
              <div className="flex flex-wrap gap-2">
                <Button type="submit" disabled={create.isPending}>
                  {t("create.submit")}
                </Button>
                <Button type="button" variant="outline" asChild>
                  <Link href="/organizations">{t("actions.cancel")}</Link>
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}
