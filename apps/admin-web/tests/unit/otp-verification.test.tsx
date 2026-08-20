import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import VerifyOTPPage from "../../app/verify-otp/page";

const router = vi.hoisted(() => ({
  replace: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => router,
}));

const storageKey = "sales_agent_otp_challenge";

function jsonResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: vi.fn().mockResolvedValue(body),
  } as unknown as Response;
}

function timestampFromNow(milliseconds: number): string {
  return new Date(Date.now() + milliseconds).toISOString();
}

function challenge(
  overrides: Partial<{
    challenge_id: string;
    expires_at: string;
    resend_available_at: string;
    destination_hint: string;
    state: "PENDING" | "EXPIRED" | "ATTEMPTS_EXCEEDED" | "INVALIDATED" | "CONSUMED";
  }> = {},
) {
  return {
    challenge_id: "challenge_0123456789abcdef",
    expires_at: timestampFromNow(10 * 60_000),
    resend_available_at: timestampFromNow(60_000),
    destination_hint: "a***@example.com",
    state: "PENDING" as const,
    ...overrides,
  };
}

function storeChallenge(overrides: Parameters<typeof challenge>[0] = {}) {
  const { state: _state, ...pendingChallenge } = challenge(overrides);
  window.sessionStorage.setItem(storageKey, JSON.stringify(pendingChallenge));
}

function errorEnvelope(code: string) {
  return {
    error: {
      code,
      message: "Safe error.",
      correlation_id: "test-correlation-id",
      field_errors: [],
    },
  };
}

