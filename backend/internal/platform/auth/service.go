package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"
)

const sessionTokenBytes = 32

// PasswordVerifier is kept narrow so authentication tests can prove account
// enumeration and failure behavior without weakening production hashing.
type PasswordVerifier interface {
	Verify(encodedHash string, password string) (bool, error)
}

// LoginRateLimiter is a fail-closed Redis-backed security boundary. The
// implementation uses both normalized email and connection identity.
type LoginRateLimiter interface {
	Allow(ctx context.Context, normalizedEmail string, requestingIP string) error
}

type ServiceOptions struct {
	OTPBypassEnabled  bool
	SessionTTL        time.Duration
	DummyPasswordHash string
	Random            io.Reader
	Now               func() time.Time
}

// Service owns authentication sequencing and session validity rules. HTTP
// handlers cannot skip password verification, OTP gating, or PostgreSQL-backed
// session checks by constructing their own response.
type Service struct {
	store             Store
	limiter           LoginRateLimiter
	passwords         PasswordVerifier
	otpBypassEnabled  bool
	sessionTTL        time.Duration
	dummyPasswordHash string
	random            io.Reader
	now               func() time.Time
}

func NewService(
	store Store,
	limiter LoginRateLimiter,
	passwords PasswordVerifier,
	options ServiceOptions,
) (*Service, error) {
	if store == nil || limiter == nil || passwords == nil {
		return nil, errors.New("authentication dependencies are required")
	}
	if options.SessionTTL <= 0 {
		return nil, errors.New("session TTL must be positive")
	}
	if options.DummyPasswordHash == "" {
		return nil, errors.New("dummy password hash is required")
	}

	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}

	return &Service{
		store:             store,
		limiter:           limiter,
		passwords:         passwords,
		otpBypassEnabled:  options.OTPBypassEnabled,
		sessionTTL:        options.SessionTTL,
		dummyPasswordHash: options.DummyPasswordHash,
		random:            options.Random,
		now:               options.Now,
	}, nil
}

// Login verifies the password before applying the local-only OTP bypass. When
// the bypass is disabled, correct credentials stop at ErrOTPRequired and no
// authenticated session is created.
func (s *Service) Login(
	ctx context.Context,
	email string,
	password string,
	requestingIP string,
) (AuthenticatedSession, string, error) {
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil || ValidateLoginPassword(password) != nil {
		return AuthenticatedSession{}, "", ErrInvalidCredentials
	}

	if err := s.limiter.Allow(ctx, normalizedEmail, requestingIP); err != nil {
		if errors.Is(err, ErrRateLimited) {
			return AuthenticatedSession{}, "", ErrRateLimited
		}

		// Redis failure must never permit password verification or session
		// creation. Its internal error is deliberately collapsed here.
		return AuthenticatedSession{}, "", ErrAuthenticationUnavailable
	}

	admin, err := s.store.FindSuperAdminByEmail(ctx, normalizedEmail)
	if errors.Is(err, ErrSuperAdminNotFound) {
		// Verify against a real Argon2id hash even for unknown accounts to
		// reduce obvious account-enumeration timing differences.
		_, _ = s.passwords.Verify(s.dummyPasswordHash, password)
		return AuthenticatedSession{}, "", ErrInvalidCredentials
	}
	if err != nil {
		return AuthenticatedSession{}, "", ErrAuthenticationUnavailable
	}

	passwordMatches, err := s.passwords.Verify(admin.PasswordHash, password)
	if err != nil {
		return AuthenticatedSession{}, "", ErrAuthenticationUnavailable
	}
	if !passwordMatches || !admin.IsActive {
		return AuthenticatedSession{}, "", ErrInvalidCredentials
	}

	if !s.otpBypassEnabled {
		return AuthenticatedSession{}, "", ErrOTPRequired
	}

	rawSessionToken, err := generateOpaqueToken(s.random)
	if err != nil {
		return AuthenticatedSession{}, "", ErrAuthenticationUnavailable
	}
	csrfToken, err := generateOpaqueToken(s.random)
	if err != nil {
		return AuthenticatedSession{}, "", ErrAuthenticationUnavailable
	}

	now := s.now().UTC()
	expiresAt := now.Add(s.sessionTTL)
	tokenHash := hashSessionToken(rawSessionToken)

	if err := s.store.CreateSession(ctx, NewSession{
		SuperAdminID: admin.ID,
		TokenHash:    tokenHash[:],
		CSRFToken:    csrfToken,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		LastSeenAt:   now,
	}); err != nil {
		return AuthenticatedSession{}, "", ErrAuthenticationUnavailable
	}

	return AuthenticatedSession{
		SuperAdmin: publicSuperAdmin(admin),
		CSRFToken:  csrfToken,
		ExpiresAt:  expiresAt,
	}, rawSessionToken, nil
}

