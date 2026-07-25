import { DashboardSidebar } from "@/components/layouts/dashboard-sidebar";
import { WorkspaceHeader } from "@/components/layouts/workspace-header";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-full">
      <DashboardSidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <WorkspaceHeader />
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
