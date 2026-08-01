import { AppSidebarProvider } from "@hublio/ui/common/app-sidebar";
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
          <div className="flex min-h-dvh">
            <AdminSidebar />
            <div className="flex min-w-0 flex-1 flex-col">
              <AdminHeader />
              <main className="flex-1 p-4 md:p-6">{children}</main>
            </div>
          </div>
        </AppSidebarProvider>
      </OrgProvider>
    </PlatformAdminGate>
  );
}
