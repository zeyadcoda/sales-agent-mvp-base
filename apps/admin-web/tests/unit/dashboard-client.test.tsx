import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { DashboardClient } from "../../app/dashboard/dashboard-client";

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
    local_development: true,
  },
};

const dashboardEnvelope = {
  data: {
    needs_attention: {
      available: false,
      reason: "SOURCE_NOT_IMPLEMENTED",
      items: [],
    },
    ai_cost_consumption: {
      available: false,
      reason: "COST_TRACKING_NOT_IMPLEMENTED",
    },
    organizations: {
      available: false,
      reason: "ORGANIZATIONS_MODULE_NOT_IMPLEMENTED",
    },
    system_health: {
      overall_state: "UNKNOWN",
      reason: "PRODUCT_HEALTH_NOT_IMPLEMENTED",
      core_runtime_readiness: {
        available: true,
        ready: true,
        reason: "CHECK_SUCCEEDED",
      },
    },
    recent_important_activity: {
      available: false,
      reason: "AUDIT_QUERY_NOT_IMPLEMENTED",
      items: [],
    },
  },
};

describe("Super Admin dashboard and protected shell", () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
    mockSuccessfulBackend();
  });

  afterEach(() => {
    fetchMock.mockReset();
    router.replace.mockReset();
    vi.unstubAllGlobals();
  });

  it("renders authenticated identity and the exact approved navigation", async () => {
    render(<DashboardClient />);

    const banner = await screen.findByRole("banner");
    expect(within(banner).getByText("Ada Admin")).toBeInTheDocument();
    expect(within(banner).getByText("ada@example.com")).toBeInTheDocument();
    expect(within(banner).getByText("Local development")).toBeInTheDocument();

    const navigation = screen.getByRole("navigation", { name: "Primary navigation" });
    expect(within(navigation).getAllByRole("link").map((link) => link.textContent)).toEqual([
      "Dashboard",
      "Organizations",
      "Applications",
      "Packages",
      "AI & Usage",
      "Integrations",
      "AI Agents",
      "System Health",
      "Logs",
      "Audit",
    ]);
    expect(within(navigation).getByRole("link", { name: "Dashboard" })).toHaveAttribute(
      "aria-current",
      "page",
    );
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/session",
      expect.objectContaining({ method: "GET", cache: "no-store" }),
    );
  });

  it("renders the five approved sections in order and the primary Organization action", async () => {
    render(<DashboardClient />);

    await screen.findByText(
      "No actionable platform issues are available from implemented modules.",
    );

    const regions = screen.getAllByRole("region");
    expect(regions).toHaveLength(5);
    expect(regions.map((region) => region.getAttribute("aria-labelledby"))).toEqual([
      "needs-attention-heading",
      "ai-cost-heading",
      "organizations-heading",
      "system-health-heading",
      "recent-activity-heading",
    ]);
    expect(regions[0]).toHaveAccessibleName("Needs Attention");
    expect(regions[1]).toHaveAccessibleName("AI Cost & Consumption");
    expect(regions[2]).toHaveAccessibleName("Organizations");
    expect(regions[3]).toHaveAccessibleName("System Health");
    expect(regions[4]).toHaveAccessibleName("Recent Important Activity");

    expect(screen.getByRole("link", { name: "Create Organization" })).toHaveAttribute(
      "href",
      "/organizations",
    );
  });

  it("shows honest unavailable states and separates runtime readiness from overall health", async () => {
    render(<DashboardClient />);

    const healthRegion = await screen.findByRole("region", { name: "System Health" });
    expect(await within(healthRegion).findByText("UNKNOWN")).toBeInTheDocument();
    expect(within(healthRegion).getByText("Ready")).toBeInTheDocument();
    expect(
      within(healthRegion).getByText(/Runtime readiness does not represent the health/),
    ).toBeInTheDocument();

    expect(
      screen.getByText(
        /AI usage and cost attribution will populate after Agent Run cost tracking is implemented/,
      ),
    ).toBeInTheDocument();
    expect(
      screen.getByText(/Organization data will appear after the Organizations module/),
    ).toBeInTheDocument();
    expect(screen.queryByText(/\$0/)).not.toBeInTheDocument();
    expect(screen.queryByText(/0 organizations/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/0 tokens/i)).not.toBeInTheDocument();
  });

  it("redirects unauthenticated shell access to login without requesting dashboard data", async () => {
    fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
      if (String(input) === "/api/v1/auth/session") {
        return jsonResponse(
          {
            error: {
              code: "NO_VALID_SESSION",
              message: "Authentication required.",
              correlation_id: "test-correlation-id",
              field_errors: [],
            },
          },
          401,
        );
      }
      throw new Error(`Unexpected request: ${String(input)}`);
    });

    render(<DashboardClient />);

    await waitFor(() => expect(router.replace).toHaveBeenCalledWith("/login"));
    expect(fetchMock).not.toHaveBeenCalledWith(
      "/api/v1/admin/dashboard",
      expect.anything(),
    );
  });

  it("uses server logout with the in-memory CSRF token", async () => {
    const user = userEvent.setup();
    render(<DashboardClient />);

    await user.click(await screen.findByRole("button", { name: "Logout" }));

    await waitFor(() => expect(router.replace).toHaveBeenCalledWith("/login"));
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/logout",
      expect.objectContaining({
        method: "POST",
        headers: { "X-CSRF-Token": "csrf-runtime-memory-only" },
      }),
    );
  });

  it("retains the authenticated shell and safe section states when the summary backend fails", async () => {
    fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
      const path = String(input);
      if (path === "/api/v1/auth/session") {
        return jsonResponse(sessionEnvelope);
      }
      if (path === "/api/v1/admin/dashboard") {
        return jsonResponse(
          {
            error: {
              code: "DASHBOARD_UNAVAILABLE",
              message: "raw dependency detail that must not render",
            },
          },
          503,
        );
      }
      throw new Error(`Unexpected request: ${path}`);
    });

    render(<DashboardClient />);

    expect(await screen.findByRole("alert")).toHaveTextContent("Dashboard summary unavailable");
    expect(screen.getByRole("navigation", { name: "Primary navigation" })).toBeInTheDocument();
    expect(screen.getByText("ada@example.com")).toBeInTheDocument();
    for (const region of screen.getAllByRole("region")) {
      expect(region).toHaveTextContent("No platform state has been assumed.");
    }
    expect(screen.queryByText(/raw dependency detail/)).not.toBeInTheDocument();
  });

  it("isolates a malformed section while rendering valid authoritative sections", async () => {
    fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
      const path = String(input);
      if (path === "/api/v1/auth/session") {
        return jsonResponse(sessionEnvelope);
      }
      if (path === "/api/v1/admin/dashboard") {
        return jsonResponse({
          data: {
            ...dashboardEnvelope.data,
            needs_attention: { available: false, items: "invalid" },
          },
        });
      }
      throw new Error(`Unexpected request: ${path}`);
    });

    render(<DashboardClient />);

    const needsAttention = await screen.findByRole("region", { name: "Needs Attention" });
    expect(
      await within(needsAttention).findByText(/This section could not be loaded/),
    ).toBeInTheDocument();
    expect(
      within(screen.getByRole("region", { name: "System Health" })).getByText("UNKNOWN"),
    ).toBeInTheDocument();
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });

  it("opens the mobile navigation, moves focus, closes with Escape, and restores focus", async () => {
    vi.stubGlobal(
      "matchMedia",
      vi.fn().mockReturnValue({
        matches: true,
        media: "(max-width: 840px)",
        onchange: null,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        addListener: vi.fn(),
        removeListener: vi.fn(),
        dispatchEvent: vi.fn(),
      }),
    );

    const user = userEvent.setup();
    render(<DashboardClient />);

    const openButton = await screen.findByRole("button", { name: "Open navigation" });
    await user.click(openButton);

    const dialog = await screen.findByRole("dialog", { name: "Super Admin navigation" });
    const closeButton = within(dialog).getByRole("button", { name: "Close navigation" });
    await waitFor(() => expect(closeButton).toHaveFocus());
    expect(openButton).toHaveAttribute("aria-expanded", "true");

    await user.keyboard("{Escape}");

    await waitFor(() => expect(openButton).toHaveFocus());
    expect(openButton).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("dialog", { name: "Super Admin navigation" })).not.toBeInTheDocument();
  });

  function mockSuccessfulBackend() {
    fetchMock.mockImplementation(
      async (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
        const path = String(input);
        if (path === "/api/v1/auth/session") {
          return jsonResponse(sessionEnvelope);
        }
        if (path === "/api/v1/admin/dashboard") {
          return jsonResponse(dashboardEnvelope);
        }
        if (path === "/api/v1/auth/logout" && init?.method === "POST") {
          return jsonResponse({}, 204);
        }
        throw new Error(`Unexpected request: ${path}`);
      },
    );
  }
});
