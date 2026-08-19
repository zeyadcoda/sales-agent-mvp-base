package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	admin            SuperAdmin
	adminErr         error
	createdSession   NewSession
	createSessionErr error
	session          Session
	sessionErr       error
	lookupTokenHash  []byte
	touchedSessionID string
	touchedAt        time.Time
	touchErr         error
	revokedSessionID string
	revokeErr        error
}

func (store *fakeStore) FindSuperAdminByEmail(_ context.Context, _ string) (SuperAdmin, error) {
	return store.admin, store.adminErr
}

func (store *fakeStore) CreateSession(_ context.Context, session NewSession) error {
	store.createdSession = session
	return store.createSessionErr
}

func (store *fakeStore) FindSessionByTokenHash(_ context.Context, tokenHash []byte) (Session, error) {
	store.lookupTokenHash = append([]byte(nil), tokenHash...)
	return store.session, store.sessionErr
}

func (store *fakeStore) TouchSession(_ context.Context, sessionID string, seenAt time.Time) error {
	store.touchedSessionID = sessionID
	store.touchedAt = seenAt
	return store.touchErr
}

func (store *fakeStore) RevokeSession(_ context.Context, sessionID string, _ time.Time) error {
	store.revokedSessionID = sessionID
	return store.revokeErr
}

type fakeLimiter struct {
	err   error
	calls int
	email string
	ip    string
}

func (limiter *fakeLimiter) Allow(_ context.Context, email string, ip string) error {
	limiter.calls++
	limiter.email = email
	limiter.ip = ip
	return limiter.err
}

type fakePasswordVerifier struct {
	results map[string]bool
	err     error
	hashes  []string
}

func (verifier *fakePasswordVerifier) Verify(hash string, _ string) (bool, error) {
	verifier.hashes = append(verifier.hashes, hash)
	if verifier.err != nil {
		return false, verifier.err
	}

	return verifier.results[hash], nil
}

func TestLoginUnknownEmailAndWrongPasswordReturnSameFailure(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		store    *fakeStore
		verifier *fakePasswordVerifier
		wantHash string
	}{
		{
			name:     "unknown email",
			store:    &fakeStore{adminErr: ErrSuperAdminNotFound},
			verifier: &fakePasswordVerifier{results: map[string]bool{}},
			wantHash: "dummy-hash",
		},
		{
			name: "wrong password",
			store: &fakeStore{admin: SuperAdmin{
				ID:           "00000000-0000-0000-0000-000000000001",
				Email:        "admin@example.com",
				PasswordHash: "stored-hash",
				IsActive:     true,
			}},
			verifier: &fakePasswordVerifier{results: map[string]bool{"stored-hash": false}},
			wantHash: "stored-hash",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			limiter := &fakeLimiter{}
			service := newTestService(t, test.store, limiter, test.verifier, true, now)

			_, _, err := service.Login(
				context.Background(),
				" ADMIN@example.com ",
				"incorrect password",
				"192.0.2.10",
			)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
			}
			if len(test.verifier.hashes) != 1 || test.verifier.hashes[0] != test.wantHash {
				t.Fatalf("verified hashes = %v, want [%s]", test.verifier.hashes, test.wantHash)
			}
			if test.store.createdSession.SuperAdminID != "" {
				t.Fatal("invalid credentials must not create a session")
			}
			if limiter.email != "admin@example.com" || limiter.ip != "192.0.2.10" {
				t.Fatalf("limiter identity = (%q, %q)", limiter.email, limiter.ip)
			}
		})
	}
}

func TestLoginLocalBypassCreatesIndependentHashedSession(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	store := &fakeStore{admin: activeAdmin()}
	verifier := &fakePasswordVerifier{results: map[string]bool{"stored-hash": true}}
	service := newTestService(t, store, &fakeLimiter{}, verifier, true, now)

	session, rawToken, err := service.Login(
		context.Background(),
		"admin@example.com",
		"correct password",
		"192.0.2.10",
	)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !validOpaqueToken(rawToken) {
		t.Fatalf("raw session token is not a canonical 256-bit token: %q", rawToken)
	}
	if session.CSRFToken == rawToken || !validOpaqueToken(session.CSRFToken) {
		t.Fatal("CSRF token must be an independent 256-bit value")
	}
	if store.createdSession.SuperAdminID != activeAdmin().ID {
		t.Fatalf("session admin ID = %q", store.createdSession.SuperAdminID)
	}
	wantHash := sha256.Sum256([]byte(rawToken))
	if !bytes.Equal(store.createdSession.TokenHash, wantHash[:]) {
		t.Fatalf("stored token hash = %x, want %x", store.createdSession.TokenHash, wantHash)
	}
	if bytes.Equal(store.createdSession.TokenHash, []byte(rawToken)) {
		t.Fatal("raw token must never be stored")
	}
	if session.SuperAdmin.PasswordHash != "" {
		t.Fatal("password hash must not leave the auth service")
	}
	if !session.ExpiresAt.Equal(now.Add(8 * time.Hour)) {
		t.Fatalf("expiry = %v, want %v", session.ExpiresAt, now.Add(8*time.Hour))
	}
}

