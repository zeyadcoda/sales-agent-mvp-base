import { render, screen, waitFor } from "@testing-library/react";
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

describe("Super Admin dashboard", () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    fetchMock.mockReset();
    router.replace.mockReset();
    vi.unstubAllGlobals();
  });

  it("renders authenticated identity from the real session endpoint", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(sessionEnvelope));

    render(<DashboardClient />);

    expect(await screen.findByText("Ada Admin")).toBeInTheDocument();
    expect(screen.getByText("ada@example.com")).toBeInTheDocument();
    expect(screen.getByText("Local development")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Dashboard" })).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/session",
      expect.objectContaining({ method: "GET", cache: "no-store" }),
    );
  });

  it("redirects unauthenticated dashboard access to login", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse(
        {
          error: {
            code: "NO_VALID_SESSION",
            message: "Authentication required.",
            correlation_id: "test-correlation-id",
            field_errors: [],
          },
        },
        401,
      ),
    );

    render(<DashboardClient />);

    await waitFor(() => expect(router.replace).toHaveBeenCalledWith("/login"));
  });

  it("sends the runtime CSRF token on logout and returns to login", async () => {
    fetchMock
      .mockResolvedValueOnce(jsonResponse(sessionEnvelope))
      .mockResolvedValueOnce(jsonResponse({}, 204));

    const user = userEvent.setup();
    render(<DashboardClient />);

    await screen.findByText("Ada Admin");
    await user.click(screen.getByRole("button", { name: "Logout" }));

    await waitFor(() => expect(router.replace).toHaveBeenCalledWith("/login"));
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      "/api/v1/auth/logout",
      expect.objectContaining({
        method: "POST",
        headers: { "X-CSRF-Token": "csrf-runtime-memory-only" },
      }),
    );
  });
});
