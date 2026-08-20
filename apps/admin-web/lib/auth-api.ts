export interface SuperAdminIdentity {
  email: string;
  display_name: string;
}

export interface AuthSession {
  super_admin: SuperAdminIdentity;
  csrf_token: string;
  local_development: boolean;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface PendingOTPChallenge {
  challenge_id: string;
  expires_at: string;
  resend_available_at: string;
  destination_hint: string;
}

export type OTPChallengeState =
  | "PENDING"
  | "EXPIRED"
  | "ATTEMPTS_EXCEEDED"
  | "INVALIDATED"
  | "CONSUMED";

export interface OTPChallengeContext extends PendingOTPChallenge {
  state: OTPChallengeState;
}

export type LoginResult =
  | { authenticationState: "AUTHENTICATED"; session: AuthSession }
  | { authenticationState: "OTP_REQUIRED"; challenge: PendingOTPChallenge };

export type AuthErrorCode =
  | "INVALID_CREDENTIALS"
  | "AUTH_INVALID_CREDENTIALS"
  | "AUTHENTICATION_RATE_LIMITED"
  | "AUTHENTICATION_UNAVAILABLE"
  | "AUTH_TEMPORARILY_UNAVAILABLE"
  | "OTP_REQUIRED"
  | "AUTH_OTP_INVALID"
  | "AUTH_OTP_EXPIRED"
  | "AUTH_OTP_ATTEMPTS_EXCEEDED"
  | "AUTH_OTP_INVALIDATED"
  | "AUTH_OTP_CONSUMED"
  | "AUTH_OTP_RESEND_TOO_EARLY"
  | "AUTH_OTP_RATE_LIMITED"
  | "INVALID_REQUEST"
  | "UNEXPECTED";

export class AuthRequestError extends Error {
  readonly code: AuthErrorCode;
  readonly status: number | null;

  constructor(code: AuthErrorCode, status: number | null) {
    super(code);
    this.name = "AuthRequestError";
    this.code = code;
    this.status = status;
  }
}

interface AuthSuccessEnvelope {
  data: AuthSession;
}

interface OTPRequiredEnvelope {
  data: {
    authentication_state: "OTP_REQUIRED";
    challenge: PendingOTPChallenge;
  };
}

const knownErrorCodes = new Set<AuthErrorCode>([
  "INVALID_CREDENTIALS",
  "AUTH_INVALID_CREDENTIALS",
  "AUTHENTICATION_RATE_LIMITED",
  "AUTHENTICATION_UNAVAILABLE",
  "AUTH_TEMPORARILY_UNAVAILABLE",
  "OTP_REQUIRED",
  "AUTH_OTP_INVALID",
  "AUTH_OTP_EXPIRED",
  "AUTH_OTP_ATTEMPTS_EXCEEDED",
  "AUTH_OTP_INVALIDATED",
  "AUTH_OTP_CONSUMED",
  "AUTH_OTP_RESEND_TOO_EARLY",
  "AUTH_OTP_RATE_LIMITED",
  "INVALID_REQUEST",
]);

export async function login(credentials: LoginCredentials): Promise<LoginResult> {
  const response = await request("/api/v1/auth/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    // Keep this DTO explicit. Privileged account or session fields must never
    // become browser-controlled through object spreading or model binding.
    body: JSON.stringify({
      email: credentials.email,
      password: credentials.password,
    }),
  });
  const payload = await readJSON(response);

  if (isOTPRequiredEnvelope(payload)) {
    return {
      authenticationState: "OTP_REQUIRED",
      challenge: payload.data.challenge,
    };
  }

  if (isAuthSuccessEnvelope(payload)) {
    return { authenticationState: "AUTHENTICATED", session: payload.data };
  }

  throw new AuthRequestError("UNEXPECTED", response.status);
}

export async function getOTPChallengeStatus(
  challengeID: string,
  signal?: AbortSignal,
): Promise<OTPChallengeContext> {
  const response = await request("/api/v1/auth/otp/status", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ challenge_id: challengeID }),
    cache: "no-store",
    signal,
  });

  return parseChallengeResponse(response);
}

export async function verifyOTP(challengeID: string, otp: string): Promise<AuthSession> {
  const response = await request("/api/v1/auth/otp/verify", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      challenge_id: challengeID,
      otp,
    }),
  });

  return parseSessionResponse(response);
}

export async function resendOTP(challengeID: string): Promise<OTPChallengeContext> {
  const response = await request("/api/v1/auth/otp/resend", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ challenge_id: challengeID }),
  });

  return parseChallengeResponse(response);
}

export async function getSession(signal?: AbortSignal): Promise<AuthSession> {
  const response = await request("/api/v1/auth/session", {
    method: "GET",
    cache: "no-store",
    signal,
  });

  return parseSessionResponse(response);
}

export async function logout(csrfToken: string): Promise<void> {
  await request("/api/v1/auth/logout", {
    method: "POST",
    headers: {
      // Logout changes authenticated state, so the session cookie alone is
      // insufficient. This token is held only in runtime memory.
      "X-CSRF-Token": csrfToken,
    },
  });
}

