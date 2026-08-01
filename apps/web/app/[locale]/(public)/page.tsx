import { getTranslations, setRequestLocale } from "next-intl/server";
import { Link } from "@/i18n/navigation";
import { buttonVariants } from "@hublio/ui/ui/button-variants";
import { Logo } from "@/features/auth/auth-ui";

export default async function LandingPage({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  setRequestLocale(locale);
  const t = await getTranslations("common");

  return (
    <div className="w-full max-w-150">
      <div className="mb-7 flex flex-col items-center text-center">
        <Logo size="xxl" priority />
        <p className="mt-1.5 text-2xl leading-relaxed text-(--muted-clr)">
          {t("meta.description")}
        </p>
      </div>
      <div className="flex flex-col gap-2.5">
        <Link
          href="/login"
          className={buttonVariants({ size: "default", className: "w-full" })}
        >
          {t("nav.signIn")}
        </Link>
        <Link
          href="/register"
          className={buttonVariants({
            size: "default",
            variant: "outline",
            className: "w-full",
          })}
        >
          {t("nav.getStarted")}
        </Link>
      </div>
      <p className="mt-5 text-center text-[11px] text-(--faint) tracking-wide">
        © {new Date().getFullYear()} Hublio
      </p>
    </div>
  );
}
