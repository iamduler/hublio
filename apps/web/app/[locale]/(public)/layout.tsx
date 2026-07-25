import { AppTopBar } from "@/components/layouts/app-top-bar";

export default function PublicLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-full flex-col">
      <AppTopBar />
      <main className="flex-1">{children}</main>
    </div>
  );
}
