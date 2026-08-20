package cache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrRedisURLRequired is returned when no Redis connection string is supplied.
//
// We fail closed instead of silently connecting to a default Redis instance.
var ErrRedisURLRequired = errors.New("redis URL is required")

// ErrRedisNotInitialized protects callers from trying to use a Redis client
// that was never initialized successfully.
var ErrRedisNotInitialized = errors.New("redis is not initialized")

// ErrInvalidRateLimitCounter protects the Redis boundary from creating
// permanent or ambiguously named security counters.
var ErrInvalidRateLimitCounter = errors.New("invalid rate limit counter input")

// incrementAttemptCountersScript updates both abuse-control dimensions in one
// Redis operation. The expiry is set only when a fixed window begins; later
// attempts cannot extend the window indefinitely.
var incrementAttemptCountersScript = redis.NewScript(`
local function increment_with_expiry(key, ttl_ms)
  local count = redis.call("INCR", key)
  if count == 1 or redis.call("PTTL", key) < 0 then
    redis.call("PEXPIRE", key, ttl_ms)
  end
  return count
end

return {
  increment_with_expiry(KEYS[1], ARGV[1]),
  increment_with_expiry(KEYS[2], ARGV[1])
}
`)

// Redis owns the Redis client used by the application.
//
// The client is private so other packages do not bypass the application
// boundary and start using Redis as an uncontrolled business-data store.
// PostgreSQL remains the authoritative source of business state.
type Redis struct {
	client *redis.Client
}

// Open creates a Redis client from the supplied Redis URL.
//
// Creating the client does not prove Redis is reachable. Ping performs the
// actual connectivity check used by readiness and integration tests.
func Open(redisURL string) (*Redis, error) {
	if strings.TrimSpace(redisURL) == "" {
		return nil, ErrRedisURLRequired
	}

	// ParseURL validates the Redis connection string.
	//
	// We intentionally do not include the raw URL in returned errors because
	// production Redis URLs may contain credentials.
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, errors.New("invalid Redis connection configuration")
	}

	// Readiness supplies a strict request deadline. go-redis ignores command
	// context deadlines unless this option is enabled, which could otherwise
	// leave a readiness request waiting on longer socket timeouts and retries.
	options.ContextTimeoutEnabled = true

	return &Redis{
		client: redis.NewClient(options),
	}, nil
}

// Ping verifies that Redis is currently reachable.
//
// Readiness checks use this method so the application reports NOT READY when a
// required Redis dependency is unavailable.
func (r *Redis) Ping(ctx context.Context) error {
	if r == nil || r.client == nil {
		return ErrRedisNotInitialized
	}

	return r.client.Ping(ctx).Err()
}

// IncrementLoginAttemptCounters atomically consumes one attempt from both
// login-rate-limit dimensions. Keys must already be opaque identifiers; the
// cache layer deliberately never receives the email address or IP address.
func (r *Redis) IncrementLoginAttemptCounters(
	ctx context.Context,
	emailCounterKey string,
	ipCounterKey string,
	window time.Duration,
) (emailAttempts int64, ipAttempts int64, err error) {
	return r.incrementAttemptCounters(ctx, emailCounterKey, ipCounterKey, window)
}

// IncrementOTPAttemptCounters atomically consumes an attempt from both the
// opaque challenge and requesting-network dimensions. PostgreSQL still owns
// OTP lifecycle and the five-failure lock; Redis adds fail-closed abuse
// protection without becoming authentication state.
func (r *Redis) IncrementOTPAttemptCounters(
	ctx context.Context,
	challengeCounterKey string,
	ipCounterKey string,
	window time.Duration,
) (challengeAttempts int64, ipAttempts int64, err error) {
	return r.incrementAttemptCounters(ctx, challengeCounterKey, ipCounterKey, window)
}

func (r *Redis) incrementAttemptCounters(
	ctx context.Context,
	firstCounterKey string,
	secondCounterKey string,
	window time.Duration,
) (firstAttempts int64, secondAttempts int64, err error) {
	if r == nil || r.client == nil {
		return 0, 0, ErrRedisNotInitialized
	}

	if strings.TrimSpace(firstCounterKey) == "" ||
		strings.TrimSpace(secondCounterKey) == "" ||
		firstCounterKey == secondCounterKey ||
		window < time.Millisecond {
		return 0, 0, ErrInvalidRateLimitCounter
	}

	result, err := incrementAttemptCountersScript.Run(
		ctx,
		r.client,
		[]string{firstCounterKey, secondCounterKey},
		window.Milliseconds(),
	).Slice()
	if err != nil {
		return 0, 0, err
	}

	if len(result) != 2 {
		return 0, 0, fmt.Errorf("unexpected rate limit counter response")
	}

	firstAttempts, firstOK := result[0].(int64)
	secondAttempts, secondOK := result[1].(int64)
	if !firstOK || !secondOK || firstAttempts < 1 || secondAttempts < 1 {
		return 0, 0, fmt.Errorf("invalid rate limit counter response")
	}

	return firstAttempts, secondAttempts, nil
}

// Close releases resources owned by the Redis client.
func (r *Redis) Close() error {
	if r == nil || r.client == nil {
		return nil
	}

	return r.client.Close()
}
