import { AppSidebarProvider } from "@hublio/ui/common/app-sidebar";
import { DashboardSidebar } from "@/components/layouts/dashboard-sidebar";
import { WorkspaceHeader } from "@/components/layouts/workspace-header";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AppSidebarProvider>
      <div className="flex min-h-dvh">
        <DashboardSidebar />
        <div className="flex min-w-0 flex-1 flex-col">
          <WorkspaceHeader />
          <main className="flex-1 p-4 md:p-6">{children}</main>
        </div>
      </div>
    </AppSidebarProvider>
  );
}
