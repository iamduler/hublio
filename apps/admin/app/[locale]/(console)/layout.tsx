import { AppSidebarProvider } from "@hublio/ui/common/app-sidebar";
import { AppShell } from "@hublio/ui/common/app-shell";
import { PlatformAdminGate } from "@/components/platform-admin-gate";
import { AdminHeader } from "@/components/layouts/admin-header";
import { AdminSidebar } from "@/components/layouts/admin-sidebar";
import { OrgProvider } from "@/providers/org-provider";

export default function ConsoleLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <PlatformAdminGate>
      <OrgProvider>
        <AppSidebarProvider>
          <AppShell sidebar={<AdminSidebar />} header={<AdminHeader />}>
            {children}
          </AppShell>
        </AppSidebarProvider>
      </OrgProvider>
    </PlatformAdminGate>
  );
}
