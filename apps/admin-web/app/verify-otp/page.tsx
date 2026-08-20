import { OTPVerification } from "./otp-verification";

export default function VerifyOTPPage() {
  return (
    <main className="auth-shell">
      <section className="auth-panel" aria-label="Super Admin authentication">
        <div className="brand-lockup" aria-label="Sales Agent">
          <span className="brand-mark" aria-hidden="true">
            SA
          </span>
          <span>Sales Agent</span>
        </div>

        <OTPVerification />
      </section>

      <p className="auth-footnote">Protected internal platform access</p>
    </main>
  );
}
