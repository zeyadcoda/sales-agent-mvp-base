package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"salesagent.local/backend/internal/platform/auth"
)

type fakeDashboardSessionResolver struct {
	session     auth.AuthenticatedSession
	err         error
	calls       int
	tokens      []string
	deadline    time.Time
	hasDeadline bool
}

func (resolver *fakeDashboardSessionResolver) ResolveSession(
	ctx context.Context,
	rawSessionToken string,
) (auth.AuthenticatedSession, error) {
	resolver.calls++
	resolver.tokens = append(resolver.tokens, rawSessionToken)
	resolver.deadline, resolver.hasDeadline = ctx.Deadline()
	return resolver.session, resolver.err
}

type fakeDashboardReadiness struct {
	err         error
	calls       int
	deadline    time.Time
	hasDeadline bool
}

func (checker *fakeDashboardReadiness) Check(ctx context.Context) error {
	checker.calls++
	checker.deadline, checker.hasDeadline = ctx.Deadline()
	return checker.err
}

func TestNewDashboardHandlerRequiresSessionResolver(t *testing.T) {
	t.Parallel()

	if _, err := NewDashboardHandler(nil, nil); err == nil {
		t.Fatal("NewDashboardHandler() error = nil, want resolver validation error")
	}

	var typedNilResolver *fakeDashboardSessionResolver
	if _, err := NewDashboardHandler(typedNilResolver, nil); err == nil {
		t.Fatal("NewDashboardHandler() error = nil, want typed nil resolver validation error")
	}
}

func TestDashboardRequiresAuthoritativeSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		withCookie    bool
		resolverError error
		wantStatus    int
		wantCode      string
		secretText    string
		wantCalls     int
	}{
		{
			name:       "logged out request has no cookie",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "UNAUTHENTICATED",
		},
		{
			name:          "resolver rejects unknown logged out session",
			withCookie:    true,
			resolverError: auth.ErrUnauthenticated,
			wantStatus:    http.StatusUnauthorized,
			wantCode:      "UNAUTHENTICATED",
			wantCalls:     1,
		},
		{
			name:          "resolver rejects expired session",
			withCookie:    true,
			resolverError: fmt.Errorf("expired_at=2026-08-20T00:00:00Z: %w", auth.ErrUnauthenticated),
			wantStatus:    http.StatusUnauthorized,
			wantCode:      "UNAUTHENTICATED",
			secretText:    "expired_at=2026-08-20T00:00:00Z",
			wantCalls:     1,
		},
		{
			name:          "resolver rejects revoked session",
			withCookie:    true,
			resolverError: fmt.Errorf("revoked_by_internal_id=secret-admin-id: %w", auth.ErrUnauthenticated),
			wantStatus:    http.StatusUnauthorized,
			wantCode:      "UNAUTHENTICATED",
			secretText:    "secret-admin-id",
			wantCalls:     1,
		},
		{
			name:          "session dependency failure is safe",
			withCookie:    true,
			resolverError: errors.New("postgres://user:password@internal.example/sales_agent"),
			wantStatus:    http.StatusServiceUnavailable,
			wantCode:      "AUTHENTICATION_UNAVAILABLE",
			secretText:    "postgres://user:password@internal.example",
			wantCalls:     1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resolver := &fakeDashboardSessionResolver{err: test.resolverError}
			readiness := &fakeDashboardReadiness{}
			router := testDashboardRouter(t, resolver, readiness)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil)
			if test.withCookie {
				request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: testRawToken(31)})
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			assertErrorCode(t, response, test.wantCode)
			assertAuthNoStore(t, response)
			if resolver.calls != test.wantCalls {
				t.Fatalf("ResolveSession calls = %d, want %d", resolver.calls, test.wantCalls)
			}
			if readiness.calls != 0 {
				t.Fatalf("readiness calls = %d, want 0 before authentication succeeds", readiness.calls)
			}
			if test.secretText != "" && strings.Contains(response.Body.String(), test.secretText) {
				t.Fatalf("resolver detail leaked in response: %s", response.Body.String())
			}
			if test.wantCalls == 1 {
				assertDashboardDeadline(
					t,
					resolver.deadline,
					resolver.hasDeadline,
					dashboardSessionOperationTimeout,
					"session resolution",
				)
			}
		})
	}
}

