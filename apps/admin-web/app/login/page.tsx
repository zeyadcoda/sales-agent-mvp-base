import { LoginForm } from "./login-form";

export default function LoginPage() {
  return (
    <main className="auth-shell">
      <section className="auth-panel" aria-labelledby="login-heading">
        <div className="brand-lockup" aria-label="Sales Agent">
          <span className="brand-mark" aria-hidden="true">
            SA
          </span>
          <span>Sales Agent</span>
        </div>

        <div className="auth-heading">
          <p className="eyebrow">Super Admin</p>
          <h1 id="login-heading">Super Admin sign in</h1>
          <p>Use your platform administrator credentials to continue.</p>
        </div>

        <LoginForm />
      </section>

      <p className="auth-footnote">Protected internal platform access</p>
    </main>
  );
}
