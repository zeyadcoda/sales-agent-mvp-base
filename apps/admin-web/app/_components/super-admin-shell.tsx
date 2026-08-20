"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import type { ReactNode } from "react";
import { useCallback, useEffect, useRef, useState } from "react";

import { AuthRequestError, type AuthSession, getSession, logout } from "../../lib/auth-api";
import { superAdminNavigation } from "../../lib/super-admin-navigation";

interface SuperAdminShellProps {
  currentPath: string;
  pageTitle: string;
  children: ReactNode;
}

type SessionState =
  | { status: "loading" }
  | { status: "ready"; session: AuthSession }
  | { status: "error" };

const mobileNavigationQuery = "(max-width: 840px)";

export function SuperAdminShell({ currentPath, pageTitle, children }: SuperAdminShellProps) {
  const router = useRouter();
  const [sessionState, setSessionState] = useState<SessionState>({ status: "loading" });
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const [logoutError, setLogoutError] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const [isMobileNavigationOpen, setIsMobileNavigationOpen] = useState(false);
  const menuButtonRef = useRef<HTMLButtonElement>(null);
  const navigationPanelRef = useRef<HTMLElement>(null);

  const loadSession = useCallback(
    async (signal?: AbortSignal) => {
      setSessionState({ status: "loading" });

      try {
        const session = await getSession(signal);
        setSessionState({ status: "ready", session });
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") {
          return;
        }

        if (error instanceof AuthRequestError && error.status === 401) {
          router.replace("/login");
          return;
        }

        setSessionState({ status: "error" });
      }
    },
    [router],
  );

  useEffect(() => {
    const controller = new AbortController();
    void loadSession(controller.signal);

    return () => controller.abort();
  }, [loadSession]);

  useEffect(() => {
    const mediaQuery = window.matchMedia?.(mobileNavigationQuery);
    if (!mediaQuery) {
      return;
    }

    const updateViewport = () => {
      setIsMobile(mediaQuery.matches);
      if (!mediaQuery.matches) {
        setIsMobileNavigationOpen(false);
      }
    };

    updateViewport();
    mediaQuery.addEventListener("change", updateViewport);
    return () => mediaQuery.removeEventListener("change", updateViewport);
  }, []);

  const closeMobileNavigation = useCallback((restoreFocus: boolean) => {
    setIsMobileNavigationOpen(false);
    if (restoreFocus) {
      window.setTimeout(() => menuButtonRef.current?.focus(), 0);
    }
  }, []);

  useEffect(() => {
    if (!isMobileNavigationOpen) {
      return;
    }

    const panel = navigationPanelRef.current;
    const previousBodyOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    const focusableElements = () =>
      Array.from(
        panel?.querySelectorAll<HTMLElement>(
          'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])',
        ) ?? [],
      );

    focusableElements()[0]?.focus();

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        event.preventDefault();
        closeMobileNavigation(true);
        return;
      }

      if (event.key !== "Tab") {
        return;
      }

      const focusable = focusableElements();
      if (focusable.length === 0) {
        event.preventDefault();
        return;
      }

      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      document.body.style.overflow = previousBodyOverflow;
    };
  }, [closeMobileNavigation, isMobileNavigationOpen]);

  async function handleLogout() {
    if (sessionState.status !== "ready") {
      return;
    }

    setIsLoggingOut(true);
    setLogoutError(false);

    try {
      await logout(sessionState.session.csrf_token);
      router.replace("/login");
    } catch (error) {
      if (error instanceof AuthRequestError && error.status === 401) {
        router.replace("/login");
        return;
      }

      setLogoutError(true);
      setIsLoggingOut(false);
    }
  }

  if (sessionState.status === "loading") {
    return (
      <main className="protected-state" aria-busy="true">
        <div className="loading-card" role="status">
          <span className="spinner" aria-hidden="true" />
          <span>Loading your Super Admin workspace…</span>
        </div>
      </main>
    );
  }

  if (sessionState.status === "error") {
    return (
      <main className="protected-state">
        <section className="status-card" aria-labelledby="session-error-heading">
          <p className="eyebrow">Sales Agent</p>
          <h1 id="session-error-heading">Workspace unavailable</h1>
          <p>We could not confirm your session. Please try again.</p>
          <button
            className="button button-primary"
            type="button"
            onClick={() => void loadSession()}
          >
            Try again
          </button>
        </section>
      </main>
    );
  }

  const { session } = sessionState;
  const displayName = session.super_admin.display_name.trim() || session.super_admin.email;

  return (
    <div className="admin-shell">
      <a className="skip-link" href="#main-content">
        Skip to main content
      </a>

      <aside
        id="super-admin-navigation"
        className={`admin-sidebar${isMobileNavigationOpen ? " admin-sidebar-open" : ""}`}
        ref={navigationPanelRef}
        aria-label="Super Admin navigation"
        aria-hidden={isMobile && !isMobileNavigationOpen ? true : undefined}
        aria-modal={isMobile && isMobileNavigationOpen ? true : undefined}
        role={isMobile && isMobileNavigationOpen ? "dialog" : undefined}
        inert={isMobile && !isMobileNavigationOpen ? true : undefined}
      >
        <div className="sidebar-heading">
          <div className="dashboard-brand" aria-label="Sales Agent, Super Admin">
            <span className="brand-mark brand-mark-small" aria-hidden="true">
              SA
            </span>
            <div>
              <span className="brand-name">Sales Agent</span>
              <span className="brand-context">Super Admin</span>
            </div>
          </div>
          <button
            className="icon-button mobile-nav-close"
            type="button"
            aria-label="Close navigation"
            aria-controls="super-admin-navigation"
            aria-expanded="true"
            onClick={() => closeMobileNavigation(true)}
          >
            <span aria-hidden="true">×</span>
          </button>
        </div>

        <nav className="admin-navigation" aria-label="Primary navigation">
          {superAdminNavigation.map((item) => {
            const isCurrent = currentPath === item.href;
            return (
              <Link
                className={`admin-navigation-link${isCurrent ? " admin-navigation-link-current" : ""}`}
                href={item.href}
                key={item.href}
                aria-current={isCurrent ? "page" : undefined}
                onClick={() => setIsMobileNavigationOpen(false)}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>
      </aside>

      {isMobileNavigationOpen ? (
        <div
          className="navigation-backdrop"
          aria-hidden="true"
          onClick={() => closeMobileNavigation(true)}
        />
      ) : null}

      <div
        className="admin-workspace"
        aria-hidden={isMobile && isMobileNavigationOpen ? true : undefined}
        inert={isMobile && isMobileNavigationOpen ? true : undefined}
      >
        <header className="admin-header">
          <div className="admin-header-leading">
            <button
              className="icon-button mobile-nav-trigger"
              ref={menuButtonRef}
              type="button"
              aria-controls="super-admin-navigation"
              aria-expanded={isMobileNavigationOpen}
              aria-label={isMobileNavigationOpen ? "Close navigation" : "Open navigation"}
              onClick={() => setIsMobileNavigationOpen((open) => !open)}
            >
              <span aria-hidden="true">☰</span>
            </button>
            <div>
              <p className="admin-page-context">Super Admin</p>
              <p className="admin-current-page">{pageTitle}</p>
            </div>
          </div>

          <div className="admin-header-actions">
            <div className="admin-identity">
              <span className="admin-identity-name">{displayName}</span>
              <span className="admin-identity-email">{session.super_admin.email}</span>
            </div>
            {session.local_development ? (
              <span className="environment-badge">Local development</span>
            ) : null}
            <button
              className="button button-secondary admin-logout"
              type="button"
              disabled={isLoggingOut}
              onClick={() => void handleLogout()}
            >
              {isLoggingOut ? "Logging out…" : "Logout"}
            </button>
          </div>
        </header>

        {logoutError ? (
          <div className="shell-alert" role="alert">
            We could not sign you out. Please try again.
          </div>
        ) : null}

        <main className="admin-main" id="main-content" tabIndex={-1}>
          {children}
        </main>
      </div>
    </div>
  );
}
