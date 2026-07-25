"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useSearchParams } from "next/navigation";
import { Link, useRouter } from "@/i18n/navigation";
import { authApi } from "@/lib/api/auth";
import { useAuth } from "@/providers/auth-provider";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { makeLoginSchema, type LoginValues } from "@/features/auth/schemas";
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

export function LoginForm() {
  const t = useTranslations("auth.login");
  const tv = useTranslations("validation");
  const { establishSession } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const getError = useApiErrorMessage();

  const schema = useMemo(() => makeLoginSchema(tv), [tv]);
  const form = useForm<LoginValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", password: "" },
  });

  async function onSubmit(values: LoginValues) {
    try {
      const data = await authApi.login(values.email, values.password);
      await establishSession(data);
      const callback = searchParams.get("callbackUrl") || "/dashboard";
      router.replace(callback);
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
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("email")}</FormLabel>
                  <FormControl>
                    <Input
                      type="email"
                      autoComplete="email"
                      {...field}
                    />
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
                      autoComplete="current-password"
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
          {t("noAccount")}{" "}
          <Link
            href="/register"
            className="text-primary no-underline hover:underline"
          >
            {t("registerLink")}
          </Link>
        </p>
      </CardContent>
    </Card>
  );
}
