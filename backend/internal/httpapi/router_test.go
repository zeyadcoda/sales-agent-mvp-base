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

	NewRouter(nil).ServeHTTP(res, req)

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

	NewRouter(panicReadinessChecker{}).ServeHTTP(res, req)

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

			NewRouter(tt.readiness).ServeHTTP(res, req)

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