func TestLoginBypassDisabledNeverCreatesSession(t *testing.T) {
	t.Parallel()

	store := &fakeStore{admin: activeAdmin()}
	service := newTestService(
		t,
		store,
		&fakeLimiter{},
		&fakePasswordVerifier{results: map[string]bool{"stored-hash": true}},
		false,
		time.Now(),
	)

	_, token, err := service.Login(
		context.Background(),
		"admin@example.com",
		"correct password",
		"192.0.2.10",
	)
	if !errors.Is(err, ErrOTPRequired) {
		t.Fatalf("Login() error = %v, want %v", err, ErrOTPRequired)
	}
	if token != "" || store.createdSession.SuperAdminID != "" {
		t.Fatal("password authentication without OTP bypass must not create a session")
	}
}

func TestInactiveSuperAdminCannotLogin(t *testing.T) {
	t.Parallel()

	admin := activeAdmin()
	admin.IsActive = false
	store := &fakeStore{admin: admin}
	service := newTestService(
		t,
		store,
		&fakeLimiter{},
		&fakePasswordVerifier{results: map[string]bool{"stored-hash": true}},
		true,
		time.Now(),
	)

	_, _, err := service.Login(
		context.Background(),
		"admin@example.com",
		"correct password",
		"192.0.2.10",
	)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want generic credentials error", err)
	}
	if store.createdSession.SuperAdminID != "" {
		t.Fatal("inactive account must not create a session")
	}
}

func TestRedisFailureStopsAuthenticationBeforeAccountLookup(t *testing.T) {
	t.Parallel()

	store := &fakeStore{admin: activeAdmin()}
	service := newTestService(
		t,
		store,
		&fakeLimiter{err: ErrRateLimitUnavailable},
		&fakePasswordVerifier{results: map[string]bool{"stored-hash": true}},
		true,
		time.Now(),
	)

	_, _, err := service.Login(
		context.Background(),
		"admin@example.com",
		"correct password",
		"192.0.2.10",
	)
	if !errors.Is(err, ErrAuthenticationUnavailable) {
		t.Fatalf("Login() error = %v, want %v", err, ErrAuthenticationUnavailable)
	}
	if store.createdSession.SuperAdminID != "" {
		t.Fatal("rate limiter failure must not create a session")
	}
}

func TestResolveSessionRejectsInvalidExpiredRevokedAndInactiveSessions(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	rawToken := tokenFromByte(7)
	revokedAt := now.Add(-time.Minute)

	tests := []struct {
		name     string
		token    string
		session  Session
		storeErr error
	}{
		{name: "malformed token", token: "not-a-session-token"},
		{name: "unknown token", token: rawToken, storeErr: ErrSessionNotFound},
		{name: "expired", token: rawToken, session: validSession(now, rawToken, now)},
		{name: "revoked", token: rawToken, session: func() Session {
			session := validSession(now, rawToken, now.Add(time.Hour))
			session.RevokedAt = &revokedAt
			return session
		}()},
		{name: "inactive admin", token: rawToken, session: func() Session {
			session := validSession(now, rawToken, now.Add(time.Hour))
			session.SuperAdmin.IsActive = false
			return session
		}()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &fakeStore{session: test.session, sessionErr: test.storeErr}
			service := newTestService(t, store, &fakeLimiter{}, &fakePasswordVerifier{}, true, now)

			_, err := service.ResolveSession(context.Background(), test.token)
			if !errors.Is(err, ErrUnauthenticated) {
				t.Fatalf("ResolveSession() error = %v, want %v", err, ErrUnauthenticated)
			}
			if store.touchedSessionID != "" {
				t.Fatal("invalid session must not be touched")
			}
		})
	}
}

