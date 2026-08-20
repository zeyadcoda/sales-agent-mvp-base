package httpapi

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"time"

	"salesagent.local/backend/internal/platform/auth"
)

const (
	dashboardSessionOperationTimeout   = 5 * time.Second
	dashboardReadinessOperationTimeout = 2 * time.Second
)

// DashboardSessionResolver is the only authentication capability needed by
// the Dashboard. The opaque cookie is resolved authoritatively on every
// request; no browser-supplied account or tenant identifier is accepted.
type DashboardSessionResolver interface {
	ResolveSession(ctx context.Context, rawSessionToken string) (auth.AuthenticatedSession, error)
}

// DashboardHandler serves the currently implemented, intentionally small
// Super Admin Dashboard summary contract.
type DashboardHandler struct {
	sessions  DashboardSessionResolver
	readiness ReadinessChecker
}

func NewDashboardHandler(
	sessions DashboardSessionResolver,
	readiness ReadinessChecker,
) (*DashboardHandler, error) {
	if isNilDashboardDependency(sessions) {
		return nil, errors.New("dashboard session resolver is required")
	}

	return &DashboardHandler{
		sessions:  sessions,
		readiness: readiness,
	}, nil
}

type dashboardResponse struct {
	Data dashboardData `json:"data"`
}

type dashboardData struct {
	NeedsAttention          needsAttentionSection          `json:"needs_attention"`
	AICostConsumption       unavailableDashboardSection    `json:"ai_cost_consumption"`
	Organizations           unavailableDashboardSection    `json:"organizations"`
	SystemHealth            systemHealthSection            `json:"system_health"`
	RecentImportantActivity recentImportantActivitySection `json:"recent_important_activity"`
}

type unavailableDashboardSection struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason"`
}

type needsAttentionSection struct {
	Available bool                 `json:"available"`
	Reason    string               `json:"reason"`
	Items     []needsAttentionItem `json:"items"`
}

// needsAttentionItem intentionally has no fields yet. There is no implemented
// authoritative issue source, and adding a speculative item shape would make
// fake business data easier to introduce accidentally.
type needsAttentionItem struct{}

type recentImportantActivitySection struct {
	Available bool                          `json:"available"`
	Reason    string                        `json:"reason"`
	Items     []recentImportantActivityItem `json:"items"`
}

// recentImportantActivityItem intentionally has no fields until a later
// milestone defines a narrow, protected Platform Audit query contract.
type recentImportantActivityItem struct{}

type systemHealthSection struct {
	OverallState         string               `json:"overall_state"`
	Reason               string               `json:"reason"`
	CoreRuntimeReadiness coreRuntimeReadiness `json:"core_runtime_readiness"`
}

type coreRuntimeReadiness struct {
	Available bool   `json:"available"`
	Ready     bool   `json:"ready"`
	Reason    string `json:"reason"`
}

func (handler *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeDashboardSessionError(w, r, auth.ErrUnauthenticated)
		return
	}

	operationCtx, cancelOperation := context.WithTimeout(
		r.Context(),
		dashboardSessionOperationTimeout,
	)
	defer cancelOperation()
	if _, err := handler.sessions.ResolveSession(operationCtx, cookie.Value); err != nil {
		writeDashboardSessionError(w, r, err)
		return
	}

	// This summary currently has no filters or selectable scope. Reject every
	// query parameter after authentication so browser-provided admin and
	// Organization identifiers cannot become an accidental authorization input.
	if r.URL.RawQuery != "" {
		writeAPIError(
			w,
			r,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"The dashboard request is invalid.",
			nil,
		)
		return
	}

	writeJSON(w, http.StatusOK, dashboardResponse{Data: dashboardData{
		NeedsAttention: needsAttentionSection{
			Available: false,
			Reason:    "SOURCE_NOT_IMPLEMENTED",
			Items:     []needsAttentionItem{},
		},
		AICostConsumption: unavailableDashboardSection{
			Available: false,
			Reason:    "COST_TRACKING_NOT_IMPLEMENTED",
		},
		Organizations: unavailableDashboardSection{
			Available: false,
			Reason:    "ORGANIZATIONS_MODULE_NOT_IMPLEMENTED",
		},
		SystemHealth: systemHealthSection{
			OverallState:         "UNKNOWN",
			Reason:               "PRODUCT_HEALTH_NOT_IMPLEMENTED",
			CoreRuntimeReadiness: handler.checkCoreRuntimeReadiness(r.Context()),
		},
		RecentImportantActivity: recentImportantActivitySection{
			Available: false,
			Reason:    "AUDIT_QUERY_NOT_IMPLEMENTED",
			Items:     []recentImportantActivityItem{},
		},
	}})
}

func (handler *DashboardHandler) checkCoreRuntimeReadiness(
	ctx context.Context,
) coreRuntimeReadiness {
	if isNilDashboardDependency(handler.readiness) {
		return coreRuntimeReadiness{
			Available: false,
			Ready:     false,
			Reason:    "CHECKER_UNAVAILABLE",
		}
	}

	checkCtx, cancelCheck := context.WithTimeout(ctx, dashboardReadinessOperationTimeout)
	defer cancelCheck()
	if err := handler.readiness.Check(checkCtx); err != nil {
		// Dependency details are intentionally discarded. The response says only
		// that the implemented core runtime readiness check did not pass.
		return coreRuntimeReadiness{
			Available: true,
			Ready:     false,
			Reason:    "CHECK_FAILED",
		}
	}

	return coreRuntimeReadiness{
		Available: true,
		Ready:     true,
		Reason:    "CHECK_SUCCEEDED",
	}
}

func isNilDashboardDependency(dependency any) bool {
	if dependency == nil {
		return true
	}

	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func writeDashboardSessionError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, auth.ErrUnauthenticated) {
		writeAPIError(
			w,
			r,
			http.StatusUnauthorized,
			"UNAUTHENTICATED",
			"A valid Super Admin session is required.",
			nil,
		)
		return
	}

	// PostgreSQL errors and all other resolver details collapse into the
	// existing safe authentication-unavailable response contract.
	writeAPIError(
		w,
		r,
		http.StatusServiceUnavailable,
		"AUTHENTICATION_UNAVAILABLE",
		"Authentication is temporarily unavailable.",
		nil,
	)
}
