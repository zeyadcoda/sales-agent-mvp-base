"use client";

import { FormEvent, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";

import {
  AuthRequestError,
  getOTPChallengeStatus,
  OTPChallengeContext,
  OTPChallengeState,
  resendOTP,
  verifyOTP,
} from "../../lib/auth-api";
import {
  clearPendingOTPChallenge,
  loadPendingOTPChallenge,
  savePendingOTPChallenge,
} from "../../lib/otp-challenge-storage";

type TerminalChallengeState = Exclude<OTPChallengeState, "PENDING">;

type VerificationViewState =
  | { status: "checking" }
  | { status: "ready"; challenge: OTPChallengeContext }
  | { status: "terminal"; reason: TerminalChallengeState }
  | { status: "unavailable" };

export function OTPVerification() {
  const router = useRouter();
  const inputRef = useRef<HTMLInputElement>(null);
  const [viewState, setViewState] = useState<VerificationViewState>({ status: "checking" });
  const [otp, setOTP] = useState("");
  const [fieldError, setFieldError] = useState<string | null>(null);
  const [requestError, setRequestError] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);
  const [isVerifying, setIsVerifying] = useState(false);
  const [isResending, setIsResending] = useState(false);
  const [now, setNow] = useState(() => Date.now());

  const validateStoredChallenge = useCallback(
    async (signal?: AbortSignal) => {
      const storedChallenge = loadPendingOTPChallenge();
      if (storedChallenge === null) {
        router.replace("/login");
        return;
      }

      setViewState({ status: "checking" });

      try {
        const challenge = await getOTPChallengeStatus(storedChallenge.challenge_id, signal);
        if (challenge.state === "PENDING") {
          savePendingOTPChallenge(challenge);
          setNow(Date.now());
          setViewState({ status: "ready", challenge });
          return;
        }

        showTerminalState(challenge.state, setViewState);
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") {
          return;
        }

        const terminalState = terminalStateFromError(error);
        if (terminalState !== null) {
          showTerminalState(terminalState, setViewState);
          return;
        }

        // Unknown or rejected challenge identifiers are never retried as a
        // privileged flow. The backend remains authoritative for this decision.
        if (
          error instanceof AuthRequestError &&
          (error.code === "AUTH_OTP_INVALID" ||
            error.code === "INVALID_REQUEST" ||
            error.status === 404 ||
            error.status === 410)
        ) {
          showTerminalState("INVALIDATED", setViewState);
          return;
        }

        setViewState({ status: "unavailable" });
      }
    },
    [router],
  );

  useEffect(() => {
    const controller = new AbortController();
    void validateStoredChallenge(controller.signal);

    return () => controller.abort();
  }, [validateStoredChallenge]);

  useEffect(() => {
    if (viewState.status !== "ready") {
      return;
    }

    const timer = window.setInterval(() => setNow(Date.now()), 1_000);
    return () => window.clearInterval(timer);
  }, [viewState.status]);

  const resendSeconds = useMemo(() => {
    if (viewState.status !== "ready") {
      return 0;
    }

    return secondsUntil(viewState.challenge.resend_available_at, now);
  }, [now, viewState]);

  function restartLogin() {
    clearPendingOTPChallenge();
    setOTP("");
    router.replace("/login");
  }

  async function handleVerify(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (viewState.status !== "ready") {
      return;
    }

    if (!/^\d{6}$/.test(otp)) {
      setFieldError("Enter the complete six-digit code.");
      return;
    }

    setFieldError(null);
    setRequestError(null);
    setStatusMessage(null);
    setIsVerifying(true);

    try {
      await verifyOTP(viewState.challenge.challenge_id, otp);
      clearPendingOTPChallenge();
      setOTP("");
      router.replace("/dashboard");
    } catch (error) {
      setOTP("");
      const terminalState = terminalStateFromError(error);
      if (terminalState !== null) {
        showTerminalState(terminalState, setViewState);
        return;
      }

      if (
        error instanceof AuthRequestError &&
        (error.code === "AUTH_OTP_INVALID" || error.code === "INVALID_REQUEST")
      ) {
        setRequestError("The verification code is invalid. Try again.");
      } else if (
        error instanceof AuthRequestError &&
        (error.code === "AUTH_OTP_RATE_LIMITED" ||
          error.code === "AUTHENTICATION_RATE_LIMITED")
      ) {
        setRequestError("Too many verification attempts. Please wait and try again.");
      } else {
        setRequestError("Verification is temporarily unavailable. Please try again.");
      }

      window.setTimeout(() => inputRef.current?.focus(), 0);
    } finally {
      setIsVerifying(false);
    }
  }

  async function handleResend() {
    if (viewState.status !== "ready" || resendSeconds > 0) {
      return;
    }

    setRequestError(null);
    setStatusMessage(null);
    setIsResending(true);

    try {
      const challenge = await resendOTP(viewState.challenge.challenge_id);
      if (challenge.state !== "PENDING") {
        showTerminalState(challenge.state, setViewState);
        return;
      }

      savePendingOTPChallenge(challenge);
      setOTP("");
      setNow(Date.now());
      setViewState({ status: "ready", challenge });
      setStatusMessage("A new verification code has been sent.");
      window.setTimeout(() => inputRef.current?.focus(), 0);
    } catch (error) {
      const terminalState = terminalStateFromError(error);
      if (terminalState !== null) {
        showTerminalState(terminalState, setViewState);
        return;
      }

      if (
        error instanceof AuthRequestError &&
        error.code === "AUTH_OTP_RESEND_TOO_EARLY"
      ) {
        setRequestError("A new code is not available yet. Wait for the countdown and try again.");
      } else if (
        error instanceof AuthRequestError &&
        (error.code === "AUTH_OTP_RATE_LIMITED" ||
          error.code === "AUTHENTICATION_RATE_LIMITED")
      ) {
        setRequestError("Too many resend requests. Please wait and try again.");
      } else {
        setRequestError("We could not send a new code. Please try again.");
      }
    } finally {
      setIsResending(false);
    }
  }

  if (viewState.status === "checking") {
    return (
      <div className="otp-state" role="status" aria-live="polite">
        <span className="spinner" aria-hidden="true" />
        <span>Checking your verification request…</span>
      </div>
    );
  }

  if (viewState.status === "unavailable") {
    return (
      <OTPStatusPanel
        title="Verification unavailable"
        message="We could not confirm your verification request. Please try again."
        primaryLabel="Try again"
        onPrimary={() => void validateStoredChallenge()}
        onRestart={restartLogin}
      />
    );
  }

  if (viewState.status === "terminal") {
    const content = terminalContent(viewState.reason);
    return (
      <OTPStatusPanel
        title={content.title}
        message={content.message}
        primaryLabel="Restart login"
        onPrimary={restartLogin}
      />
    );
  }

  const isBusy = isVerifying || isResending;
  const describedBy = [
    "otp-guidance",
    fieldError ? "otp-field-error" : null,
    requestError ? "otp-request-error" : null,
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <>
      <div className="auth-heading otp-heading">
        <p className="eyebrow">Super Admin</p>
        <h1 id="otp-heading">Verify your email</h1>
        <p>
          Enter the six-digit code sent to {viewState.challenge.destination_hint || "your email"}.
        </p>
      </div>

      <form className="auth-form" onSubmit={handleVerify} noValidate aria-busy={isBusy}>
        {requestError ? (
          <div className="alert alert-error" id="otp-request-error" role="alert">
            {requestError}
          </div>
        ) : null}

        {statusMessage ? (
          <div className="alert alert-success" role="status" aria-live="polite">
            {statusMessage}
          </div>
        ) : null}

        <div className="field-group">
          <label htmlFor="otp">Six-digit verification code</label>
          <input
            ref={inputRef}
            className="otp-input"
            id="otp"
            name="otp"
            type="text"
            inputMode="numeric"
            pattern="[0-9]*"
            autoComplete="one-time-code"
            maxLength={6}
            value={otp}
            disabled={isBusy}
            aria-invalid={fieldError || requestError ? true : undefined}
            aria-describedby={describedBy}
            onChange={(event) => {
              setOTP(event.target.value.replace(/\D/g, "").slice(0, 6));
              setFieldError(null);
              setRequestError(null);
            }}
          />
          <p className="field-guidance" id="otp-guidance">
            Codes expire approximately 10 minutes after they are sent.
          </p>
          {fieldError ? (
            <p className="field-error" id="otp-field-error">
              {fieldError}
            </p>
          ) : null}
        </div>

        <button className="button button-primary button-full" type="submit" disabled={isBusy}>
          {isVerifying ? "Verifying…" : "Verify"}
        </button>

        <div className="otp-resend" aria-live="polite">
          <p>{resendAvailabilityMessage(resendSeconds)}</p>
          <button
            className="button-link"
            type="button"
            disabled={isBusy || resendSeconds > 0}
            onClick={() => void handleResend()}
          >
            {isResending ? "Sending…" : "Resend code"}
          </button>
        </div>

        <button className="button-link otp-restart" type="button" disabled={isBusy} onClick={restartLogin}>
          Restart login
        </button>
      </form>
    </>
  );
}

