"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "@/i18n/navigation";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Textarea } from "@hublio/ui/ui/textarea";
import { Card, CardContent } from "@hublio/ui/ui/card";
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
import { useConnectors } from "@/features/connectors/hooks";
import { useCreateConnection } from "../hooks";
import {
  SECRET_FIELDS,
  makeConnectionSchema,
  type ConnectionFormValues,
} from "../schemas";
import type { CredentialType } from "../types";

const CREDENTIAL_TYPES: CredentialType[] = [
  "api_key",
  "bearer_token",
  "basic_auth",
  "oauth2",
  "jwt",
];

const ENVIRONMENTS = ["production", "staging", "development"];

export function ConnectionForm() {
  const t = useTranslations("connections");
  const tv = useTranslations("validation");
  const router = useRouter();
  const getError = useApiErrorMessage();
  const { data: connectors } = useConnectors();
  const createConnection = useCreateConnection();

  const schema = useMemo(() => makeConnectionSchema(tv), [tv]);
  const form = useForm<ConnectionFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      connector_id: "",
      name: "",
      environment: "production",
      credential_type: "api_key",
      description: "",
      secret: {},
    },
  });

  const credentialType = form.watch("credential_type");
  const secretFields = SECRET_FIELDS[credentialType] ?? [];

  async function onSubmit(values: ConnectionFormValues) {
    try {
      const connection = await createConnection.mutateAsync({
        connector_id: values.connector_id,
        name: values.name,
        environment: values.environment,
        credential_type: values.credential_type,
        secret: values.secret,
        description: values.description || undefined,
      });
      toast.success(t("created"));
      router.replace(`/dashboard/connections/${connection.id}`);
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
            onSubmit={form.handleSubmit(onSubmit)}
            noValidate
          >
            <FormField
              control={form.control}
              name="connector_id"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("form.connector")}</FormLabel>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder={t("form.selectConnector")} />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {(connectors ?? []).map((connector) => (
                        <SelectItem key={connector.id} value={connector.id}>
                          {connector.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("form.name")}</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid gap-5 sm:grid-cols-2">
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

              <FormField
                control={form.control}
                name="credential_type"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("form.credentialType")}</FormLabel>
                    <Select
                      value={field.value}
                      onValueChange={(v) => {
                        field.onChange(v);
                        form.setValue("secret", {});
                      }}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {CREDENTIAL_TYPES.map((type) => (
                          <SelectItem key={type} value={type}>
                            {type}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="space-y-3 rounded-md border border-(--line) bg-(--line-2) p-4">
              <p className="text-sm font-medium text-(--ink)">
                {t("form.secret")}
              </p>
              {secretFields.map((secretField) => (
                <div key={secretField} className="space-y-2">
                  <FormLabel htmlFor={`secret-${secretField}`}>
                    {secretField}
                  </FormLabel>
                  <Input
                    id={`secret-${secretField}`}
                    type={
                      secretField.includes("password") ||
                        secretField.includes("secret") ||
                        secretField.includes("key") ||
                        secretField.includes("token")
                        ? "password"
                        : "text"
                    }
                    autoComplete="off"
                    {...form.register(`secret.${secretField}`)}
                  />
                </div>
              ))}
            </div>

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("form.description")}</FormLabel>
                  <FormControl>
                    <Textarea rows={2} {...field} />
                  </FormControl>
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
              <Button
                type="submit"
                disabled={form.formState.isSubmitting}
              >
                {t("form.submit")}
              </Button>
            </div>
          </form>
        </Form>
      </CardContent>
    </Card>
  );
}
