"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { ChevronRight, Loader2 } from "lucide-react";
import { useRouter } from "@/i18n/navigation";
import { authApi } from "@/lib/api/auth";
import { useAuth } from "@/providers/auth-provider";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { OnboardingShell, cx } from "@/features/auth/auth-ui";
import { writeOnboardingDraft } from "@/lib/onboarding-state";
import { toast } from "@/lib/toast";

type FormValues = {
  organizationName: string;
};

export function OrganizationOnboardingForm() {
  const t = useTranslations("auth.onboarding");
  const tv = useTranslations("validation");
  const router = useRouter();
  const { establishSession } = useAuth();
  const getError = useApiErrorMessage();
  const [email, setEmail] = useState<string>("");
  const [loadingPreview, setLoadingPreview] = useState(true);

  const schema = useMemo(
    () =>
      z.object({
        organizationName: z
          .string()
          .trim()
          .min(1, tv("required"))
          .max(255, tv("tooLong")),
      }),
    [tv],
  );

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { organizationName: "" },
  });

  useEffect(() => {
    let cancelled = false;
    authApi
      .oauthOnboardingPreview()
      .then((preview) => {
        if (!cancelled) {
          setEmail(preview.email);
          setLoadingPreview(false);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          toast.error(getError(err));
          router.replace("/login?oauth_error=" + encodeURIComponent(t("expired")));
        }
      });
    return () => {
      cancelled = true;
    };
  }, [getError, router, t]);

  async function onSubmit(values: FormValues) {
    try {
      const data = await authApi.completeOAuthRegistration({
        organization_name: values.organizationName,
      });
      await establishSession(data);
      writeOnboardingDraft({
        organizationName: values.organizationName,
        workspace: undefined,
        invitedEmails: [],
        pendingInviteEmails: [],
      });
      router.replace("/onboarding/workspace");
    } catch (err) {
      toast.error(getError(err));
    }
  }

  if (loadingPreview) {
    return (
      <OnboardingShell step={0}>
        <div className="flex min-h-[160px] items-center justify-center text-(--muted-clr)">
          <Loader2 className="animate-spin" size={18} />
        </div>
      </OnboardingShell>
    );
  }

  return (
    <OnboardingShell step={0}>
      <div className="mb-6">
        <h2 className="text-[18px] font-semibold tracking-tight text-(--ink)">
          {t("title")}
        </h2>
        <p className="mt-1.5 text-[13px] text-(--muted-clr)">{t("subtitle")}</p>
      </div>

      <form className="space-y-4" onSubmit={form.handleSubmit(onSubmit)} noValidate>
        <div>
          <label className="mb-1.5 block text-[13px] font-medium text-(--ink-2)">
            {t("email")}
          </label>
          <input
            type="email"
            value={email}
            readOnly
            className="h-9 w-full rounded-lg border border-(--line) bg-(--line-2) px-3 text-sm text-(--ink-2)"
          />
        </div>

        <div>
          <label
            htmlFor="organizationName"
            className="mb-1.5 block text-[13px] font-medium text-(--ink-2)"
          >
            {t("organizationName")} <span className="text-red-500">*</span>
          </label>
          <input
            id="organizationName"
            autoFocus
            placeholder={t("organizationPlaceholder")}
            className={cx(
              "h-9 w-full rounded-lg border bg-(--white) px-3 text-sm text-(--ink) placeholder:text-(--faint) transition-colors focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-600/20",
              form.formState.errors.organizationName
                ? "border-red-300 bg-red-50/40"
                : "border-(--line) hover:border-(--line-3)",
            )}
            {...form.register("organizationName")}
          />
          {form.formState.errors.organizationName && (
            <p className="mt-1 text-[12px] text-red-600">
              {form.formState.errors.organizationName.message}
            </p>
          )}
        </div>

        <div className="mt-7 flex justify-end">
          <button
            type="submit"
            disabled={form.formState.isSubmitting}
            className="flex h-9 items-center gap-2 rounded-lg bg-blue-600 px-5 text-[13px] font-medium text-white transition-colors hover:bg-blue-700 disabled:opacity-50"
          >
            {form.formState.isSubmitting ? (
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
