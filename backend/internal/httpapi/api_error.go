package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"salesagent.local/backend/internal/requestmeta"
)

type fieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type apiError struct {
	Code          string       `json:"code"`
	Message       string       `json:"message"`
	CorrelationID string       `json:"correlation_id"`
	FieldErrors   []fieldError `json:"field_errors"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

// withRequestMetadata creates a server-controlled correlation identifier and
// baseline API security headers. Client-supplied IDs are not trusted as log or
// response control data.
func withRequestMetadata(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := newCorrelationID()
		ctx := requestmeta.WithCorrelationID(r.Context(), correlationID)

		w.Header().Set("X-Correlation-ID", correlationID)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newCorrelationID() string {
	material := make([]byte, 16)
	if _, err := rand.Read(material); err != nil {
		// This identifier is diagnostic rather than an authentication secret.
		// A stable fallback preserves the non-empty error contract if the OS
		// random source is unavailable while auth token generation still fails.
		return "unavailable"
	}

	return hex.EncodeToString(material)
}

func correlationID(ctx context.Context) string {
	return requestmeta.CorrelationID(ctx)
}

func writeAPIError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	code string,
	message string,
	fields []fieldError,
) {
	if fields == nil {
		fields = []fieldError{}
	}

	writeJSON(w, status, errorEnvelope{Error: apiError{
		Code:          code,
		Message:       message,
		CorrelationID: correlationID(r.Context()),
		FieldErrors:   fields,
	}})
}
