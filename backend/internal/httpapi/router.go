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
func NewRouter(readiness ReadinessChecker) http.Handler {
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

	return mux
}

// writeJSON provides one small, consistent helper for JSON HTTP responses.
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
