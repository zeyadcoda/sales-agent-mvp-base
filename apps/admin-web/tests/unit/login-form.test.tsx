import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import LoginPage from "../../app/login/page";
import { AuthRequestError, getLoginErrorMessage } from "../../lib/auth-api";

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

describe("Super Admin login", () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    fetchMock.mockReset();
    router.replace.mockReset();
    vi.unstubAllGlobals();
  });

  it("renders an accessible login form", () => {
    render(<LoginPage />);

    expect(screen.getByRole("heading", { name: "Super Admin sign in" })).toBeInTheDocument();
    expect(screen.getByLabelText("Email")).toHaveAttribute("autocomplete", "username");
    expect(screen.getByLabelText("Password")).toHaveAttribute(
      "autocomplete",
      "current-password",
    );
    expect(screen.getByRole("button", { name: "Sign in" })).toBeEnabled();
  });

  it("validates required and malformed fields before calling the API", async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(screen.getByText("Enter your email address.")).toBeInTheDocument();
    expect(screen.getByText("Enter your password.")).toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalled();

    await user.type(screen.getByLabelText("Email"), "not-an-email");
    await user.type(screen.getByLabelText("Password"), "a password");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(screen.getByText("Enter a valid email address.")).toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("shows the generic invalid-credentials message", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse(
        {
          error: {
            code: "INVALID_CREDENTIALS",
            message: "Invalid credentials.",
            correlation_id: "test-correlation-id",
            field_errors: [],
          },
        },
        401,
      ),
    );

    const user = userEvent.setup();
    render(<LoginPage />);

    await user.type(screen.getByLabelText("Email"), "admin@example.com");
    await user.type(screen.getByLabelText("Password"), "incorrect-password");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "The email or password is incorrect.",
    );
    expect(router.replace).not.toHaveBeenCalled();
  });

  it("shows a backend-unavailable state when the request cannot reach the API", async () => {
    fetchMock.mockRejectedValueOnce(new TypeError("network unavailable"));

    const user = userEvent.setup();
    render(<LoginPage />);

    await user.type(screen.getByLabelText("Email"), "admin@example.com");
    await user.type(screen.getByLabelText("Password"), "a valid-looking password");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Sign in is temporarily unavailable. Please try again shortly.",
    );
  });

  it("sends only email and password and navigates after a successful local login", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        data: {
          super_admin: {
            email: "admin@example.com",
            display_name: "Super Admin",
          },
          csrf_token: "runtime-only-csrf-token",
          local_development: true,
        },
      }),
    );

    const user = userEvent.setup();
    render(<LoginPage />);

    await user.type(screen.getByLabelText("Email"), "  admin@example.com  ");
    await user.type(screen.getByLabelText("Password"), "correct password");
    await user.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => expect(router.replace).toHaveBeenCalledWith("/dashboard"));
    expect(fetchMock).toHaveBeenCalledTimes(1);

    const [path, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(path).toBe("/api/v1/auth/login");
    expect(init.method).toBe("POST");
    expect(JSON.parse(String(init.body))).toEqual({
      email: "admin@example.com",
      password: "correct password",
    });
  });

  it.each([
    ["AUTHENTICATION_RATE_LIMITED", "Too many sign-in attempts. Please wait and try again."],
    [
      "AUTHENTICATION_UNAVAILABLE",
      "Sign in is temporarily unavailable. Please try again shortly.",
    ],
    [
      "OTP_REQUIRED",
      "Email verification is required. OTP verification will be available in the next milestone.",
    ],
    ["INVALID_REQUEST", "Check your email and password and try again."],
    ["UNEXPECTED", "Something went wrong. Please try again."],
  ] as const)("maps %s to a safe user-facing message", (code, expectedMessage) => {
    expect(getLoginErrorMessage(new AuthRequestError(code, 400))).toBe(expectedMessage);
  });
});
