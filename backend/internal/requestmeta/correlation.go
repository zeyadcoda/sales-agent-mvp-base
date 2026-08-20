package requestmeta

import (
	"context"
	"strings"
)

const unavailableCorrelationID = "unavailable"

type correlationIDKey struct{}

// WithCorrelationID attaches a server-controlled diagnostic identifier to a
// context. It deliberately accepts only a compact, single-line value so
// untrusted text can never become audit or log control content.
func WithCorrelationID(ctx context.Context, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 || strings.ContainsAny(value, "\r\n\x00") {
		value = unavailableCorrelationID
	}

	return context.WithValue(ctx, correlationIDKey{}, value)
}

// CorrelationID returns only the server-controlled value installed by
// WithCorrelationID. Callers cannot influence it through HTTP headers or
// recovery CLI arguments.
func CorrelationID(ctx context.Context) string {
	if ctx == nil {
		return unavailableCorrelationID
	}

	value, _ := ctx.Value(correlationIDKey{}).(string)
	if value == "" {
		return unavailableCorrelationID
	}

	return value
}