func TestDashboardResolvesSessionOnEveryRequestWithoutExposingIdentityOrSecrets(t *testing.T) {
	t.Parallel()

	rawToken := testRawToken(32)
	resolver := &fakeDashboardSessionResolver{session: auth.AuthenticatedSession{
		SuperAdmin: auth.SuperAdmin{
			ID:           "1df6ae29-36e1-4ef7-b5af-1f1ee524f086",
			Email:        "authoritative-admin@example.test",
			PasswordHash: "$argon2id$never-return-this",
			DisplayName:  "Authoritative Admin",
		},
		CSRFToken: "never-return-csrf",
	}}
	readiness := &fakeDashboardReadiness{}
	router := testDashboardRouter(t, resolver, readiness)

	for requestNumber := 1; requestNumber <= 2; requestNumber++ {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil)
		request.Header.Set("X-Super-Admin-ID", "browser-selected-admin")
		request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: rawToken})
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("request %d status = %d; body = %s", requestNumber, response.Code, response.Body.String())
		}
		assertAuthNoStore(t, response)
		if got := response.Header().Get("Pragma"); got != "no-cache" {
			t.Fatalf("Pragma = %q, want no-cache", got)
		}
		if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
		}
		for _, secret := range []string{
			rawToken,
			"1df6ae29-36e1-4ef7-b5af-1f1ee524f086",
			"authoritative-admin@example.test",
			"$argon2id$never-return-this",
			"never-return-csrf",
			"browser-selected-admin",
		} {
			if strings.Contains(response.Body.String(), secret) {
				t.Fatalf("sensitive or browser-selected value %q leaked: %s", secret, response.Body.String())
			}
		}
	}

	if resolver.calls != 2 {
		t.Fatalf("ResolveSession calls = %d, want one per request", resolver.calls)
	}
	if len(resolver.tokens) != 2 || resolver.tokens[0] != rawToken || resolver.tokens[1] != rawToken {
		t.Fatalf("resolved tokens = %#v, want cookie token twice", resolver.tokens)
	}
	if readiness.calls != 2 {
		t.Fatalf("readiness calls = %d, want one per authenticated request", readiness.calls)
	}
}

func TestDashboardRejectsBrowserSelectedIdentityAndScope(t *testing.T) {
	t.Parallel()

	for _, query := range []string{
		"admin_id=browser-selected-admin",
		"super_admin_id=browser-selected-admin",
		"organization_id=browser-selected-organization",
	} {
		query := query
		t.Run(query, func(t *testing.T) {
			t.Parallel()

			resolver := &fakeDashboardSessionResolver{}
			readiness := &fakeDashboardReadiness{}
			router := testDashboardRouter(t, resolver, readiness)
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/admin/dashboard?"+query,
				nil,
			)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: testRawToken(33)})
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusBadRequest, response.Body.String())
			}
			assertErrorCode(t, response, "INVALID_REQUEST")
			assertAuthNoStore(t, response)
			if resolver.calls != 1 {
				t.Fatalf("ResolveSession calls = %d, want authoritative session check before request validation", resolver.calls)
			}
			if readiness.calls != 0 {
				t.Fatalf("readiness calls = %d, want 0 for rejected selectable scope", readiness.calls)
			}
			if strings.Contains(response.Body.String(), "browser-selected") {
				t.Fatalf("browser-selected identifier reflected in response: %s", response.Body.String())
			}
		})
	}
}