export function getLoginErrorMessage(error: unknown): string {
  if (!(error instanceof AuthRequestError)) {
    return "Something went wrong. Please try again.";
  }

  switch (error.code) {
    case "INVALID_CREDENTIALS":
    case "AUTH_INVALID_CREDENTIALS":
      return "The email or password is incorrect.";
    case "AUTHENTICATION_RATE_LIMITED":
      return "Too many sign-in attempts. Please wait and try again.";
    case "AUTHENTICATION_UNAVAILABLE":
    case "AUTH_TEMPORARILY_UNAVAILABLE":
      return "Sign in is temporarily unavailable. Please try again shortly.";
    case "INVALID_REQUEST":
      return "Check your email and password and try again.";
    case "OTP_REQUIRED":
      return "Email verification is required. Please sign in again.";
    case "AUTH_OTP_INVALID":
    case "AUTH_OTP_EXPIRED":
    case "AUTH_OTP_ATTEMPTS_EXCEEDED":
    case "AUTH_OTP_INVALIDATED":
    case "AUTH_OTP_CONSUMED":
    case "AUTH_OTP_RESEND_TOO_EARLY":
    case "AUTH_OTP_RATE_LIMITED":
    case "UNEXPECTED":
      return "Something went wrong. Please try again.";
  }
}

async function request(path: string, init: RequestInit): Promise<Response> {
  let response: Response;

  try {
    response = await fetch(path, {
      ...init,
      credentials: "same-origin",
    });
  } catch (error) {
    if (isAbortError(error)) {
      throw error;
    }

    throw new AuthRequestError("AUTHENTICATION_UNAVAILABLE", null);
  }

  if (response.ok) {
    return response;
  }

  const payload = await readJSON(response);
  throw new AuthRequestError(readErrorCode(payload, response.status), response.status);
}

async function parseSessionResponse(response: Response): Promise<AuthSession> {
  const payload = await readJSON(response);

  if (!isAuthSuccessEnvelope(payload)) {
    throw new AuthRequestError("UNEXPECTED", response.status);
  }

  return payload.data;
}

async function parseChallengeResponse(response: Response): Promise<OTPChallengeContext> {
  const payload = await readJSON(response);

  if (!isRecord(payload) || !isRecord(payload.data)) {
    throw new AuthRequestError("UNEXPECTED", response.status);
  }

  if (isOTPChallengeContext(payload.data)) {
    return payload.data;
  }

  // Status and resend use the same context. Accepting a nested challenge keeps
  // the client compatible with the login envelope without weakening validation.
  if (isRecord(payload.data.challenge)) {
    const candidate = { ...payload.data.challenge, state: payload.data.state };
    if (isOTPChallengeContext(candidate)) {
      return candidate;
    }
  }

  throw new AuthRequestError("UNEXPECTED", response.status);
}

async function readJSON(response: Response): Promise<unknown> {
  try {
    return await response.json();
  } catch {
    return null;
  }
}

function readErrorCode(payload: unknown, status: number): AuthErrorCode {
  if (isRecord(payload) && isRecord(payload.error)) {
    const code = payload.error.code;
    if (typeof code === "string" && knownErrorCodes.has(code as AuthErrorCode)) {
      return code as AuthErrorCode;
    }
  }

  // The Next.js proxy may generate a non-JSON response when the Go process is
  // unreachable. Status fallbacks preserve useful UX without exposing proxy or
  // infrastructure details.
  if (status === 429) {
    return "AUTHENTICATION_RATE_LIMITED";
  }
  if (status >= 500) {
    return "AUTHENTICATION_UNAVAILABLE";
  }
  if (status === 400) {
    return "INVALID_REQUEST";
  }
  if (status === 401) {
    return "INVALID_CREDENTIALS";
  }

  return "UNEXPECTED";
}

function isAuthSuccessEnvelope(payload: unknown): payload is AuthSuccessEnvelope {
  if (!isRecord(payload) || !isRecord(payload.data)) {
    return false;
  }

  const { data } = payload;
  return (
    isRecord(data.super_admin) &&
    typeof data.super_admin.email === "string" &&
    typeof data.super_admin.display_name === "string" &&
    typeof data.csrf_token === "string" &&
    typeof data.local_development === "boolean"
  );
}

function isOTPRequiredEnvelope(payload: unknown): payload is OTPRequiredEnvelope {
  return (
    isRecord(payload) &&
    isRecord(payload.data) &&
    payload.data.authentication_state === "OTP_REQUIRED" &&
    isPendingOTPChallenge(payload.data.challenge)
  );
}

export function isPendingOTPChallenge(value: unknown): value is PendingOTPChallenge {
  if (!isRecord(value)) {
    return false;
  }

  return (
    isOpaqueChallengeID(value.challenge_id) &&
    isTimestamp(value.expires_at) &&
    isTimestamp(value.resend_available_at) &&
    typeof value.destination_hint === "string" &&
    value.destination_hint.length <= 254
  );
}

function isOTPChallengeContext(value: unknown): value is OTPChallengeContext {
  if (!isPendingOTPChallenge(value)) {
    return false;
  }

  const state = (value as { state?: unknown }).state;
  return (
    state === "PENDING" ||
    state === "EXPIRED" ||
    state === "ATTEMPTS_EXCEEDED" ||
    state === "INVALIDATED" ||
    state === "CONSUMED"
  );
}

function isOpaqueChallengeID(value: unknown): value is string {
  return (
    typeof value === "string" &&
    value.length >= 16 &&
    value.length <= 256 &&
    !/\s/.test(value)
  );
}

function isTimestamp(value: unknown): value is string {
  return typeof value === "string" && value.length <= 64 && Number.isFinite(Date.parse(value));
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isAbortError(error: unknown): boolean {
  return error instanceof Error && error.name === "AbortError";
}
