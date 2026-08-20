package cache

import (
	"context"
	"errors"
	"fmt"
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

func TestIncrementLoginAttemptCountersRejectsMissingClient(t *testing.T) {
	t.Parallel()

	var client *Redis

	_, _, err := client.IncrementLoginAttemptCounters(
		context.Background(),
		"opaque-email-key",
		"opaque-ip-key",
		15*time.Minute,
	)
	if !errors.Is(err, ErrRedisNotInitialized) {
		t.Fatalf("error = %v, want %v", err, ErrRedisNotInitialized)
	}
}

func TestIncrementOTPAttemptCountersRejectsMissingClient(t *testing.T) {
	t.Parallel()

	var client *Redis

	_, _, err := client.IncrementOTPAttemptCounters(
		context.Background(),
		"opaque-challenge-key",
		"opaque-ip-key",
		15*time.Minute,
	)
	if !errors.Is(err, ErrRedisNotInitialized) {
		t.Fatalf("error = %v, want %v", err, ErrRedisNotInitialized)
	}
}

func TestIncrementLoginAttemptCountersRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client, err := Open("redis://127.0.0.1:6379/0")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Redis client: %v", err)
		}
	})

	tests := []struct {
		name     string
		emailKey string
		ipKey    string
		window   time.Duration
	}{
		{name: "empty email key", ipKey: "opaque-ip-key", window: time.Minute},
		{name: "empty IP key", emailKey: "opaque-email-key", window: time.Minute},
		{name: "same key", emailKey: "opaque-key", ipKey: "opaque-key", window: time.Minute},
		{name: "zero window", emailKey: "opaque-email-key", ipKey: "opaque-ip-key"},
		{name: "sub-millisecond window", emailKey: "opaque-email-key", ipKey: "opaque-ip-key", window: time.Nanosecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := client.IncrementLoginAttemptCounters(
				context.Background(),
				tt.emailKey,
				tt.ipKey,
				tt.window,
			)
			if !errors.Is(err, ErrInvalidRateLimitCounter) {
				t.Fatalf("error = %v, want %v", err, ErrInvalidRateLimitCounter)
			}
		})
	}
}

func TestIncrementLoginAttemptCountersReturnsRedisError(t *testing.T) {
	t.Parallel()

	client, err := Open("redis://127.0.0.1:6379/0")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Redis client: %v", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	emailAttempts, ipAttempts, err := client.IncrementLoginAttemptCounters(
		ctx,
		"opaque-email-key",
		"opaque-ip-key",
		15*time.Minute,
	)
	if err == nil {
		t.Fatal("expected Redis failure, got nil")
	}
	if emailAttempts != 0 || ipAttempts != 0 {
		t.Fatalf("attempts = (%d, %d), want (0, 0) on Redis failure", emailAttempts, ipAttempts)
	}
}

func TestIncrementOTPAttemptCountersReturnsRedisError(t *testing.T) {
	t.Parallel()

	client, err := Open("redis://127.0.0.1:6379/0")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Redis client: %v", err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	challengeAttempts, ipAttempts, err := client.IncrementOTPAttemptCounters(
		ctx,
		"opaque-challenge-key",
		"opaque-ip-key",
		15*time.Minute,
	)
	if err == nil {
		t.Fatal("expected Redis failure, got nil")
	}
	if challengeAttempts != 0 || ipAttempts != 0 {
		t.Fatalf("attempts = (%d, %d), want (0, 0) on Redis failure", challengeAttempts, ipAttempts)
	}
}

func TestIncrementLoginAttemptCountersWithRedis(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("TEST_REDIS_URL is not set; skipping Redis rate-limit integration test")
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

	unique := time.Now().UnixNano()
	emailKey := fmt.Sprintf("sales-agent:test:{attempts}:email:%d", unique)
	ipKey := fmt.Sprintf("sales-agent:test:{attempts}:ip:%d", unique)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Second)
		defer cleanupCancel()
		_ = client.client.Del(cleanupCtx, emailKey, ipKey).Err()
	})

	window := 15 * time.Minute
	for want := int64(1); want <= 2; want++ {
		emailAttempts, ipAttempts, err := client.IncrementLoginAttemptCounters(
			ctx,
			emailKey,
			ipKey,
			window,
		)
		if err != nil {
			t.Fatalf("increment %d: %v", want, err)
		}
		if emailAttempts != want || ipAttempts != want {
			t.Fatalf(
				"increment %d returned (%d, %d), want (%d, %d)",
				want,
				emailAttempts,
				ipAttempts,
				want,
				want,
			)
		}
	}

	for _, key := range []string{emailKey, ipKey} {
		ttl, err := client.client.PTTL(ctx, key).Result()
		if err != nil {
			t.Fatalf("read TTL for %q: %v", key, err)
		}
		if ttl <= 0 || ttl > window {
			t.Fatalf("TTL for %q = %s, want within (0, %s]", key, ttl, window)
		}
	}
}

func TestIncrementOTPAttemptCountersWithRedis(t *testing.T) {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("TEST_REDIS_URL is not set; skipping Redis OTP rate-limit integration test")
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

	unique := time.Now().UnixNano()
	challengeKey := fmt.Sprintf("sales-agent:test:{attempts}:otp-challenge:%d", unique)
	ipKey := fmt.Sprintf("sales-agent:test:{attempts}:otp-ip:%d", unique)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Second)
		defer cleanupCancel()
		_ = client.client.Del(cleanupCtx, challengeKey, ipKey).Err()
	})

	window := 15 * time.Minute
	for want := int64(1); want <= 2; want++ {
		challengeAttempts, ipAttempts, err := client.IncrementOTPAttemptCounters(
			ctx,
			challengeKey,
			ipKey,
			window,
		)
		if err != nil {
			t.Fatalf("increment %d: %v", want, err)
		}
		if challengeAttempts != want || ipAttempts != want {
			t.Fatalf(
				"increment %d returned (%d, %d), want (%d, %d)",
				want,
				challengeAttempts,
				ipAttempts,
				want,
				want,
			)
		}
	}

	for _, key := range []string{challengeKey, ipKey} {
		ttl, err := client.client.PTTL(ctx, key).Result()
		if err != nil {
			t.Fatalf("read TTL for %q: %v", key, err)
		}
		if ttl <= 0 || ttl > window {
			t.Fatalf("TTL for %q = %s, want within (0, %s]", key, ttl, window)
		}
	}
}
