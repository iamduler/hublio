import { LocaleSwitcher } from "@/components/locale-switcher";
import { ThemeSwitcher } from "@/components/theme-switcher";
import { RedirectIfAuthenticated } from "@/components/auth/RedirectIfAuthenticated";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-full flex-col bg-[var(--bg)]">
      <div className="flex justify-end gap-2 p-4">
        <ThemeSwitcher variant="auth" />
        <LocaleSwitcher variant="auth" />
      </div>
      <div className="flex flex-1 items-center justify-center px-4 pb-8">
        <RedirectIfAuthenticated>{children}</RedirectIfAuthenticated>
      </div>
    </div>
  );
}
