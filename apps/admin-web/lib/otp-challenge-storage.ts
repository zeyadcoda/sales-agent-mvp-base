import { isPendingOTPChallenge } from "./auth-api";
import type { PendingOTPChallenge } from "./auth-api";

const storageKey = "sales_agent_otp_challenge";

export function savePendingOTPChallenge(challenge: PendingOTPChallenge): void {
  if (typeof window === "undefined") {
    return;
  }

  // A challenge ID is not authentication. It is the only browser-persisted
  // flow identifier; passwords, OTPs, session tokens, and CSRF tokens stay out.
  window.sessionStorage.setItem(
    storageKey,
    JSON.stringify({
      challenge_id: challenge.challenge_id,
      expires_at: challenge.expires_at,
      resend_available_at: challenge.resend_available_at,
      destination_hint: challenge.destination_hint,
    }),
  );
}

export function loadPendingOTPChallenge(): PendingOTPChallenge | null {
  if (typeof window === "undefined") {
    return null;
  }

  let stored: string | null;
  try {
    stored = window.sessionStorage.getItem(storageKey);
  } catch {
    return null;
  }
  if (stored === null) {
    return null;
  }

  try {
    const candidate: unknown = JSON.parse(stored);
    if (isPendingOTPChallenge(candidate)) {
      return candidate;
    }
  } catch {
    // Invalid browser state is discarded and cannot become a backend request.
  }

  clearPendingOTPChallenge();
  return null;
}

export function clearPendingOTPChallenge(): void {
  if (typeof window !== "undefined") {
    try {
      window.sessionStorage.removeItem(storageKey);
    } catch {
      // A blocked storage API behaves like absent pending browser context.
    }
  }
}
