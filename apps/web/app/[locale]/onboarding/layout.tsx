import { LocaleSwitcher } from "@/components/locale-switcher";
import { ThemeSwitcher } from "@/components/theme-switcher";

export default function OnboardingLayout({
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
      {children}
    </div>
  );
}
