"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";

import {
  AuthRequestError,
  AuthSession,
  getSession,
  logout,
} from "../../lib/auth-api";

type DashboardState =
  | { status: "loading" }
  | { status: "ready"; session: AuthSession }
  | { status: "error"; message: string };

export function DashboardClient() {
  const router = useRouter();
  const [state, setState] = useState<DashboardState>({ status: "loading" });
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const [logoutError, setLogoutError] = useState<string | null>(null);

  const loadSession = useCallback(
    async (signal?: AbortSignal) => {
      setState({ status: "loading" });

      try {
        const session = await getSession(signal);
        setState({ status: "ready", session });
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") {
          return;
        }

        if (error instanceof AuthRequestError && error.status === 401) {
          router.replace("/login");
          return;
        }

        setState({
          status: "error",
          message: "We could not confirm your session. Please try again.",
        });
      }
    },
    [router],
  );

  useEffect(() => {
    const controller = new AbortController();
    void loadSession(controller.signal);

    return () => controller.abort();
  }, [loadSession]);

  async function handleLogout() {
    if (state.status !== "ready") {
      return;
    }

    setIsLoggingOut(true);
    setLogoutError(null);

    try {
      await logout(state.session.csrf_token);
      router.replace("/login");
    } catch (error) {
      if (error instanceof AuthRequestError && error.status === 401) {
        router.replace("/login");
        return;
      }

      setLogoutError("We could not sign you out. Please try again.");
      setIsLoggingOut(false);
    }
  }

  if (state.status === "loading") {
    return (
      <main className="dashboard-shell dashboard-centered" aria-busy="true">
        <div className="loading-card" role="status">
          <span className="spinner" aria-hidden="true" />
          <span>Loading your dashboard…</span>
        </div>
      </main>
    );
  }

  if (state.status === "error") {
    return (
      <main className="dashboard-shell dashboard-centered">
        <section className="status-card" aria-labelledby="session-error-heading">
          <p className="eyebrow">Sales Agent</p>
          <h1 id="session-error-heading">Dashboard unavailable</h1>
          <p>{state.message}</p>
          <button className="button button-primary" type="button" onClick={() => void loadSession()}>
            Try again
          </button>
        </section>
      </main>
    );
  }

  const { super_admin: superAdmin, local_development: isLocalDevelopment } = state.session;
  const displayName = superAdmin.display_name.trim() || superAdmin.email;

  return (
    <div className="dashboard-shell">
      <header className="dashboard-header">
        <div className="dashboard-brand">
          <span className="brand-mark brand-mark-small" aria-hidden="true">
            SA
          </span>
          <div>
            <span className="brand-name">Sales Agent</span>
            <span className="brand-context">Super Admin</span>
          </div>
        </div>

        <div className="header-actions">
          {isLocalDevelopment ? <span className="environment-badge">Local development</span> : null}
          <button
            className="button button-secondary"
            type="button"
            disabled={isLoggingOut}
            onClick={() => void handleLogout()}
          >
            {isLoggingOut ? "Logging out…" : "Logout"}
          </button>
        </div>
      </header>

      <main className="dashboard-content">
        <section className="dashboard-welcome" aria-labelledby="dashboard-heading">
          <p className="eyebrow">Super Admin workspace</p>
          <h1 id="dashboard-heading">Dashboard</h1>
          <p className="welcome-copy">
            Welcome, <strong>{displayName}</strong>.
          </p>
          <dl className="identity-list">
            <div>
              <dt>Signed in as</dt>
              <dd>{superAdmin.email}</dd>
            </div>
            <div>
              <dt>Access</dt>
              <dd>Super Admin</dd>
            </div>
          </dl>
          <p className="milestone-note">
            Authentication is ready. Additional platform modules will be added in later milestones.
          </p>
          {logoutError ? (
            <div className="alert alert-error dashboard-alert" role="alert">
              {logoutError}
            </div>
          ) : null}
        </section>
      </main>
    </div>
  );
}