// ResolveSession performs every authoritative lookup and validity check from
// the opaque cookie. It never trusts an admin identifier supplied by a browser.
func (s *Service) ResolveSession(ctx context.Context, rawSessionToken string) (AuthenticatedSession, error) {
	session, validatedAt, err := s.findValidSession(ctx, rawSessionToken)
	if err != nil {
		return AuthenticatedSession{}, err
	}

	// Reuse the exact instant that established validity. Calling the clock again
	// here could cross absolute expiry and violate last_seen_at <= expires_at.
	if err := s.store.TouchSession(ctx, session.ID, validatedAt); err != nil {
		return AuthenticatedSession{}, ErrAuthenticationUnavailable
	}

	return AuthenticatedSession{
		SuperAdmin: publicSuperAdmin(session.SuperAdmin),
		CSRFToken:  session.CSRFToken,
		ExpiresAt:  session.ExpiresAt,
	}, nil
}

// Logout validates both the session cookie and the synchronizer CSRF token,
// then revokes the PostgreSQL session before the HTTP layer clears the cookie.
func (s *Service) Logout(ctx context.Context, rawSessionToken string, csrfToken string) error {
	session, validatedAt, err := s.findValidSession(ctx, rawSessionToken)
	if err != nil {
		return err
	}

	if subtle.ConstantTimeCompare([]byte(session.CSRFToken), []byte(csrfToken)) != 1 {
		return ErrInvalidCSRFToken
	}

	if err := s.store.RevokeSession(ctx, session.ID, validatedAt); err != nil {
		return ErrAuthenticationUnavailable
	}

	return nil
}

func (s *Service) findValidSession(
	ctx context.Context,
	rawSessionToken string,
) (Session, time.Time, error) {
	if !validOpaqueToken(rawSessionToken) {
		return Session{}, time.Time{}, ErrUnauthenticated
	}

	tokenHash := hashSessionToken(rawSessionToken)
	session, err := s.store.FindSessionByTokenHash(ctx, tokenHash[:])
	if errors.Is(err, ErrSessionNotFound) {
		return Session{}, time.Time{}, ErrUnauthenticated
	}
	if err != nil {
		return Session{}, time.Time{}, ErrAuthenticationUnavailable
	}

	validatedAt := s.now().UTC()
	if session.RevokedAt != nil || !session.ExpiresAt.After(validatedAt) || !session.SuperAdmin.IsActive {
		return Session{}, time.Time{}, ErrUnauthenticated
	}

	return session, validatedAt, nil
}

func generateOpaqueToken(random io.Reader) (string, error) {
	material := make([]byte, sessionTokenBytes)
	if _, err := io.ReadFull(random, material); err != nil {
		return "", fmt.Errorf("generate secure token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(material), nil
}

func validOpaqueToken(token string) bool {
	material, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(material) != sessionTokenBytes {
		return false
	}

	// Reject alternate textual encodings so one raw token has exactly one
	// canonical cookie representation and therefore one stored hash.
	return base64.RawURLEncoding.EncodeToString(material) == token
}

func hashSessionToken(rawToken string) [sha256.Size]byte {
	return sha256.Sum256([]byte(rawToken))
}

func publicSuperAdmin(admin SuperAdmin) SuperAdmin {
	admin.PasswordHash = ""
	return admin
}
