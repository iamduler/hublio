"use client";

import { useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Check, ChevronLeft, ChevronRight, Copy, Loader2 } from "lucide-react";
import { useRouter } from "@/i18n/navigation";
import { useAuth } from "@/providers/auth-provider";
import { useWorkspace } from "@/providers/workspace-provider";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useCreateWorkspace } from "@/features/workspaces/hooks";
import { OnboardingShell, AuthFieldLabel, cx } from "@/features/auth/auth-ui";
import { writeOnboardingDraft } from "@/lib/onboarding-state";
import { toast } from "@/lib/toast";

type FormValues = {
  name: string;
  environment: "production" | "staging" | "development";
};

function slugify(s: string) {
  return s
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 40);
}

export function WorkspaceOnboardingForm() {
  const t = useTranslations("auth.onboarding");
  const tv = useTranslations("validation");
  const router = useRouter();
  const getError = useApiErrorMessage();
  const { user } = useAuth();
  const { setActiveWorkspace } = useWorkspace();
  const createWorkspace = useCreateWorkspace(user?.organization_id);
  const [copied, setCopied] = useState(false);

  const schema = useMemo(
    () =>
      z.object({
        name: z.string().trim().min(1, tv("required")).max(255, tv("tooLong")),
        environment: z.enum(["production", "staging", "development"]),
      }),
    [tv],
  );

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", environment: "production" },
  });

  const name = form.watch("name");
  const environment = form.watch("environment");
  const slug = slugify(name) || "my-workspace";

  async function onSubmit(values: FormValues) {
    try {
      const workspace = await createWorkspace.mutateAsync({
        name: values.name,
        environment: values.environment,
      });
      setActiveWorkspace(workspace.id);
      writeOnboardingDraft({
        workspace: {
          id: workspace.id,
          name: workspace.name,
          environment: workspace.environment,
        },
      });
      router.push("/onboarding/invite");
    } catch (err) {
      toast.error(getError(err));
    }
  }

  function handleCopyId(id: string) {
    void navigator.clipboard.writeText(id).then(() => {
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2000);
    });
  }

  const envBadge =
    environment === "production"
      ? "bg-green-50 text-green-700"
      : environment === "development"
        ? "bg-blue-50 text-blue-700"
        : "bg-amber-50 text-amber-700";

  return (
    <OnboardingShell step={1}>
      <div className="mb-6">
        <h2 className="text-[18px] font-semibold tracking-tight text-(--ink)">
          {t("workspaceTitle")}
        </h2>
        <p className="mt-1.5 text-[13px] text-(--muted-clr)">
          {t("workspaceSubtitle")}
        </p>
      </div>

      <form className="space-y-4" onSubmit={form.handleSubmit(onSubmit)} noValidate>
        <div>
          <AuthFieldLabel htmlFor="ws-name" required>
            {t("workspaceName")}
          </AuthFieldLabel>
          <input
            id="ws-name"
            autoFocus
            placeholder={t("workspaceNamePlaceholder")}
            className={cx(
              "h-9 w-full rounded-lg border bg-(--white) px-3 text-sm text-(--ink) placeholder:text-(--faint) transition-colors focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-600/20",
              form.formState.errors.name
                ? "border-red-300 bg-red-50/40"
                : "border-(--line) hover:border-(--line-3)",
            )}
            {...form.register("name")}
          />
          {form.formState.errors.name ? (
            <p className="mt-1 text-[12px] text-red-600">
              {form.formState.errors.name.message}
            </p>
          ) : name ? (
            <p className="mt-1 text-[11px] text-(--faint)">
              {t("slug")}: <span className="text-(--ink-2)">{slug}</span>
            </p>
          ) : null}
        </div>

        <div>
          <AuthFieldLabel htmlFor="ws-env">{t("environment")}</AuthFieldLabel>
          <select
            id="ws-env"
            className="h-9 w-full rounded-lg border border-(--line) bg-(--white) px-3 text-sm text-(--ink) transition-colors hover:border-(--line-3) focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-600/20"
            {...form.register("environment")}
          >
            <option value="production">{t("env.production")}</option>
            <option value="development">{t("env.development")}</option>
            <option value="staging">{t("env.staging")}</option>
          </select>
        </div>

        {name ? (
          <div className="rounded-xl border border-(--line) bg-(--line-2) p-4">
            <p className="mb-2.5 text-[11px] font-semibold uppercase tracking-wider text-(--muted-clr)">
              {t("workspacePreview")}
            </p>
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-[12px] text-(--muted-clr)">{t("slug")}</span>
                <span className="text-[12px] text-(--ink-2)">{slug}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-[12px] text-(--muted-clr)">
                  {t("environment")}
                </span>
                <span
                  className={cx(
                    "rounded-md px-2 py-0.5 text-[11px] font-semibold",
                    envBadge,
                  )}
                >
                  {t(`env.${environment}`)}
                </span>
              </div>
              {createWorkspace.data?.id ? (
                <div className="flex items-center justify-between">
                  <span className="text-[12px] text-(--muted-clr)">
                    {t("workspaceId")}
                  </span>
                  <div className="flex items-center gap-1.5">
                    <span className="text-[12px] text-(--ink-2)">
                      {createWorkspace.data.id.slice(0, 13)}…
                    </span>
                    <button
                      type="button"
                      onClick={() => handleCopyId(createWorkspace.data!.id)}
                      className="text-(--faint) hover:text-(--ink-2)"
                    >
                      {copied ? (
                        <Check size={12} className="text-green-600" />
                      ) : (
                        <Copy size={12} />
                      )}
                    </button>
                  </div>
                </div>
              ) : null}
            </div>
          </div>
        ) : null}

        <div className="mt-7 flex items-center justify-between">
          <button
            type="button"
            onClick={() => router.push("/onboarding/organization")}
            className="flex items-center gap-1.5 text-[13px] text-(--muted-clr) transition-colors hover:text-(--ink-2)"
          >
            <ChevronLeft size={14} />
            {t("back")}
          </button>
          <button
            type="submit"
            disabled={form.formState.isSubmitting || createWorkspace.isPending}
            className="flex h-9 items-center gap-2 rounded-lg bg-blue-600 px-5 text-[13px] font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
          >
            {form.formState.isSubmitting || createWorkspace.isPending ? (
              <Loader2 size={13} className="animate-spin" />
            ) : null}
            {t("continue")}
            <ChevronRight size={14} />
          </button>
        </div>
      </form>
    </OnboardingShell>
  );
}
