package cache

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

// TestOpenRejectsMissingRedisURL verifies that Redis configuration fails
// closed when no explicit Redis URL is supplied.
func TestOpenRejectsMissingRedisURL(t *testing.T) {
	t.Parallel()

	client, err := Open("")

	if client != nil {
		t.Fatal("Redis client should be nil when the Redis URL is missing")
	}

	if !errors.Is(err, ErrRedisURLRequired) {
		t.Fatalf("error = %v, want %v", err, ErrRedisURLRequired)
	}
}

// TestOpenEnablesContextTimeouts protects the readiness request deadline from
// being replaced by go-redis socket timeout and retry defaults.
func TestOpenEnablesContextTimeouts(t *testing.T) {
	t.Parallel()

	client, err := Open("redis://127.0.0.1:6379/0")
	if err != nil {
		t.Fatalf("Open() returned unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Redis client: %v", err)
		}
	})

	if !client.client.Options().ContextTimeoutEnabled {
		t.Fatal("Redis command context deadlines must be enabled")
	}
}

// TestRedisPing proves that the backend can communicate with a real Redis
// instance.
//
// The integration test runs only when TEST_REDIS_URL is explicitly supplied.
// This prevents normal test runs from accidentally contacting an unintended
// Redis service.
func TestRedisPing(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")

	if redisURL == "" {
		t.Skip("TEST_REDIS_URL is not set; skipping Redis integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := Open(redisURL)
	if err != nil {
		t.Fatalf("open Redis client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Redis client: %v", err)
		}
	}()

	if err := client.Ping(ctx); err != nil {
		t.Fatalf("ping Redis: %v", err)
	}
}
