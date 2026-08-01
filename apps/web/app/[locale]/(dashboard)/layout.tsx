import { AppSidebarProvider } from "@hublio/ui/common/app-sidebar";
import { AppShell } from "@hublio/ui/common/app-shell";
import { DashboardSidebar } from "@/components/layouts/dashboard-sidebar";
import { WorkspaceHeader } from "@/components/layouts/workspace-header";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AppSidebarProvider>
      <AppShell sidebar={<DashboardSidebar />} header={<WorkspaceHeader />}>
        {children}
      </AppShell>
    </AppSidebarProvider>
  );
}
