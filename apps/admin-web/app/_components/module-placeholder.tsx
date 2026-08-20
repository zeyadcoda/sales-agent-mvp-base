import Link from "next/link";

import { SuperAdminShell } from "./super-admin-shell";

export type ModuleKey =
  | "organizations"
  | "applications"
  | "packages"
  | "ai-usage"
  | "integrations"
  | "ai-agents"
  | "system-health"
  | "logs"
  | "audit";

interface ModuleDefinition {
  title: string;
  href: string;
  purpose: string;
  implementationNote: string;
  detail?: string;
}

const moduleDefinitions: Record<ModuleKey, ModuleDefinition> = {
  organizations: {
    title: "Organizations",
    href: "/organizations",
    purpose: "Create and manage the customer organizations that use Sales Agent.",
    implementationNote:
      "The navigation shell is ready. Organization management will be implemented in the Organizations milestone.",
    detail:
      "The Create Organization action leads here intentionally; no organization form or record creation is available yet.",
  },
  applications: {
    title: "Applications",
    href: "/applications",
    purpose: "Review and manage platform access applications.",
    implementationNote:
      "The navigation shell is ready. Application workflows will be implemented in the Applications milestone.",
  },
  packages: {
    title: "Packages",
    href: "/packages",
    purpose: "Define the packages available to customer organizations.",
    implementationNote:
      "The navigation shell is ready. Package management will be implemented in the Packages milestone.",
  },
  "ai-usage": {
    title: "AI & Usage",
    href: "/ai-usage",
    purpose: "Review measured AI consumption and attributed Agent Run cost.",
    implementationNote:
      "The navigation shell is ready. AI usage and cost attribution require Agent Run cost tracking.",
    detail: "Strategy Credits are separate and are not represented here as AI cost.",
  },
  integrations: {
    title: "Integrations",
    href: "/integrations",
    purpose: "Configure and monitor external platform integrations.",
    implementationNote:
      "The navigation shell is ready. Integration management will be implemented in a later milestone.",
  },
  "ai-agents": {
    title: "AI Agents",
    href: "/ai-agents",
    purpose: "Manage the platform Agent registry and operational configuration.",
    implementationNote:
      "The navigation shell is ready. The Agent registry will be implemented in the AI Agents milestone.",
  },
  "system-health": {
    title: "System Health",
    href: "/system-health",
    purpose: "Understand product-level availability across platform dependencies and operations.",
    implementationNote:
      "The navigation shell is ready. Full product-level System Health has not been implemented.",
    detail:
      "Core runtime readiness exists. Future sources include AI Runtime, Core Providers, Search/Research, Meta dependencies, Notification Email, Background Jobs, File Processing where applicable, and Agent Operations.",
  },
  logs: {
    title: "Logs",
    href: "/logs",
    purpose: "Inspect protected operational logs for platform investigation.",
    implementationNote:
      "The navigation shell is ready. Log search and viewing will be implemented in the Logs milestone.",
  },
  audit: {
    title: "Audit",
    href: "/audit",
    purpose: "Review protected platform audit events and administrative activity.",
    implementationNote:
      "The navigation shell is ready. The Platform Audit user interface will be implemented in the Audit milestone.",
  },
};

export function ModulePlaceholder({ moduleKey }: { moduleKey: ModuleKey }) {
  const module = moduleDefinitions[moduleKey];

  return (
    <SuperAdminShell currentPath={module.href} pageTitle={module.title}>
      <div className="module-page">
        <div className="page-heading">
          <p className="eyebrow">Super Admin module</p>
          <h1>{module.title}</h1>
          <p className="page-subtitle">{module.purpose}</p>
        </div>

        <section className="module-placeholder" aria-labelledby={`${moduleKey}-state-heading`}>
          <p className="module-state-label">Current implementation state</p>
          <h2 id={`${moduleKey}-state-heading`}>This module has not been implemented yet.</h2>
          <p>{module.implementationNote}</p>
          {module.detail ? <p className="module-detail">{module.detail}</p> : null}
          <Link className="text-link" href="/dashboard">
            Back to Dashboard
          </Link>
        </section>
      </div>
    </SuperAdminShell>
  );
}
