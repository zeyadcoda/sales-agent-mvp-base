import { render, screen, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  ModulePlaceholder,
  type ModuleKey,
} from "../../app/_components/module-placeholder";

const router = vi.hoisted(() => ({
  replace: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => router,
}));

function jsonResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: vi.fn().mockResolvedValue(body),
  } as unknown as Response;
}

const sessionEnvelope = {
  data: {
    super_admin: {
      email: "ada@example.com",
      display_name: "Ada Admin",
    },
    csrf_token: "csrf-runtime-memory-only",
    local_development: false,
  },
};

const modules: Array<{ key: ModuleKey; label: string; href: string }> = [
  { key: "organizations", label: "Organizations", href: "/organizations" },
  { key: "applications", label: "Applications", href: "/applications" },
  { key: "packages", label: "Packages", href: "/packages" },
  { key: "ai-usage", label: "AI & Usage", href: "/ai-usage" },
  { key: "integrations", label: "Integrations", href: "/integrations" },
  { key: "ai-agents", label: "AI Agents", href: "/ai-agents" },
  { key: "system-health", label: "System Health", href: "/system-health" },
  { key: "logs", label: "Logs", href: "/logs" },
  { key: "audit", label: "Audit", href: "/audit" },
];

describe("protected module placeholders", () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
    fetchMock.mockResolvedValue(jsonResponse(sessionEnvelope));
  });

  afterEach(() => {
    fetchMock.mockReset();
    router.replace.mockReset();
    vi.unstubAllGlobals();
  });

  it.each(modules)("renders the $label route as an honest protected placeholder", async (module) => {
    render(<ModulePlaceholder moduleKey={module.key} />);

    expect(await screen.findByRole("heading", { level: 1, name: module.label })).toBeInTheDocument();
    expect(
      screen.getByRole("heading", {
        level: 2,
        name: "This module has not been implemented yet.",
      }),
    ).toBeInTheDocument();

    const navigation = screen.getByRole("navigation", { name: "Primary navigation" });
    expect(within(navigation).getByRole("link", { name: module.label })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(screen.getByRole("link", { name: "Back to Dashboard" })).toHaveAttribute(
      "href",
      "/dashboard",
    );
    expect(screen.queryByRole("button", { name: /create|add|configure/i })).not.toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/session",
      expect.objectContaining({ method: "GET", cache: "no-store" }),
    );
  });

  it("explains why Create Organization leads to the Organizations placeholder", async () => {
    render(<ModulePlaceholder moduleKey="organizations" />);

    expect(
      await screen.findByText(/Create Organization action leads here intentionally/),
    ).toBeInTheDocument();
    expect(screen.getByText(/no organization form or record creation is available yet/)).toBeInTheDocument();
  });
});
