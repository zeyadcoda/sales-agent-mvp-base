package cache

import (
	"context"
	"errors"
	"strings"

	"github.com/redis/go-redis/v9"
)

// ErrRedisURLRequired is returned when no Redis connection string is supplied.
//
// We fail closed instead of silently connecting to a default Redis instance.
var ErrRedisURLRequired = errors.New("redis URL is required")

// ErrRedisNotInitialized protects callers from trying to use a Redis client
// that was never initialized successfully.
var ErrRedisNotInitialized = errors.New("redis is not initialized")

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

// Close releases resources owned by the Redis client.
func (r *Redis) Close() error {
	if r == nil || r.client == nil {
		return nil
	}

	return r.client.Close()
}