func TestResolveSessionUsesTokenHashAndReturnsIdentity(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	rawToken := tokenFromByte(9)
	store := &fakeStore{session: validSession(now, rawToken, now.Add(time.Hour))}
	service := newTestService(t, store, &fakeLimiter{}, &fakePasswordVerifier{}, true, now)

	resolved, err := service.ResolveSession(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("ResolveSession() error = %v", err)
	}
	wantHash := sha256.Sum256([]byte(rawToken))
	if !bytes.Equal(store.lookupTokenHash, wantHash[:]) {
		t.Fatalf("lookup hash = %x, want %x", store.lookupTokenHash, wantHash)
	}
	if store.touchedSessionID != "session-id" {
		t.Fatalf("touched session ID = %q", store.touchedSessionID)
	}
	if resolved.SuperAdmin.Email != "admin@example.com" || resolved.SuperAdmin.PasswordHash != "" {
		t.Fatalf("resolved identity = %#v", resolved.SuperAdmin)
	}
}

func TestResolveSessionUsesValidationInstantForLastSeen(t *testing.T) {
	t.Parallel()

	validatedAt := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	expiresAt := validatedAt.Add(time.Nanosecond)
	rawToken := tokenFromByte(10)
	store := &fakeStore{session: validSession(validatedAt, rawToken, expiresAt)}

	clockCalls := 0
	service, err := NewService(
		store,
		&fakeLimiter{},
		&fakePasswordVerifier{},
		ServiceOptions{
			OTPBypassEnabled:  true,
			SessionTTL:        8 * time.Hour,
			DummyPasswordHash: "dummy-hash",
			Now: func() time.Time {
				clockCalls++
				if clockCalls == 1 {
					return validatedAt
				}
				return expiresAt
			},
		},
	)
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if _, err := service.ResolveSession(context.Background(), rawToken); err != nil {
		t.Fatalf("ResolveSession() error = %v", err)
	}
	if clockCalls != 1 {
		t.Fatalf("clock calls = %d, want one validity instant", clockCalls)
	}
	if !store.touchedAt.Equal(validatedAt) || !store.touchedAt.Before(expiresAt) {
		t.Fatalf("last_seen_at = %s, want validated instant %s before expiry", store.touchedAt, validatedAt)
	}
}

func TestLogoutRequiresCSRFAndRevokesServerSession(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 19, 10, 0, 0, 0, time.UTC)
	rawToken := tokenFromByte(11)

	t.Run("invalid CSRF", func(t *testing.T) {
		t.Parallel()
		store := &fakeStore{session: validSession(now, rawToken, now.Add(time.Hour))}
		service := newTestService(t, store, &fakeLimiter{}, &fakePasswordVerifier{}, true, now)

		err := service.Logout(context.Background(), rawToken, "wrong-token")
		if !errors.Is(err, ErrInvalidCSRFToken) {
			t.Fatalf("Logout() error = %v, want %v", err, ErrInvalidCSRFToken)
		}
		if store.revokedSessionID != "" {
			t.Fatal("invalid CSRF request must not revoke a session")
		}
	})

	t.Run("valid CSRF", func(t *testing.T) {
		t.Parallel()
		store := &fakeStore{session: validSession(now, rawToken, now.Add(time.Hour))}
		service := newTestService(t, store, &fakeLimiter{}, &fakePasswordVerifier{}, true, now)

		err := service.Logout(context.Background(), rawToken, "csrf-token")
		if err != nil {
			t.Fatalf("Logout() error = %v", err)
		}
		if store.revokedSessionID != "session-id" {
			t.Fatalf("revoked session ID = %q", store.revokedSessionID)
		}
	})
}

func newTestService(
	t *testing.T,
	store Store,
	limiter LoginRateLimiter,
	verifier PasswordVerifier,
	bypass bool,
	now time.Time,
) *Service {
	t.Helper()

	randomMaterial := append(bytes.Repeat([]byte{1}, sessionTokenBytes), bytes.Repeat([]byte{2}, sessionTokenBytes)...)
	service, err := NewService(store, limiter, verifier, ServiceOptions{
		OTPBypassEnabled:  bypass,
		SessionTTL:        8 * time.Hour,
		DummyPasswordHash: "dummy-hash",
		Random:            bytes.NewReader(randomMaterial),
		Now:               func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	return service
}

func activeAdmin() SuperAdmin {
	return SuperAdmin{
		ID:           "00000000-0000-0000-0000-000000000001",
		Email:        "admin@example.com",
		PasswordHash: "stored-hash",
		DisplayName:  "Super Admin",
		IsActive:     true,
	}
}

func validSession(now time.Time, _ string, expiresAt time.Time) Session {
	return Session{
		ID:         "session-id",
		SuperAdmin: activeAdmin(),
		CSRFToken:  "csrf-token",
		CreatedAt:  now.Add(-time.Minute),
		ExpiresAt:  expiresAt,
		LastSeenAt: now.Add(-time.Minute),
	}
}

func tokenFromByte(value byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, sessionTokenBytes))
}
