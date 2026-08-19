package readiness

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// ErrDependencyMissing means the readiness checker was not given every
// dependency required for this API process.
var ErrDependencyMissing = errors.New("required readiness dependency is missing")

// Pinger is the narrow infrastructure capability needed for readiness.
//
// Keeping this port small prevents callers from acquiring arbitrary
// PostgreSQL or Redis access through the readiness subsystem.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Checker verifies the dependencies required for the API to serve traffic.
type Checker struct {
	postgres Pinger
	redis    Pinger
}

// New constructs a readiness checker with both required dependencies.
func New(postgres Pinger, redis Pinger) *Checker {
	return &Checker{
		postgres: postgres,
		redis:    redis,
	}
}

// Check succeeds only when PostgreSQL and Redis are both healthy.
//
// Missing dependencies fail closed. The HTTP boundary deliberately discards
// the returned error so infrastructure details cannot escape in responses.
func (c *Checker) Check(ctx context.Context) error {
	if c == nil || isNilPinger(c.postgres) || isNilPinger(c.redis) {
		return ErrDependencyMissing
	}

	if err := c.postgres.Ping(ctx); err != nil {
		return fmt.Errorf("check PostgreSQL readiness: %w", err)
	}

	if err := c.redis.Ping(ctx); err != nil {
		return fmt.Errorf("check Redis readiness: %w", err)
	}

	return nil
}

// isNilPinger also detects typed nil pointers stored inside an interface.
// Without this guard, a partially initialized dependency can appear present
// and panic when readiness invokes it.
func isNilPinger(pinger Pinger) bool {
	if pinger == nil {
		return true
	}

	value := reflect.ValueOf(pinger)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
