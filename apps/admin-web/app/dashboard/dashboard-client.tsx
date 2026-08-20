import { DashboardOverview } from "../_components/dashboard-overview";
import { SuperAdminShell } from "../_components/super-admin-shell";

export function DashboardClient() {
  return (
    <SuperAdminShell currentPath="/dashboard" pageTitle="Dashboard">
      <DashboardOverview />
    </SuperAdminShell>
  );
}
