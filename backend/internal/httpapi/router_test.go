package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeReadinessChecker lets HTTP tests exercise response behavior without
// granting the HTTP package access to real infrastructure clients.
type fakeReadinessChecker struct {
	err error
}

func (f fakeReadinessChecker) Check(_ context.Context) error {
	return f.err
}

type panicReadinessChecker struct{}

func (panicReadinessChecker) Check(_ context.Context) error {
	panic("liveness must not invoke the readiness checker")
}

// TestLivenessSucceedsWithoutDependencies proves liveness is independent of
// PostgreSQL, Redis, and the readiness checker itself.
func TestLivenessSucceedsWithoutDependencies(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	res := httptest.NewRecorder()

	NewRouter(nil, nil, nil).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}

	if body := res.Body.String(); body != "{\"status\":\"alive\"}\n" {
		t.Fatalf("body = %q, want %q", body, "{\"status\":\"alive\"}\n")
	}

	if contentType := res.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
	}
}

func TestLivenessDoesNotCheckDependencies(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	res := httptest.NewRecorder()

	NewRouter(panicReadinessChecker{}, nil, nil).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
}

func TestReadinessResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		readiness  ReadinessChecker
		wantStatus int
		wantBody   string
	}{
		{
			name:       "all required dependencies healthy",
			readiness:  fakeReadinessChecker{},
			wantStatus: http.StatusOK,
			wantBody:   "{\"status\":\"ready\"}\n",
		},
		{
			name: "dependency unavailable",
			readiness: fakeReadinessChecker{
				err: errors.New("redis://user:secret@internal.example/0"),
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "{\"status\":\"not_ready\"}\n",
		},
		{
			name:       "readiness checker missing",
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   "{\"status\":\"not_ready\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
			res := httptest.NewRecorder()

			NewRouter(tt.readiness, nil, nil).ServeHTTP(res, req)

			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tt.wantStatus)
			}

			if body := res.Body.String(); body != tt.wantBody {
				t.Fatalf("body = %q, want %q", body, tt.wantBody)
			}

			if contentType := res.Header().Get("Content-Type"); contentType != "application/json" {
				t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
			}
		})
	}
}

func TestAuthRouteErrorsRetainJSONNoStoreAndAllowSemantics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
		wantCode   string
		wantAllow  string
	}{
		{
			name:       "wrong login method",
			method:     http.MethodGet,
			target:     "/api/v1/auth/login",
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   "METHOD_NOT_ALLOWED",
			wantAllow:  http.MethodPost,
		},
		{
			name:       "wrong session method",
			method:     http.MethodDelete,
			target:     "/api/v1/auth/session",
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   "METHOD_NOT_ALLOWED",
			wantAllow:  "GET, HEAD",
		},
		{
			name:       "wrong logout method",
			method:     http.MethodPut,
			target:     "/api/v1/auth/logout",
			wantStatus: http.StatusMethodNotAllowed,
			wantCode:   "METHOD_NOT_ALLOWED",
			wantAllow:  http.MethodPost,
		},
		{
			name:       "unknown auth endpoint",
			method:     http.MethodGet,
			target:     "/api/v1/auth/not-a-route",
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "no browser recovery endpoint",
			method:     http.MethodPost,
			target:     "/api/v1/auth/recovery",
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "no hidden browser recovery authorization endpoint",
			method:     http.MethodPost,
			target:     "/api/v1/auth/recovery/authorize",
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "auth prefix without route",
			method:     http.MethodGet,
			target:     "/api/v1/auth",
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &fakeAuthenticationService{}
			request := httptest.NewRequest(test.method, test.target, nil)
			response := httptest.NewRecorder()

			testAuthRouter(t, service, false).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body.String())
			}
			if got := response.Header().Get("Allow"); got != test.wantAllow {
				t.Fatalf("Allow = %q, want %q", got, test.wantAllow)
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", got)
			}
			assertErrorCode(t, response, test.wantCode)
			assertAuthNoStore(t, response)
			if service.loginCalls != 0 || service.resolveToken != "" || service.logoutToken != "" {
				t.Fatal("invalid auth route must not reach the authentication service")
			}
		})
	}
}