func TestDashboardContractIsHonestAndReadinessIsIndependent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		readiness           *fakeDashboardReadiness
		wantAvailable       bool
		wantReady           bool
		wantReadinessReason string
		wantReadinessCalls  int
		secretText          string
	}{
		{
			name:                "core runtime ready",
			readiness:           &fakeDashboardReadiness{},
			wantAvailable:       true,
			wantReady:           true,
			wantReadinessReason: "CHECK_SUCCEEDED",
			wantReadinessCalls:  1,
		},
		{
			name: "core runtime not ready",
			readiness: &fakeDashboardReadiness{
				err: errors.New("redis://runtime-user:runtime-secret@redis.internal/0"),
			},
			wantAvailable:       true,
			wantReady:           false,
			wantReadinessReason: "CHECK_FAILED",
			wantReadinessCalls:  1,
			secretText:          "runtime-secret",
		},
		{
			name:                "readiness checker missing",
			wantAvailable:       false,
			wantReady:           false,
			wantReadinessReason: "CHECKER_UNAVAILABLE",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resolver := &fakeDashboardSessionResolver{}
			router := testDashboardRouter(t, resolver, test.readiness)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: testRawToken(34)})
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body.String())
			}
			assertAuthNoStore(t, response)
			if test.secretText != "" && strings.Contains(response.Body.String(), test.secretText) {
				t.Fatalf("raw readiness error leaked: %s", response.Body.String())
			}

			var envelope dashboardResponse
			if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode dashboard response: %v; body = %s", err, response.Body.String())
			}
			data := envelope.Data
			if data.NeedsAttention.Available || data.NeedsAttention.Reason != "SOURCE_NOT_IMPLEMENTED" || data.NeedsAttention.Items == nil || len(data.NeedsAttention.Items) != 0 {
				t.Fatalf("needs_attention = %#v, want honest unavailable empty state", data.NeedsAttention)
			}
			if data.AICostConsumption.Available || data.AICostConsumption.Reason != "COST_TRACKING_NOT_IMPLEMENTED" {
				t.Fatalf("ai_cost_consumption = %#v, want unavailable", data.AICostConsumption)
			}
			if data.Organizations.Available || data.Organizations.Reason != "ORGANIZATIONS_MODULE_NOT_IMPLEMENTED" {
				t.Fatalf("organizations = %#v, want unavailable", data.Organizations)
			}
			if data.RecentImportantActivity.Available || data.RecentImportantActivity.Reason != "AUDIT_QUERY_NOT_IMPLEMENTED" || data.RecentImportantActivity.Items == nil || len(data.RecentImportantActivity.Items) != 0 {
				t.Fatalf("recent_important_activity = %#v, want honest unavailable empty state", data.RecentImportantActivity)
			}
			if data.SystemHealth.OverallState != "UNKNOWN" || data.SystemHealth.Reason != "PRODUCT_HEALTH_NOT_IMPLEMENTED" {
				t.Fatalf("system_health = %#v, overall health must remain UNKNOWN", data.SystemHealth)
			}
			if data.SystemHealth.CoreRuntimeReadiness.Available != test.wantAvailable ||
				data.SystemHealth.CoreRuntimeReadiness.Ready != test.wantReady ||
				data.SystemHealth.CoreRuntimeReadiness.Reason != test.wantReadinessReason {
				t.Fatalf(
					"core_runtime_readiness = %#v, want available=%t ready=%t reason=%q",
					data.SystemHealth.CoreRuntimeReadiness,
					test.wantAvailable,
					test.wantReady,
					test.wantReadinessReason,
				)
			}
			if strings.Contains(response.Body.String(), "HEALTHY") ||
				strings.Contains(response.Body.String(), `"count":0`) ||
				strings.Contains(response.Body.String(), `"cost":0`) ||
				strings.Contains(response.Body.String(), `"tokens":0`) ||
				strings.Contains(response.Body.String(), "$0") {
				t.Fatalf("response contains a false health or zero-value business metric: %s", response.Body.String())
			}

			if test.readiness != nil {
				if test.readiness.calls != test.wantReadinessCalls {
					t.Fatalf("readiness calls = %d, want %d", test.readiness.calls, test.wantReadinessCalls)
				}
				assertDashboardDeadline(
					t,
					test.readiness.deadline,
					test.readiness.hasDeadline,
					dashboardReadinessOperationTimeout,
					"readiness check",
				)
			}
		})
	}
}

