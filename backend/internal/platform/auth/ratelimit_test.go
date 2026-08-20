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

type memoryOTPAttemptCounterStore struct {
	mu sync.Mutex

	challengeCounts map[string]int64
	ipCounts        map[string]int64
	calls           []counterCall
	err             error
	returnZero      bool
}

func newMemoryOTPAttemptCounterStore() *memoryOTPAttemptCounterStore {
	return &memoryOTPAttemptCounterStore{
		challengeCounts: make(map[string]int64),
		ipCounts:        make(map[string]int64),
	}
}

func (store *memoryOTPAttemptCounterStore) IncrementOTPAttemptCounters(
	_ context.Context,
	challengeCounterKey string,
	ipCounterKey string,
	window time.Duration,
) (int64, int64, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.calls = append(store.calls, counterCall{
		emailKey: challengeCounterKey,
		ipKey:    ipCounterKey,
		window:   window,
	})
	if store.err != nil {
		return 0, 0, store.err
	}
	if store.returnZero {
		return 0, 0, nil
	}

	store.challengeCounts[challengeCounterKey]++
	store.ipCounts[ipCounterKey]++
	return store.challengeCounts[challengeCounterKey], store.ipCounts[ipCounterKey], nil
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

func TestOTPRateLimiterThrottlesVerifyAndResendSeparately(t *testing.T) {
	t.Parallel()

	store := newMemoryOTPAttemptCounterStore()
	limiter := NewOTPRateLimiter(store)
	challengeID := tokenFromByte(31)

	for attempt := int64(1); attempt <= defaultOTPChallengeLimit; attempt++ {
		if err := limiter.AllowVerify(context.Background(), challengeID, "192.0.2.10"); err != nil {
			t.Fatalf("verify attempt %d: error = %v", attempt, err)
		}
	}
	if err := limiter.AllowVerify(context.Background(), challengeID, "192.0.2.10"); !errors.Is(err, ErrOTPVerifyRateLimited) {
		t.Fatalf("verify over limit error = %v, want %v", err, ErrOTPVerifyRateLimited)
	}

	// Resend has a distinct keyspace, so verification traffic cannot consume
	// its allowance (PostgreSQL still independently enforces the cooldown).
	for attempt := int64(1); attempt <= defaultOTPChallengeLimit; attempt++ {
		if err := limiter.AllowResend(context.Background(), challengeID, "192.0.2.10"); err != nil {
			t.Fatalf("resend attempt %d: error = %v", attempt, err)
		}
	}
	if err := limiter.AllowResend(context.Background(), challengeID, "192.0.2.10"); !errors.Is(err, ErrOTPResendRateLimited) {
		t.Fatalf("resend over limit error = %v, want %v", err, ErrOTPResendRateLimited)
	}
}

func TestOTPRateLimiterUsesOpaqueLayeredKeys(t *testing.T) {
	t.Parallel()

	store := newMemoryOTPAttemptCounterStore()
	limiter := NewOTPRateLimiter(store)
	challengeID := tokenFromByte(32)
	requestingIP := "2001:db8::20"

	if err := limiter.AllowVerify(context.Background(), challengeID, requestingIP); err != nil {
		t.Fatalf("AllowVerify() error = %v", err)
	}
	if err := limiter.AllowResend(context.Background(), challengeID, requestingIP); err != nil {
		t.Fatalf("AllowResend() error = %v", err)
	}

	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.calls) != 2 {
		t.Fatalf("counter calls = %d, want 2", len(store.calls))
	}
	verifyCall, resendCall := store.calls[0], store.calls[1]
	if verifyCall.emailKey == resendCall.emailKey || verifyCall.ipKey == resendCall.ipKey {
		t.Fatal("verify and resend counter keys are not operation-separated")
	}
	for _, call := range store.calls {
		if strings.Contains(call.emailKey, challengeID) || strings.Contains(call.ipKey, requestingIP) {
			t.Fatalf("OTP Redis key contains plaintext identity: %#v", call)
		}
		if call.window != defaultOTPWindow {
			t.Fatalf("OTP counter window = %s, want %s", call.window, defaultOTPWindow)
		}
	}
}

func TestOTPRateLimiterFailsClosed(t *testing.T) {
	t.Parallel()

	challengeID := tokenFromByte(33)
	tests := []struct {
		name      string
		limiter   OTPRateLimiter
		challenge string
		ip        string
	}{
		{name: "missing store", limiter: NewOTPRateLimiter(nil), challenge: challengeID, ip: "192.0.2.1"},
		{name: "Redis failure", limiter: NewOTPRateLimiter(&memoryOTPAttemptCounterStore{err: errors.New("Redis failure")}), challenge: challengeID, ip: "192.0.2.1"},
		{name: "zero counter", limiter: NewOTPRateLimiter(&memoryOTPAttemptCounterStore{returnZero: true}), challenge: challengeID, ip: "192.0.2.1"},
		{name: "malformed challenge", limiter: NewOTPRateLimiter(newMemoryOTPAttemptCounterStore()), challenge: "bad", ip: "192.0.2.1"},
		{name: "malformed IP", limiter: NewOTPRateLimiter(newMemoryOTPAttemptCounterStore()), challenge: challengeID, ip: "bad"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.limiter.AllowVerify(context.Background(), test.challenge, test.ip); !errors.Is(err, ErrRateLimitUnavailable) {
				t.Fatalf("AllowVerify() error = %v, want %v", err, ErrRateLimitUnavailable)
			}
			if err := test.limiter.AllowResend(context.Background(), test.challenge, test.ip); !errors.Is(err, ErrRateLimitUnavailable) {
				t.Fatalf("AllowResend() error = %v, want %v", err, ErrRateLimitUnavailable)
			}
		})
	}
}
