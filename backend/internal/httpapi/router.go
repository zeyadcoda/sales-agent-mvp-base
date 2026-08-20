package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// ReadinessChecker defines the minimum capability required by the API
// readiness endpoint.
//
// Keeping the interface small prevents the HTTP layer from gaining direct
// PostgreSQL or Redis access.
type ReadinessChecker interface {
	Check(ctx context.Context) error
}

type healthResponse struct {
	Status string `json:"status"`
}

// NewRouter creates the HTTP routes exposed by the API.
//
// The readiness dependency is injected instead of created inside the router.
// This keeps infrastructure concerns separate and makes failure cases easy to
// test without requiring real infrastructure for every HTTP unit test.
func NewRouter(
	readiness ReadinessChecker,
	authHandler *AuthHandler,
	dashboardHandler *DashboardHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Liveness answers only one question: is this Go process alive?
	// It deliberately does not depend on PostgreSQL, Redis, or other providers.
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "alive"})
	})

	// Readiness answers a different question: can this process currently serve
	// requests that depend on its required infrastructure?
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		// Missing dependencies fail closed. We never report READY just because
		// the Go process itself happens to be running.
		if readiness == nil {
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "not_ready"})
			return
		}

		// Bound all dependency checks with one deadline so a slow dependency
		// cannot leave readiness requests hanging indefinitely.
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := readiness.Check(ctx); err != nil {
			// Do not return raw dependency errors to the caller because they may
			// reveal internal infrastructure details or credentials.
			writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "not_ready"})
			return
		}

		writeJSON(w, http.StatusOK, healthResponse{Status: "ready"})
	})

	if authHandler != nil {
		mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
		mux.HandleFunc("POST /api/v1/auth/otp/verify", authHandler.VerifyOTP)
		mux.HandleFunc("POST /api/v1/auth/otp/resend", authHandler.ResendOTP)
		mux.HandleFunc("POST /api/v1/auth/otp/status", authHandler.OTPChallengeStatus)
		mux.HandleFunc("GET /api/v1/auth/session", authHandler.Session)
		mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

		// Method-specific patterns above remain authoritative for valid requests.
		// This narrower API fallback replaces ServeMux's plain-text errors so all
		// authentication responses retain the JSON and no-store contract.
		mux.HandleFunc("/api/v1/auth", authRouteFallback)
		mux.HandleFunc("/api/v1/auth/", authRouteFallback)
	}

	if dashboardHandler != nil {
		mux.HandleFunc("GET /api/v1/admin/dashboard", dashboardHandler.Dashboard)

		// Preserve JSON, correlation, and no-store semantics for wrong methods and
		// unknown paths within the protected admin API namespace.
		mux.HandleFunc("/api/v1/admin", adminRouteFallback)
		mux.HandleFunc("/api/v1/admin/", adminRouteFallback)
	}

	return withRequestMetadata(mux)
}

func adminRouteFallback(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)

	if r.URL.Path == "/api/v1/admin/dashboard" {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		writeAPIError(
			w,
			r,
			http.StatusMethodNotAllowed,
			"METHOD_NOT_ALLOWED",
			"The requested method is not allowed for this endpoint.",
			nil,
		)
		return
	}

	writeAPIError(
		w,
		r,
		http.StatusNotFound,
		"NOT_FOUND",
		"The requested endpoint was not found.",
		nil,
	)
}

func authRouteFallback(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)

	var allowed string
	switch r.URL.Path {
	case "/api/v1/auth/login",
		"/api/v1/auth/logout",
		"/api/v1/auth/otp/verify",
		"/api/v1/auth/otp/resend",
		"/api/v1/auth/otp/status":
		allowed = http.MethodPost
	case "/api/v1/auth/session":
		allowed = http.MethodGet + ", " + http.MethodHead
	}

	if allowed != "" {
		w.Header().Set("Allow", allowed)
		writeAPIError(
			w,
			r,
			http.StatusMethodNotAllowed,
			"METHOD_NOT_ALLOWED",
			"The requested method is not allowed for this endpoint.",
			nil,
		)
		return
	}

	writeAPIError(
		w,
		r,
		http.StatusNotFound,
		"NOT_FOUND",
		"The requested endpoint was not found.",
		nil,
	)
}

// writeJSON provides one small, consistent helper for JSON HTTP responses.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
