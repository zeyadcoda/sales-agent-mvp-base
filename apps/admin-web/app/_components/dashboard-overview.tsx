"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useState } from "react";

import {
  DashboardRequestError,
  type DashboardSummary,
  getDashboardSummary,
} from "../../lib/dashboard-api";

type DashboardState =
  | { status: "loading" }
  | { status: "ready"; summary: DashboardSummary }
  | { status: "error" };

export function DashboardOverview() {
  const router = useRouter();
  const [state, setState] = useState<DashboardState>({ status: "loading" });

  const loadDashboard = useCallback(
    async (signal?: AbortSignal) => {
      setState({ status: "loading" });

      try {
        const summary = await getDashboardSummary(signal);
        setState({ status: "ready", summary });
      } catch (error) {
        if (error instanceof Error && error.name === "AbortError") {
          return;
        }

        if (error instanceof DashboardRequestError && error.status === 401) {
          router.replace("/login");
          return;
        }

        setState({ status: "error" });
      }
    },
    [router],
  );

  useEffect(() => {
    const controller = new AbortController();
    void loadDashboard(controller.signal);
    return () => controller.abort();
  }, [loadDashboard]);

  const summary = state.status === "ready" ? state.summary : null;

  return (
    <div className="dashboard-page">
      <div className="page-heading page-heading-with-action">
        <div>
          <p className="eyebrow">Super Admin control panel</p>
          <h1>Dashboard</h1>
          <p className="page-subtitle">
            Review implemented operational signals and move into the next platform task.
          </p>
        </div>
        <Link className="button button-primary primary-page-action" href="/organizations">
          Create Organization
        </Link>
      </div>

      {state.status === "loading" ? (
        <div className="dashboard-loading" role="status" aria-live="polite">
          <span className="spinner" aria-hidden="true" />
          <span>Loading dashboard summary…</span>
        </div>
      ) : null}

      {state.status === "error" ? (
        <div className="dashboard-summary-error" role="alert">
          <div>
            <strong>Dashboard summary unavailable</strong>
            <span>
              Operational data could not be loaded. No platform state has been assumed.
            </span>
          </div>
          <button className="button button-secondary" type="button" onClick={() => void loadDashboard()}>
            Try again
          </button>
        </div>
      ) : null}

      <div className="dashboard-sections">
        <section className="dashboard-section" aria-labelledby="needs-attention-heading">
          <SectionHeading
            id="needs-attention-heading"
            title="Needs Attention"
            description="Actionable signals from platform modules that are currently implemented."
          />
          <div className="section-body">
            {state.status === "loading" ? (
              <LoadingSectionCopy />
            ) : summary?.needs_attention ? (
              <NeedsAttentionContent summary={summary.needs_attention} />
            ) : (
              <SectionErrorCopy />
            )}
          </div>
        </section>

        <section className="dashboard-section" aria-labelledby="ai-cost-heading">
          <SectionHeading
            id="ai-cost-heading"
            title="AI Cost & Consumption"
            description="Measured Agent Run usage and attributed AI cost."
          />
          <div className="section-body">
            {state.status === "loading" ? (
              <LoadingSectionCopy />
            ) : summary?.ai_cost_consumption ? (
              summary.ai_cost_consumption.available ? (
                <p>AI cost data is available but cannot be displayed by this dashboard version.</p>
              ) : (
                <EmptyStateCopy>
                  AI usage and cost attribution will populate after Agent Run cost tracking is
                  implemented.
                </EmptyStateCopy>
              )
            ) : (
              <SectionErrorCopy />
            )}
          </div>
        </section>

        <section className="dashboard-section" aria-labelledby="organizations-heading">
          <SectionHeading
            id="organizations-heading"
            title="Organizations"
            description="Customer organization setup and administration."
          />
          <div className="section-body">
            {state.status === "loading" ? (
              <LoadingSectionCopy />
            ) : summary?.organizations ? (
              summary.organizations.available ? (
                <p>Organization data is available but cannot be displayed by this dashboard version.</p>
              ) : (
                <EmptyStateCopy>
                  Organization data will appear after the Organizations module is implemented. The
                  Create Organization action opens that module&apos;s current implementation state.
                </EmptyStateCopy>
              )
            ) : (
              <SectionErrorCopy />
            )}
          </div>
        </section>

        <section className="dashboard-section" aria-labelledby="system-health-heading">
          <SectionHeading
            id="system-health-heading"
            title="System Health"
            description="Product-level health and the currently implemented runtime readiness signal."
          />
          <div className="section-body">
            {state.status === "loading" ? (
              <LoadingSectionCopy />
            ) : summary?.system_health ? (
              <SystemHealthContent summary={summary.system_health} />
            ) : (
              <SectionErrorCopy />
            )}
          </div>
        </section>

        <section className="dashboard-section" aria-labelledby="recent-activity-heading">
          <SectionHeading
            id="recent-activity-heading"
            title="Recent Important Activity"
            description="Security-relevant and operational events from protected audit sources."
          />
          <div className="section-body">
            {state.status === "loading" ? (
              <LoadingSectionCopy />
            ) : summary?.recent_important_activity ? (
              <RecentActivityContent summary={summary.recent_important_activity} />
            ) : (
              <SectionErrorCopy />
            )}
          </div>
        </section>
      </div>
    </div>
  );
}

