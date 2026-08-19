package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

type counterCall struct {
	emailKey string
	ipKey    string
	window   time.Duration
}

type memoryLoginCounterStore struct {
	mu sync.Mutex

	emailCounts map[string]int64
	ipCounts    map[string]int64
	calls       []counterCall
	err         error
	returnZero  bool
}

func newMemoryLoginCounterStore() *memoryLoginCounterStore {
	return &memoryLoginCounterStore{
		emailCounts: make(map[string]int64),
		ipCounts:    make(map[string]int64),
	}
}

func (s *memoryLoginCounterStore) IncrementLoginAttemptCounters(
	_ context.Context,
	emailCounterKey string,
	ipCounterKey string,
	window time.Duration,
) (int64, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, counterCall{
		emailKey: emailCounterKey,
		ipKey:    ipCounterKey,
		window:   window,
	})

	if s.err != nil {
		return 0, 0, s.err
	}
	if s.returnZero {
		return 0, 0, nil
	}

	s.emailCounts[emailCounterKey]++
	s.ipCounts[ipCounterKey]++

	return s.emailCounts[emailCounterKey], s.ipCounts[ipCounterKey], nil
}

func TestLoginRateLimiterRepeatedAttemptsThrottle(t *testing.T) {
	t.Parallel()

	limiter := NewLoginRateLimiter(newMemoryLoginCounterStore())

	for attempt := int64(1); attempt <= defaultLoginEmailLimit; attempt++ {
		if err := limiter.Allow(context.Background(), "admin@example.com", "192.0.2.10"); err != nil {
			t.Fatalf("attempt %d: Allow() error = %v, want nil", attempt, err)
		}
	}

	err := limiter.Allow(context.Background(), "admin@example.com", "192.0.2.10")
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("attempt after limit: Allow() error = %v, want %v", err, ErrRateLimited)
	}
}

func TestLoginRateLimiterNormalizesAndHashesCounterKeys(t *testing.T) {
	t.Parallel()

	store := newMemoryLoginCounterStore()
	limiter := NewLoginRateLimiter(store)

	if err := limiter.Allow(
		context.Background(),
		"  Admin@Example.COM  ",
		"2001:0db8:0:0:0:0:0:1",
	); err != nil {
		t.Fatalf("Allow() error = %v, want nil", err)
	}
	if err := limiter.Allow(
		context.Background(),
		"admin@example.com",
		"2001:db8::1",
	); err != nil {
		t.Fatalf("Allow() with normalized identifiers error = %v, want nil", err)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if len(store.calls) != 2 {
		t.Fatalf("counter calls = %d, want 2", len(store.calls))
	}

	first := store.calls[0]
	second := store.calls[1]
	if first.emailKey != second.emailKey {
		t.Fatalf("normalized email keys differ: %q and %q", first.emailKey, second.emailKey)
	}
	if first.ipKey != second.ipKey {
		t.Fatalf("canonical IP keys differ: %q and %q", first.ipKey, second.ipKey)
	}

	if first.emailKey != counterKey("email", "admin@example.com") {
		t.Fatalf("email key = %q, want SHA-256-derived key", first.emailKey)
	}
	if first.ipKey != counterKey("ip", "2001:db8::1") {
		t.Fatalf("IP key = %q, want SHA-256-derived key", first.ipKey)
	}

	for _, plaintext := range []string{"admin@example.com", "example.com", "2001:db8::1"} {
		if strings.Contains(first.emailKey, plaintext) || strings.Contains(first.ipKey, plaintext) {
			t.Fatalf("Redis counter key contains plaintext identifier %q", plaintext)
		}
	}

	if first.window != defaultLoginWindow {
		t.Fatalf("counter window = %s, want %s", first.window, defaultLoginWindow)
	}
}

func TestLoginRateLimiterAppliesLayeredLimits(t *testing.T) {
	t.Parallel()

	t.Run("normalized email across IP addresses", func(t *testing.T) {
		t.Parallel()

		limiter := NewLoginRateLimiter(newMemoryLoginCounterStore())
		for attempt := int64(1); attempt <= defaultLoginEmailLimit; attempt++ {
			ip := fmt.Sprintf("192.0.2.%d", attempt)
			if err := limiter.Allow(context.Background(), "admin@example.com", ip); err != nil {
				t.Fatalf("attempt %d: Allow() error = %v, want nil", attempt, err)
			}
		}

		err := limiter.Allow(context.Background(), "admin@example.com", "192.0.2.200")
		if !errors.Is(err, ErrRateLimited) {
			t.Fatalf("email limit error = %v, want %v", err, ErrRateLimited)
		}
	})

	t.Run("requesting IP across email addresses", func(t *testing.T) {
		t.Parallel()

		limiter := NewLoginRateLimiter(newMemoryLoginCounterStore())
		for attempt := int64(1); attempt <= defaultLoginIPLimit; attempt++ {
			email := fmt.Sprintf("admin-%d@example.com", attempt)
			if err := limiter.Allow(context.Background(), email, "198.51.100.40"); err != nil {
				t.Fatalf("attempt %d: Allow() error = %v, want nil", attempt, err)
			}
		}

		err := limiter.Allow(context.Background(), "blocked@example.com", "198.51.100.40")
		if !errors.Is(err, ErrRateLimited) {
			t.Fatalf("IP limit error = %v, want %v", err, ErrRateLimited)
		}
	})
}

func TestLoginRateLimiterFailsClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		limiter LoginRateLimiter
		email   string
		ip      string
	}{
		{
			name:    "missing store",
			limiter: NewLoginRateLimiter(nil),
			email:   "admin@example.com",
			ip:      "192.0.2.1",
		},
		{
			name: "Redis store failure",
			limiter: NewLoginRateLimiter(&memoryLoginCounterStore{
				err: errors.New("simulated Redis failure"),
			}),
			email: "admin@example.com",
			ip:    "192.0.2.1",
		},
		{
			name: "invalid zero counter response",
			limiter: NewLoginRateLimiter(&memoryLoginCounterStore{
				returnZero: true,
			}),
			email: "admin@example.com",
			ip:    "192.0.2.1",
		},
		{
			name:    "empty normalized email",
			limiter: NewLoginRateLimiter(newMemoryLoginCounterStore()),
			email:   "  ",
			ip:      "192.0.2.1",
		},
		{
			name:    "invalid requesting IP",
			limiter: NewLoginRateLimiter(newMemoryLoginCounterStore()),
			email:   "admin@example.com",
			ip:      "not-an-ip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.limiter.Allow(context.Background(), tt.email, tt.ip)
			if !errors.Is(err, ErrRateLimitUnavailable) {
				t.Fatalf("Allow() error = %v, want %v", err, ErrRateLimitUnavailable)
			}
			if errors.Is(err, ErrRateLimited) {
				t.Fatalf("Allow() error = %v; dependency failure must not appear as ordinary throttling", err)
			}
		})
	}
}
