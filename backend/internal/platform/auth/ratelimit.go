package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"strings"
	"time"
)

const (
	defaultLoginEmailLimit int64 = 5
	defaultLoginIPLimit    int64 = 30
	defaultLoginWindow           = 15 * time.Minute

	loginRateLimitKeyPrefix = "sales-agent:login:{attempts}:"
)

var (
	// ErrRateLimited is intentionally generic across account and network
	// limits so the response cannot reveal whether an account exists.
	ErrRateLimited = errors.New("login rate limit exceeded")

	// ErrRateLimitUnavailable prevents callers from treating a broken security
	// dependency as permission to continue authentication.
	ErrRateLimitUnavailable = errors.New("login rate limiter unavailable")
)

// LoginAttemptCounterStore is implemented by the Redis adapter. It accepts
// only opaque counter keys so plaintext login identifiers never cross into
// Redis keys.
type LoginAttemptCounterStore interface {
	IncrementLoginAttemptCounters(
		ctx context.Context,
		emailCounterKey string,
		ipCounterKey string,
		window time.Duration,
	) (emailAttempts int64, ipAttempts int64, err error)
}

type loginRateLimiter struct {
	store      LoginAttemptCounterStore
	emailLimit int64
	ipLimit    int64
	window     time.Duration
}

// NewLoginRateLimiter returns the fail-closed limiter used by password login.
func NewLoginRateLimiter(store LoginAttemptCounterStore) LoginRateLimiter {
	return &loginRateLimiter{
		store:      store,
		emailLimit: defaultLoginEmailLimit,
		ipLimit:    defaultLoginIPLimit,
		window:     defaultLoginWindow,
	}
}

// Allow consumes one attempt for both the normalized email and requesting IP.
// Correct and incorrect credentials must take this same path to avoid making
// account existence observable through throttling behavior.
func (l *loginRateLimiter) Allow(
	ctx context.Context,
	normalizedEmail string,
	requestingIP string,
) error {
	if l == nil || l.store == nil || l.emailLimit < 1 || l.ipLimit < 1 || l.window < time.Millisecond {
		return ErrRateLimitUnavailable
	}

	email, err := NormalizeEmail(normalizedEmail)
	if err != nil {
		return ErrRateLimitUnavailable
	}

	ip, ok := normalizeRateLimitIP(requestingIP)
	if !ok {
		return ErrRateLimitUnavailable
	}

	emailAttempts, ipAttempts, err := l.store.IncrementLoginAttemptCounters(
		ctx,
		counterKey("email", email),
		counterKey("ip", ip),
		l.window,
	)
	if err != nil || emailAttempts < 1 || ipAttempts < 1 {
		return ErrRateLimitUnavailable
	}

	if emailAttempts > l.emailLimit || ipAttempts > l.ipLimit {
		return ErrRateLimited
	}

	return nil
}

func normalizeRateLimitIP(value string) (string, bool) {
	parsed := net.ParseIP(strings.TrimSpace(value))
	if parsed == nil {
		return "", false
	}

	return parsed.String(), true
}

func counterKey(dimension string, identity string) string {
	digest := sha256.Sum256([]byte(identity))
	return loginRateLimitKeyPrefix + dimension + ":" + hex.EncodeToString(digest[:])
}
