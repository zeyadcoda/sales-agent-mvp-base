package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakeReadinessChecker lets HTTP tests simulate a healthy or failed database
// without connecting to real infrastructure.
type fakeReadinessChecker struct {
	err error
}

func (f fakeReadinessChecker) Ping(_ context.Context) error {
	return f.err
}

// TestLiveness proves that the API reports itself alive independently of
// database availability.
func TestLiveness(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	res := httptest.NewRecorder()

	NewRouter(nil).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
}

// TestReadinessFailsClosedWithoutDependency proves that the application never
// reports READY when its required dependency was not configured.
func TestReadinessFailsClosedWithoutDependency(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	res := httptest.NewRecorder()

	NewRouter(nil).ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"status = %d, want %d",
			res.Code,
			http.StatusServiceUnavailable,
		)
	}
}

// TestReadinessSucceedsWhenDatabaseIsHealthy proves that a healthy database
// allows the API to advertise readiness.
func TestReadinessSucceedsWhenDatabaseIsHealthy(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	res := httptest.NewRecorder()

	NewRouter(fakeReadinessChecker{}).ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
}

// TestReadinessFailsWhenDatabaseIsUnavailable proves that infrastructure
// failure is surfaced as NOT READY rather than a false healthy status.
func TestReadinessFailsWhenDatabaseIsUnavailable(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	res := httptest.NewRecorder()

	checker := fakeReadinessChecker{
		err: errors.New("simulated database failure"),
	}

	NewRouter(checker).ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"status = %d, want %d",
			res.Code,
			http.StatusServiceUnavailable,
		)
	}
}