describe("Super Admin OTP verification", () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.stubGlobal("fetch", fetchMock);
    window.sessionStorage.clear();
  });

  afterEach(() => {
    fetchMock.mockReset();
    router.replace.mockReset();
    window.sessionStorage.clear();
    vi.unstubAllGlobals();
  });

  it("redirects direct access without pending challenge context to login", async () => {
    render(<VerifyOTPPage />);

    await waitFor(() => expect(router.replace).toHaveBeenCalledWith("/login"));
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("discards malformed persisted challenge context and redirects to login", async () => {
    window.sessionStorage.setItem(storageKey, JSON.stringify({ challenge_id: "short" }));

    render(<VerifyOTPPage />);

    await waitFor(() => expect(router.replace).toHaveBeenCalledWith("/login"));
    expect(window.sessionStorage.getItem(storageKey)).toBeNull();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("treats a modified but well-formed challenge identifier as invalid", async () => {
    storeChallenge({ challenge_id: "modified_0123456789abcdef" });
    fetchMock.mockResolvedValueOnce(jsonResponse(errorEnvelope("AUTH_OTP_INVALID"), 401));

    render(<VerifyOTPPage />);

    expect(
      await screen.findByRole("heading", { name: "Verification request invalid" }),
    ).toBeInTheDocument();
    expect(window.sessionStorage.getItem(storageKey)).toBeNull();
  });

  it("validates server-side challenge status and renders the accessible OTP form", async () => {
    const pendingChallenge = challenge();
    storeChallenge(pendingChallenge);
    fetchMock.mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }));

    render(<VerifyOTPPage />);

    expect(await screen.findByRole("heading", { name: "Verify your email" })).toBeInTheDocument();
    expect(screen.getByText(/a\*\*\*@example\.com/)).toBeInTheDocument();

    const input = screen.getByLabelText("Six-digit verification code");
    expect(input).toHaveAttribute("inputmode", "numeric");
    expect(input).toHaveAttribute("autocomplete", "one-time-code");
    expect(input).toHaveAttribute("maxlength", "6");
    expect(screen.getByRole("button", { name: "Verify" })).toBeEnabled();

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/auth/otp/status",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ challenge_id: pendingChallenge.challenge_id }),
      }),
    );
  });

  it("does not submit an incomplete or malformed code", async () => {
    const pendingChallenge = challenge();
    storeChallenge(pendingChallenge);
    fetchMock.mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }));

    const user = userEvent.setup();
    render(<VerifyOTPPage />);

    const input = await screen.findByLabelText("Six-digit verification code");
    await user.type(input, "12a34");
    expect(input).toHaveValue("1234");
    await user.click(screen.getByRole("button", { name: "Verify" }));

    expect(screen.getByText("Enter the complete six-digit code.")).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("shows a generic error for a wrong OTP and never persists it", async () => {
    const pendingChallenge = challenge({ resend_available_at: timestampFromNow(-1_000) });
    storeChallenge(pendingChallenge);
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }))
      .mockResolvedValueOnce(jsonResponse(errorEnvelope("AUTH_OTP_INVALID"), 400));

    const user = userEvent.setup();
    render(<VerifyOTPPage />);

    const input = await screen.findByLabelText("Six-digit verification code");
    await user.type(input, "123456");
    await user.click(screen.getByRole("button", { name: "Verify" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "The verification code is invalid. Try again.",
    );
    expect(input).toHaveValue("");
    expect(Object.keys(JSON.parse(window.sessionStorage.getItem(storageKey) ?? "{}"))).toEqual([
      "challenge_id",
      "expires_at",
      "resend_available_at",
      "destination_hint",
    ]);

    const [, verifyInit] = fetchMock.mock.calls[1] as [string, RequestInit];
    expect(JSON.parse(String(verifyInit.body))).toEqual({
      challenge_id: pendingChallenge.challenge_id,
      otp: "123456",
    });
  });

  it("navigates to Dashboard and clears challenge state after a correct OTP", async () => {
    const pendingChallenge = challenge();
    storeChallenge(pendingChallenge);
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }))
      .mockResolvedValueOnce(
        jsonResponse({
          data: {
            super_admin: { email: "admin@example.com", display_name: "Super Admin" },
            csrf_token: "runtime-only-csrf-token",
            local_development: true,
          },
        }),
      );

    const user = userEvent.setup();
    render(<VerifyOTPPage />);

    await user.type(await screen.findByLabelText("Six-digit verification code"), "001284");
    await user.click(screen.getByRole("button", { name: "Verify" }));

    await waitFor(() => expect(router.replace).toHaveBeenCalledWith("/dashboard"));
    expect(window.sessionStorage.getItem(storageKey)).toBeNull();
  });

  it("shows the resend countdown and disables resend until it reaches zero", async () => {
    const pendingChallenge = challenge({ resend_available_at: timestampFromNow(60_000) });
    storeChallenge(pendingChallenge);
    fetchMock.mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }));

    render(<VerifyOTPPage />);

    expect(await screen.findByText(/Resend available in (0:5\d|1:00)\./)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Resend code" })).toBeDisabled();
  });

  it("handles an authoritative resend-too-early response", async () => {
    const pendingChallenge = challenge({ resend_available_at: timestampFromNow(-1_000) });
    storeChallenge(pendingChallenge);
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }))
      .mockResolvedValueOnce(
        jsonResponse(errorEnvelope("AUTH_OTP_RESEND_TOO_EARLY"), 429),
      );

    const user = userEvent.setup();
    render(<VerifyOTPPage />);

    await user.click(await screen.findByRole("button", { name: "Resend code" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "A new code is not available yet. Wait for the countdown and try again.",
    );
  });

  it("updates safe context and announces a successful resend", async () => {
    const pendingChallenge = challenge({ resend_available_at: timestampFromNow(-1_000) });
    const resentChallenge = challenge({
      expires_at: timestampFromNow(11 * 60_000),
      resend_available_at: timestampFromNow(60_000),
    });
    storeChallenge(pendingChallenge);
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }))
      .mockResolvedValueOnce(jsonResponse({ data: resentChallenge }));

    const user = userEvent.setup();
    render(<VerifyOTPPage />);

    await user.click(await screen.findByRole("button", { name: "Resend code" }));

    expect(await screen.findByRole("status")).toHaveTextContent(
      "A new verification code has been sent.",
    );
    expect(window.sessionStorage.getItem(storageKey)).toContain(resentChallenge.expires_at);
    expect(screen.getByRole("button", { name: "Resend code" })).toBeDisabled();
  });

  it.each([
    ["EXPIRED", "Verification code expired"],
    ["ATTEMPTS_EXCEEDED", "Verification attempts exceeded"],
    ["INVALIDATED", "Verification request invalid"],
    ["CONSUMED", "Verification request already used"],
  ] as const)("renders the controlled %s terminal state", async (state, heading) => {
    const terminalChallenge = challenge({ state });
    storeChallenge(terminalChallenge);
    fetchMock.mockResolvedValueOnce(jsonResponse({ data: terminalChallenge }));

    render(<VerifyOTPPage />);

    expect(await screen.findByRole("heading", { name: heading })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Restart login" })).toBeEnabled();
    expect(window.sessionStorage.getItem(storageKey)).toBeNull();
  });

  it("maps a fifth-attempt response to the attempts-exceeded state", async () => {
    const pendingChallenge = challenge();
    storeChallenge(pendingChallenge);
    fetchMock
      .mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }))
      .mockResolvedValueOnce(
        jsonResponse(errorEnvelope("AUTH_OTP_ATTEMPTS_EXCEEDED"), 423),
      );

    const user = userEvent.setup();
    render(<VerifyOTPPage />);

    await user.type(await screen.findByLabelText("Six-digit verification code"), "999999");
    await user.click(screen.getByRole("button", { name: "Verify" }));

    expect(
      await screen.findByRole("heading", { name: "Verification attempts exceeded" }),
    ).toBeInTheDocument();
    expect(window.sessionStorage.getItem(storageKey)).toBeNull();
  });

  it("shows an unavailable state and allows server status retry", async () => {
    const pendingChallenge = challenge();
    storeChallenge(pendingChallenge);
    fetchMock
      .mockResolvedValueOnce(
        jsonResponse(errorEnvelope("AUTH_TEMPORARILY_UNAVAILABLE"), 503),
      )
      .mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }));

    const user = userEvent.setup();
    render(<VerifyOTPPage />);

    expect(
      await screen.findByRole("heading", { name: "Verification unavailable" }),
    ).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Try again" }));
    expect(await screen.findByRole("heading", { name: "Verify your email" })).toBeInTheDocument();
  });

  it("clears pending context when restarting login", async () => {
    const pendingChallenge = challenge();
    storeChallenge(pendingChallenge);
    fetchMock.mockResolvedValueOnce(jsonResponse({ data: pendingChallenge }));

    const user = userEvent.setup();
    render(<VerifyOTPPage />);

    await user.click(await screen.findByRole("button", { name: "Restart login" }));

    expect(window.sessionStorage.getItem(storageKey)).toBeNull();
    expect(router.replace).toHaveBeenCalledWith("/login");
  });
});
