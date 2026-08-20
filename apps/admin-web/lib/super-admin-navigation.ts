export interface SuperAdminNavigationItem {
  label: string;
  href: string;
}

export const superAdminNavigation = [
  { label: "Dashboard", href: "/dashboard" },
  { label: "Organizations", href: "/organizations" },
  { label: "Applications", href: "/applications" },
  { label: "Packages", href: "/packages" },
  { label: "AI & Usage", href: "/ai-usage" },
  { label: "Integrations", href: "/integrations" },
  { label: "AI Agents", href: "/ai-agents" },
  { label: "System Health", href: "/system-health" },
  { label: "Logs", href: "/logs" },
  { label: "Audit", href: "/audit" },
] as const satisfies readonly SuperAdminNavigationItem[];
