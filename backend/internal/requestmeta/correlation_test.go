package requestmeta

import (
	"context"
	"strings"
	"testing"
)

func TestCorrelationIDRoundTripAndSafeFallback(t *testing.T) {
	t.Parallel()

	if got := CorrelationID(WithCorrelationID(context.Background(), "server-id")); got != "server-id" {
		t.Fatalf("CorrelationID() = %q, want server-id", got)
	}

	for _, value := range []string{"", "bad\nvalue", strings.Repeat("a", 129)} {
		if got := CorrelationID(WithCorrelationID(context.Background(), value)); got != unavailableCorrelationID {
			t.Fatalf("CorrelationID(%q) = %q, want safe fallback", value, got)
		}
	}

	if got := CorrelationID(nil); got != unavailableCorrelationID {
		t.Fatalf("CorrelationID(nil) = %q, want safe fallback", got)
	}
}
