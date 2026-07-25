"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Link, useRouter } from "@/i18n/navigation";
import { authApi } from "@/lib/api/auth";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import {
  makeRegisterSchema,
  type RegisterValues,
} from "@/features/auth/schemas";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@hublio/ui/ui/form";
import { Card, CardContent, CardHeader, CardTitle } from "@hublio/ui/ui/card";
import { toast } from "@/lib/toast";

export default function RegisterPage() {
  const t = useTranslations("auth.register");
  const tv = useTranslations("validation");
  const router = useRouter();
  const getError = useApiErrorMessage();

  const schema = useMemo(() => makeRegisterSchema(tv), [tv]);
  const form = useForm<RegisterValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      organization_name: "",
      full_name: "",
      email: "",
      password: "",
    },
  });

  async function onSubmit(values: RegisterValues) {
    try {
      await authApi.register(values);
      toast.success(t("success"));
      router.replace("/login");
    } catch (err) {
      toast.error(getError(err));
    }
  }

  return (
    <Card className="w-full max-w-md shadow-md">
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
        <p className="text-sm text-[var(--muted-clr)]">{t("subtitle")}</p>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form
            className="space-y-4"
            onSubmit={form.handleSubmit(onSubmit)}
            noValidate
          >
            <FormField
              control={form.control}
              name="organization_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("organizationName")}</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="full_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("fullName")}</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("email")}</FormLabel>
                  <FormControl>
                    <Input type="email" autoComplete="email" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("password")}</FormLabel>
                  <FormControl>
                    <Input
                      type="password"
                      autoComplete="new-password"
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button
              type="submit"
              className="w-full"
              disabled={form.formState.isSubmitting}
            >
              {t("submit")}
            </Button>
          </form>
        </Form>
        <p className="mt-4 text-center text-sm text-[var(--ink-2)]">
          {t("hasAccount")}{" "}
          <Link
            href="/login"
            className="text-primary no-underline hover:underline"
          >
            {t("loginLink")}
          </Link>
        </p>
      </CardContent>
    </Card>
  );
}