function SectionHeading({ id, title, description }: { id: string; title: string; description: string }) {
  return (
    <div className="section-heading">
      <div>
        <h2 id={id}>{title}</h2>
        <p>{description}</p>
      </div>
    </div>
  );
}

function NeedsAttentionContent({
  summary,
}: {
  summary: NonNullable<DashboardSummary["needs_attention"]>;
}) {
  if (!summary.available) {
    return (
      <EmptyStateCopy>
        No actionable platform issues are available from implemented modules.
      </EmptyStateCopy>
    );
  }

  if (summary.items.length === 0) {
    return <EmptyStateCopy>No actionable issues were reported by implemented sources.</EmptyStateCopy>;
  }

  return <p>Actionable platform items are available from an implemented source.</p>;
}

function SystemHealthContent({
  summary,
}: {
  summary: NonNullable<DashboardSummary["system_health"]>;
}) {
  const readiness = summary.core_runtime_readiness;
  return (
    <div className="health-summary">
      <dl className="health-status-list">
        <div>
          <dt>Overall System Health</dt>
          <dd>
            <span className={`status-chip status-${summary.overall_state.toLowerCase()}`}>
              {summary.overall_state}
            </span>
          </dd>
        </div>
        <div>
          <dt>Core runtime readiness</dt>
          <dd>
            {readiness.available ? (
              <span className={`status-chip ${readiness.ready ? "status-ready" : "status-not-ready"}`}>
                {readiness.ready ? "Ready" : "Not ready"}
              </span>
            ) : (
              <span className="status-chip status-unknown">Unavailable</span>
            )}
          </dd>
        </div>
      </dl>
      <p className="section-note">
        Overall health remains UNKNOWN until product-level monitoring is implemented. Runtime
        readiness does not represent the health of the entire platform.
      </p>
    </div>
  );
}

function RecentActivityContent({
  summary,
}: {
  summary: NonNullable<DashboardSummary["recent_important_activity"]>;
}) {
  if (!summary.available) {
    return (
      <EmptyStateCopy>
        Recent important activity will appear after the protected Platform Audit query is implemented.
      </EmptyStateCopy>
    );
  }

  if (summary.items.length === 0) {
    return <EmptyStateCopy>No important activity was returned by implemented sources.</EmptyStateCopy>;
  }

  return <p>Important activity is available from an implemented audit source.</p>;
}

function EmptyStateCopy({ children }: { children: React.ReactNode }) {
  return <p className="empty-state-copy">{children}</p>;
}

function LoadingSectionCopy() {
  return <p className="section-loading-copy">Waiting for the protected dashboard summary.</p>;
}

function SectionErrorCopy() {
  return (
    <p className="section-error-copy">
      This section could not be loaded. No platform state has been assumed.
    </p>
  );
}