func TestDashboardResponseHasOnlyApprovedTopLevelSections(t *testing.T) {
	t.Parallel()

	resolver := &fakeDashboardSessionResolver{}
	router := testDashboardRouter(t, resolver, &fakeDashboardReadiness{})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/dashboard", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: testRawToken(35)})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v", err)
	}
	if len(envelope) != 1 || envelope["data"] == nil {
		t.Fatalf("response envelope keys = %#v, want data only", envelope)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(envelope["data"], &data); err != nil {
		t.Fatalf("decode response data: %v", err)
	}
	wantSections := []string{
		"needs_attention",
		"ai_cost_consumption",
		"organizations",
		"system_health",
		"recent_important_activity",
	}
	if len(data) != len(wantSections) {
		t.Fatalf("section count = %d, want %d; body = %s", len(data), len(wantSections), response.Body.String())
	}
	for _, section := range wantSections {
		if data[section] == nil {
			t.Errorf("missing section %q; body = %s", section, response.Body.String())
		}
	}
}

func TestDashboardAdminRouteAndMethodBehavior(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		target     string
		withCookie bool
		wantStatus int
		wantCode   string
		wantAllow  string
		wantCalls  int
	}{
		{
			name:       "GET dashboard",
			method:     http.MethodGet,
			target:     "/api/v1/admin/dashboard",
			withCookie: true,
			wantStatus: http.StatusOK,
			wantCalls:  1,
		},
		{
			name:       "HEAD dashboard",
			method:     http.MethodHead,
			target:     "/api/v1/admin/dashboard",
			withCookie: true,
			wantStatus: http.StatusOK,
			wantCalls:  1,
		},
		{
			name:       "wrong dashboard method",
			method:     http.MethodPost,
			target:     "/api/v1/admin/dashboard",
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   "METHOD_NOT_ALLOWED",
			wantAllow:  "GET, HEAD",
		},
		{
			name:       "unknown admin route",
			method:     http.MethodGet,
			target:     "/api/v1/admin/not-a-route",
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "admin prefix is not a route",
			method:     http.MethodGet,
			target:     "/api/v1/admin",
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			resolver := &fakeDashboardSessionResolver{}
			router := testDashboardRouter(t, resolver, &fakeDashboardReadiness{})
			request := httptest.NewRequest(test.method, test.target, nil)
			if test.withCookie {
				request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: testRawToken(36)})
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			if got := response.Header().Get("Allow"); got != test.wantAllow {
				t.Fatalf("Allow = %q, want %q", got, test.wantAllow)
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}
			assertAuthNoStore(t, response)
			if test.wantCode != "" {
				assertErrorCode(t, response, test.wantCode)
			}
			if resolver.calls != test.wantCalls {
				t.Fatalf("ResolveSession calls = %d, want %d", resolver.calls, test.wantCalls)
			}
		})
	}
}

func testDashboardRouter(
	t *testing.T,
	resolver DashboardSessionResolver,
	readiness ReadinessChecker,
) http.Handler {
	t.Helper()

	handler, err := NewDashboardHandler(resolver, readiness)
	if err != nil {
		t.Fatalf("NewDashboardHandler() error = %v", err)
	}
	return NewRouter(readiness, nil, handler)
}

func assertDashboardDeadline(
	t *testing.T,
	deadline time.Time,
	hasDeadline bool,
	maximum time.Duration,
	operation string,
) {
	t.Helper()
	if !hasDeadline {
		t.Fatalf("%s context must have a deadline", operation)
	}
	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > maximum {
		t.Fatalf("%s deadline has %s remaining, want within (0, %s]", operation, remaining, maximum)
	}
}
