import { afterEach, describe, expect, it, vi } from "vitest";

import { getDashboardSummary } from "../../lib/dashboard-api";

function jsonResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: vi.fn().mockResolvedValue(body),
  } as unknown as Response;
}

describe("dashboard API client", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("uses the protected canonical endpoint without browser-selected identity", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({
        data: {
          needs_attention: { available: false, reason: "SOURCE_NOT_IMPLEMENTED", items: [] },
          ai_cost_consumption: { available: false, reason: "COST_TRACKING_NOT_IMPLEMENTED" },
          organizations: { available: false, reason: "ORGANIZATIONS_MODULE_NOT_IMPLEMENTED" },
          system_health: {
            overall_state: "UNKNOWN",
            reason: "PRODUCT_HEALTH_NOT_IMPLEMENTED",
            core_runtime_readiness: {
              available: false,
              ready: false,
              reason: "CHECKER_UNAVAILABLE",
            },
          },
          recent_important_activity: {
            available: false,
            reason: "AUDIT_QUERY_NOT_IMPLEMENTED",
            items: [],
          },
        },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const summary = await getDashboardSummary();

    expect(summary.system_health?.overall_state).toBe("UNKNOWN");
    expect(summary.system_health?.core_runtime_readiness.available).toBe(false);
    expect(fetchMock).toHaveBeenCalledWith("/api/v1/admin/dashboard", {
      method: "GET",
      credentials: "same-origin",
      cache: "no-store",
      signal: undefined,
    });
    expect(String(fetchMock.mock.calls[0]?.[0])).not.toContain("?");
  });

  it("returns null only for a malformed section", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse({
          data: {
            needs_attention: { available: false, items: [] },
            ai_cost_consumption: { available: "false" },
            organizations: { available: false },
            system_health: {
              overall_state: "UNKNOWN",
              core_runtime_readiness: {
                available: true,
                ready: false,
                reason: "CHECK_FAILED",
              },
            },
            recent_important_activity: { available: false, items: [] },
          },
        }),
      ),
    );

    const summary = await getDashboardSummary();

    expect(summary.needs_attention).not.toBeNull();
    expect(summary.ai_cost_consumption).toBeNull();
    expect(summary.organizations).not.toBeNull();
    expect(summary.system_health?.core_runtime_readiness.ready).toBe(false);
    expect(summary.recent_important_activity).not.toBeNull();
  });
});