interface OTPStatusPanelProps {
  title: string;
  message: string;
  primaryLabel: string;
  onPrimary: () => void;
  onRestart?: () => void;
}

function OTPStatusPanel({
  title,
  message,
  primaryLabel,
  onPrimary,
  onRestart,
}: OTPStatusPanelProps) {
  return (
    <div className="auth-heading otp-terminal">
      <p className="eyebrow">Super Admin</p>
      <h1 id="otp-heading">{title}</h1>
      <p role="status" aria-live="polite">
        {message}
      </p>
      <div className="otp-status-actions">
        <button className="button button-primary" type="button" onClick={onPrimary}>
          {primaryLabel}
        </button>
        {onRestart ? (
          <button className="button button-secondary" type="button" onClick={onRestart}>
            Restart login
          </button>
        ) : null}
      </div>
    </div>
  );
}

function showTerminalState(
  reason: TerminalChallengeState,
  setViewState: (state: VerificationViewState) => void,
) {
  clearPendingOTPChallenge();
  setViewState({ status: "terminal", reason });
}

function terminalStateFromError(error: unknown): TerminalChallengeState | null {
  if (!(error instanceof AuthRequestError)) {
    return null;
  }

  switch (error.code) {
    case "AUTH_OTP_EXPIRED":
      return "EXPIRED";
    case "AUTH_OTP_ATTEMPTS_EXCEEDED":
      return "ATTEMPTS_EXCEEDED";
    case "AUTH_OTP_INVALIDATED":
      return "INVALIDATED";
    case "AUTH_OTP_CONSUMED":
      return "CONSUMED";
    default:
      return null;
  }
}

function terminalContent(reason: TerminalChallengeState): { title: string; message: string } {
  switch (reason) {
    case "EXPIRED":
      return {
        title: "Verification code expired",
        message: "This code can no longer be used. Restart login to request a new one.",
      };
    case "ATTEMPTS_EXCEEDED":
      return {
        title: "Verification attempts exceeded",
        message: "This verification request is locked. Restart login to try again.",
      };
    case "INVALIDATED":
      return {
        title: "Verification request invalid",
        message: "This verification request can no longer be used. Restart login to continue.",
      };
    case "CONSUMED":
      return {
        title: "Verification request already used",
        message: "This verification request can no longer be used. Restart login to continue.",
      };
  }
}

function secondsUntil(timestamp: string, now: number): number {
  return Math.max(0, Math.ceil((Date.parse(timestamp) - now) / 1_000));
}

function resendAvailabilityMessage(seconds: number): string {
  if (seconds <= 0) {
    return "You can request a new code now.";
  }

  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = String(seconds % 60).padStart(2, "0");
  return `Resend available in ${minutes}:${remainingSeconds}.`;
}
