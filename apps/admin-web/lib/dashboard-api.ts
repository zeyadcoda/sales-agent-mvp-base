export type ProductHealthState = "HEALTHY" | "DEGRADED" | "DOWN" | "UNKNOWN";

export type CoreRuntimeReadinessReason =
  | "CHECK_SUCCEEDED"
  | "CHECK_FAILED"
  | "CHECKER_UNAVAILABLE";

export interface NeedsAttentionSummary {
  available: boolean;
  reason?: string;
  items: unknown[];
}

export interface AvailabilitySummary {
  available: boolean;
  reason?: string;
}

export interface CoreRuntimeReadiness {
  available: boolean;
  ready: boolean;
  reason: CoreRuntimeReadinessReason;
}

export interface SystemHealthSummary {
  overall_state: ProductHealthState;
  reason?: string;
  core_runtime_readiness: CoreRuntimeReadiness;
}

export interface RecentImportantActivitySummary {
  available: boolean;
  reason?: string;
  items: unknown[];
}

/**
 * A null section means the server returned a dashboard envelope but that
 * section did not match the protected contract. Keeping that distinction lets
 * the UI fail one section without discarding the other authoritative values.
 */
export interface DashboardSummary {
  needs_attention: NeedsAttentionSummary | null;
  ai_cost_consumption: AvailabilitySummary | null;
  organizations: AvailabilitySummary | null;
  system_health: SystemHealthSummary | null;
  recent_important_activity: RecentImportantActivitySummary | null;
}

export class DashboardRequestError extends Error {
  readonly status: number | null;

  constructor(status: number | null) {
    super("DASHBOARD_UNAVAILABLE");
    this.name = "DashboardRequestError";
    this.status = status;
  }
}

export async function getDashboardSummary(signal?: AbortSignal): Promise<DashboardSummary> {
  let response: Response;

  try {
    response = await fetch("/api/v1/admin/dashboard", {
      method: "GET",
      credentials: "same-origin",
      cache: "no-store",
      signal,
    });
  } catch (error) {
    if (error instanceof Error && error.name === "AbortError") {
      throw error;
    }

    throw new DashboardRequestError(null);
  }

  if (!response.ok) {
    // The UI deliberately does not parse or display dependency/error details.
    throw new DashboardRequestError(response.status);
  }

  const payload = await readJSON(response);
  if (!isRecord(payload) || !isRecord(payload.data)) {
    throw new DashboardRequestError(response.status);
  }

  return {
    needs_attention: parseNeedsAttention(payload.data.needs_attention),
    ai_cost_consumption: parseAvailability(payload.data.ai_cost_consumption),
    organizations: parseAvailability(payload.data.organizations),
    system_health: parseSystemHealth(payload.data.system_health),
    recent_important_activity: parseRecentImportantActivity(
      payload.data.recent_important_activity,
    ),
  };
}

async function readJSON(response: Response): Promise<unknown> {
  try {
    return await response.json();
  } catch {
    return null;
  }
}

function parseNeedsAttention(value: unknown): NeedsAttentionSummary | null {
  if (!isRecord(value) || typeof value.available !== "boolean" || !Array.isArray(value.items)) {
    return null;
  }

  return {
    available: value.available,
    reason: readOptionalReason(value.reason),
    items: value.items,
  };
}

function parseAvailability(value: unknown): AvailabilitySummary | null {
  if (!isRecord(value) || typeof value.available !== "boolean") {
    return null;
  }

  return {
    available: value.available,
    reason: readOptionalReason(value.reason),
  };
}

function parseSystemHealth(value: unknown): SystemHealthSummary | null {
  if (
    !isRecord(value) ||
    !isProductHealthState(value.overall_state) ||
    !isRecord(value.core_runtime_readiness)
  ) {
    return null;
  }

  const readiness = value.core_runtime_readiness;
  if (
    typeof readiness.available !== "boolean" ||
    typeof readiness.ready !== "boolean" ||
    !isCoreRuntimeReadinessReason(readiness.reason)
  ) {
    return null;
  }

  return {
    overall_state: value.overall_state,
    reason: readOptionalReason(value.reason),
    core_runtime_readiness: {
      available: readiness.available,
      ready: readiness.ready,
      reason: readiness.reason,
    },
  };
}

function parseRecentImportantActivity(value: unknown): RecentImportantActivitySummary | null {
  if (!isRecord(value) || typeof value.available !== "boolean" || !Array.isArray(value.items)) {
    return null;
  }

  return {
    available: value.available,
    reason: readOptionalReason(value.reason),
    items: value.items,
  };
}

function readOptionalReason(value: unknown): string | undefined {
  return typeof value === "string" ? value : undefined;
}

function isProductHealthState(value: unknown): value is ProductHealthState {
  return value === "HEALTHY" || value === "DEGRADED" || value === "DOWN" || value === "UNKNOWN";
}

function isCoreRuntimeReadinessReason(value: unknown): value is CoreRuntimeReadinessReason {
  return (
    value === "CHECK_SUCCEEDED" || value === "CHECK_FAILED" || value === "CHECKER_UNAVAILABLE"
  );
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}
