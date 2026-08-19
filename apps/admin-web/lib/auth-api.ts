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

export type AuthErrorCode =
  | "INVALID_CREDENTIALS"
  | "AUTHENTICATION_RATE_LIMITED"
  | "AUTHENTICATION_UNAVAILABLE"
  | "OTP_REQUIRED"
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

const knownErrorCodes = new Set<AuthErrorCode>([
  "INVALID_CREDENTIALS",
  "AUTHENTICATION_RATE_LIMITED",
  "AUTHENTICATION_UNAVAILABLE",
  "OTP_REQUIRED",
  "INVALID_REQUEST",
]);

export async function login(credentials: LoginCredentials): Promise<AuthSession> {
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

  return parseSessionResponse(response);
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
      return "The email or password is incorrect.";
    case "AUTHENTICATION_RATE_LIMITED":
      return "Too many sign-in attempts. Please wait and try again.";
    case "AUTHENTICATION_UNAVAILABLE":
      return "Sign in is temporarily unavailable. Please try again shortly.";
    case "OTP_REQUIRED":
      return "Email verification is required. OTP verification will be available in the next milestone.";
    case "INVALID_REQUEST":
      return "Check your email and password and try again.";
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

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function isAbortError(error: unknown): boolean {
  return error instanceof Error && error.name === "AbortError";
}
