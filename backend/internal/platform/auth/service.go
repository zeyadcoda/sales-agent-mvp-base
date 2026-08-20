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

	"salesagent.local/backend/internal/requestmeta"
)

const (
	sessionTokenBytes = 32
	otpCleanupTimeout = 2 * time.Second
)

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
	OTPHashSecret     []byte
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
	loginLimiter      LoginRateLimiter
	otpLimiter        OTPRateLimiter
	passwords         PasswordVerifier
	emailSender       OTPEmailSender
	otpHasher         *otpHasher
	otpBypassEnabled  bool
	sessionTTL        time.Duration
	dummyPasswordHash string
	random            io.Reader
	now               func() time.Time
}

func NewService(
	store Store,
	loginLimiter LoginRateLimiter,
	otpLimiter OTPRateLimiter,
	passwords PasswordVerifier,
	emailSender OTPEmailSender,
	options ServiceOptions,
) (*Service, error) {
	if store == nil || loginLimiter == nil || passwords == nil {
		return nil, errors.New("authentication dependencies are required")
	}
	if options.SessionTTL <= 0 {
		return nil, errors.New("session TTL must be positive")
	}
	if options.DummyPasswordHash == "" {
		return nil, errors.New("dummy password hash is required")
	}

	var hasher *otpHasher
	if len(options.OTPHashSecret) > 0 {
		var err error
		hasher, err = newOTPHasher(options.OTPHashSecret)
		if err != nil {
			return nil, err
		}
	}
	if !options.OTPBypassEnabled && (otpLimiter == nil || emailSender == nil || hasher == nil) {
		return nil, errors.New("OTP authentication dependencies are required when bypass is disabled")
	}

	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}

	return &Service{
		store:             store,
		loginLimiter:      loginLimiter,
		otpLimiter:        otpLimiter,
		passwords:         passwords,
		emailSender:       emailSender,
		otpHasher:         hasher,
		otpBypassEnabled:  options.OTPBypassEnabled,
		sessionTTL:        options.SessionTTL,
		dummyPasswordHash: options.DummyPasswordHash,
		random:            options.Random,
		now:               options.Now,
	}, nil
}

// Login verifies the password before applying either server-authorized
// exception. Local bypass remains first and local-only. In real OTP mode, a
// one-time deployment recovery may create the normal session; otherwise the
// password match creates only an inactive challenge that becomes
// browser-visible after successful email delivery and activation.
func (s *Service) Login(
	ctx context.Context,
	email string,
	password string,
	requestingIP string,
) (LoginResult, error) {
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil || ValidateLoginPassword(password) != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	if err := s.loginLimiter.Allow(ctx, normalizedEmail, requestingIP); err != nil {
		if errors.Is(err, ErrRateLimited) {
			return LoginResult{}, ErrRateLimited
		}

		// Redis failure must never permit password verification or session
		// creation. Its internal error is deliberately collapsed here.
		return LoginResult{}, ErrAuthenticationUnavailable
	}

	admin, err := s.store.FindSuperAdminByEmail(ctx, normalizedEmail)
	if errors.Is(err, ErrSuperAdminNotFound) {
		// Verify against a real Argon2id hash even for unknown accounts to
		// reduce obvious account-enumeration timing differences.
		_, _ = s.passwords.Verify(s.dummyPasswordHash, password)
		return LoginResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return LoginResult{}, ErrAuthenticationUnavailable
	}

	passwordMatches, err := s.passwords.Verify(admin.PasswordHash, password)
	if err != nil {
		return LoginResult{}, ErrAuthenticationUnavailable
	}
	if !passwordMatches || !admin.IsActive {
		return LoginResult{}, ErrInvalidCredentials
	}

	if s.otpBypassEnabled {
		authenticated, err := s.createAuthenticatedLogin(ctx, admin)
		if err != nil {
			return LoginResult{}, err
		}

		return LoginResult{Authenticated: &authenticated}, nil
	}

	// Emergency recovery is deliberately below Redis throttling, account-state
	// validation, and correct password verification. This read is only a hint
	// that avoids generating and discarding session material on ordinary OTP
	// logins; the consuming transaction rechecks every condition atomically.
	recoveryCheckedAt := s.now().UTC()
	hasRecovery, err := s.store.HasActiveSuperAdminRecovery(
		ctx,
		admin.ID,
		recoveryCheckedAt,
	)
	if err != nil {
		return LoginResult{}, ErrAuthenticationUnavailable
	}
	if hasRecovery {
		consumedAt := s.now().UTC()
		authenticated, session, err := s.prepareAuthenticatedLogin(admin, consumedAt)
		if err != nil {
			return LoginResult{}, err
		}

		err = s.store.ConsumeSuperAdminRecoveryAndCreateSession(
			ctx,
			admin.ID,
			consumedAt,
			session,
			requestmeta.CorrelationID(ctx),
		)
		switch {
		case err == nil:
			return LoginResult{Authenticated: &authenticated}, nil
		case errors.Is(err, ErrRecoveryNotActive):
			// Another correct-password request may have consumed or revoked the
			// grant after the read hint. This request receives no recovery session
			// and follows the normal OTP path.
		default:
			return LoginResult{}, ErrAuthenticationUnavailable
		}
	}

	challenge, err := s.createAndDeliverOTPChallenge(ctx, admin)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Challenge: &challenge}, nil
}

func (s *Service) createAuthenticatedLogin(
	ctx context.Context,
	admin SuperAdmin,
) (AuthenticatedLogin, error) {
	now := s.now().UTC()
	authenticated, session, err := s.prepareAuthenticatedLogin(admin, now)
	if err != nil {
		return AuthenticatedLogin{}, err
	}

	if err := s.store.CreateSession(ctx, session); err != nil {
		return AuthenticatedLogin{}, ErrAuthenticationUnavailable
	}

	return authenticated, nil
}

// prepareAuthenticatedLogin creates the exact same independent session and
// CSRF material for local bypass and emergency recovery. Persistence remains a
// separate step so recovery consumption, session insertion, and Audit can be
// committed by PostgreSQL as one transaction.
func (s *Service) prepareAuthenticatedLogin(
	admin SuperAdmin,
	now time.Time,
) (AuthenticatedLogin, NewSession, error) {
	rawSessionToken, err := generateOpaqueToken(s.random)
	if err != nil {
		return AuthenticatedLogin{}, NewSession{}, ErrAuthenticationUnavailable
	}
	csrfToken, err := generateOpaqueToken(s.random)
	if err != nil {
		return AuthenticatedLogin{}, NewSession{}, ErrAuthenticationUnavailable
	}

	expiresAt := now.Add(s.sessionTTL)
	tokenHash := hashSessionToken(rawSessionToken)
	session := NewSession{
		SuperAdminID: admin.ID,
		TokenHash:    tokenHash[:],
		CSRFToken:    csrfToken,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
		LastSeenAt:   now,
	}

	return AuthenticatedLogin{
		Session: AuthenticatedSession{
			SuperAdmin: publicSuperAdmin(admin),
			CSRFToken:  csrfToken,
			ExpiresAt:  expiresAt,
		},
		RawSessionToken: rawSessionToken,
	}, session, nil
}

func (s *Service) createAndDeliverOTPChallenge(
	ctx context.Context,
	admin SuperAdmin,
) (PendingChallenge, error) {
	if !s.realOTPAvailable() {
		return PendingChallenge{}, ErrAuthenticationUnavailable
	}

	challengeID, err := generateChallengeID(s.random)
	if err != nil {
		return PendingChallenge{}, ErrAuthenticationUnavailable
	}
	otp, err := generateSixDigitOTP(s.random)
	if err != nil {
		return PendingChallenge{}, ErrAuthenticationUnavailable
	}
	digest, err := s.otpHasher.hash(challengeID, otp)
	if err != nil {
		return PendingChallenge{}, ErrAuthenticationUnavailable
	}

	now := s.now().UTC()
	expiresAt := now.Add(otpValidity)
	resendAvailableAt := now.Add(otpResendCooldown)
	const deliveryVersion = 1
	if err := s.store.CreateOTPChallenge(ctx, NewOTPChallenge{
		ID:                challengeID,
		SuperAdminID:      admin.ID,
		OTPHash:           digest[:],
		CreatedAt:         now,
		ExpiresAt:         expiresAt,
		ResendAvailableAt: resendAvailableAt,
		DeliveryVersion:   deliveryVersion,
		DeliveryStartedAt: now,
	}); err != nil {
		return PendingChallenge{}, ErrAuthenticationUnavailable
	}

	if err := s.emailSender.SendOTP(ctx, OTPEmail{
		RecipientEmail: admin.Email,
		DisplayName:    admin.DisplayName,
		OTP:            otp,
		ExpiresAt:      expiresAt,
	}); err != nil {
		s.invalidateFailedDelivery(ctx, challengeID, deliveryVersion)
		return PendingChallenge{}, ErrAuthenticationUnavailable
	}
	if err := s.store.ActivateOTPChallenge(ctx, challengeID, deliveryVersion, s.now().UTC()); err != nil {
		s.invalidateFailedDelivery(ctx, challengeID, deliveryVersion)
		return PendingChallenge{}, ErrAuthenticationUnavailable
	}

	return pendingChallenge(challengeID, admin.Email, expiresAt, resendAvailableAt), nil
}

// VerifyOTP lets PostgreSQL serialize the lifecycle check, failed-attempt
// update, challenge consumption, and session insert in one transaction.
func (s *Service) VerifyOTP(
	ctx context.Context,
	challengeID string,
	otp string,
	requestingIP string,
) (AuthenticatedLogin, error) {
	if !s.realOTPAvailable() {
		return AuthenticatedLogin{}, ErrAuthenticationUnavailable
	}
	if !validChallengeID(challengeID) || !validOTP(otp) {
		return AuthenticatedLogin{}, ErrOTPInvalid
	}
	if err := s.otpLimiter.AllowVerify(ctx, challengeID, requestingIP); err != nil {
		if errors.Is(err, ErrOTPVerifyRateLimited) {
			return AuthenticatedLogin{}, ErrOTPVerifyRateLimited
		}
		return AuthenticatedLogin{}, ErrAuthenticationUnavailable
	}

	digest, err := s.otpHasher.hash(challengeID, otp)
	if err != nil {
		return AuthenticatedLogin{}, ErrAuthenticationUnavailable
	}
	rawSessionToken, err := generateOpaqueToken(s.random)
	if err != nil {
		return AuthenticatedLogin{}, ErrAuthenticationUnavailable
	}
	csrfToken, err := generateOpaqueToken(s.random)
	if err != nil {
		return AuthenticatedLogin{}, ErrAuthenticationUnavailable
	}

	now := s.now().UTC()
	expiresAt := now.Add(s.sessionTTL)
	tokenHash := hashSessionToken(rawSessionToken)
	admin, err := s.store.VerifyOTPChallengeAndCreateSession(
		ctx,
		challengeID,
		digest[:],
		now,
		NewSession{
			TokenHash:  tokenHash[:],
			CSRFToken:  csrfToken,
			CreatedAt:  now,
			ExpiresAt:  expiresAt,
			LastSeenAt: now,
		},
	)
	if err != nil {
		return AuthenticatedLogin{}, publicOTPError(err)
	}

	return AuthenticatedLogin{
		Session: AuthenticatedSession{
			SuperAdmin: publicSuperAdmin(admin),
			CSRFToken:  csrfToken,
			ExpiresAt:  expiresAt,
		},
		RawSessionToken: rawSessionToken,
	}, nil
}

// ResendOTP rotates PostgreSQL state before calling email. A send failure
// invalidates the pending version; it never revives the previous code.
func (s *Service) ResendOTP(
	ctx context.Context,
	challengeID string,
	requestingIP string,
) (PendingChallenge, error) {
	if !s.realOTPAvailable() {
		return PendingChallenge{}, ErrAuthenticationUnavailable
	}
	if !validChallengeID(challengeID) {
		return PendingChallenge{}, ErrOTPInvalid
	}
	if err := s.otpLimiter.AllowResend(ctx, challengeID, requestingIP); err != nil {
		if errors.Is(err, ErrOTPResendRateLimited) {
			return PendingChallenge{}, ErrOTPResendRateLimited
		}
		return PendingChallenge{}, ErrAuthenticationUnavailable
	}

	startedAt := s.now().UTC()
	expiresAt := startedAt.Add(otpValidity)
	resendAvailableAt := startedAt.Add(otpResendCooldown)
	for attempt := 0; attempt < maxOTPRotationGenerateTries; attempt++ {
		otp, err := generateSixDigitOTP(s.random)
		if err != nil {
			return PendingChallenge{}, ErrAuthenticationUnavailable
		}
		digest, err := s.otpHasher.hash(challengeID, otp)
		if err != nil {
			return PendingChallenge{}, ErrAuthenticationUnavailable
		}

		delivery, err := s.store.BeginOTPChallengeResend(
			ctx,
			challengeID,
			digest[:],
			startedAt,
			expiresAt,
			resendAvailableAt,
		)
		if errors.Is(err, ErrOTPRotationCollision) {
			continue
		}
		if err != nil {
			return PendingChallenge{}, publicOTPError(err)
		}

		if err := s.emailSender.SendOTP(ctx, OTPEmail{
			RecipientEmail: delivery.SuperAdmin.Email,
			DisplayName:    delivery.SuperAdmin.DisplayName,
			OTP:            otp,
			ExpiresAt:      delivery.ExpiresAt,
		}); err != nil {
			s.invalidateFailedDelivery(ctx, challengeID, delivery.DeliveryVersion)
			return PendingChallenge{}, ErrAuthenticationUnavailable
		}
		if err := s.store.ActivateOTPChallenge(
			ctx,
			challengeID,
			delivery.DeliveryVersion,
			s.now().UTC(),
		); err != nil {
			s.invalidateFailedDelivery(ctx, challengeID, delivery.DeliveryVersion)
			return PendingChallenge{}, ErrAuthenticationUnavailable
		}

		return pendingChallenge(
			delivery.ChallengeID,
			delivery.SuperAdmin.Email,
			delivery.ExpiresAt,
			delivery.ResendAvailableAt,
		), nil
	}

	return PendingChallenge{}, ErrAuthenticationUnavailable
}

func (s *Service) GetOTPChallengeStatus(
	ctx context.Context,
	challengeID string,
) (ChallengeState, error) {
	if !s.realOTPAvailable() {
		return ChallengeState{}, ErrAuthenticationUnavailable
	}
	if !validChallengeID(challengeID) {
		return ChallengeState{}, ErrOTPInvalid
	}

	challenge, err := s.store.FindOTPChallenge(ctx, challengeID)
	if err != nil {
		return ChallengeState{}, publicOTPError(err)
	}

	state := OTPChallengePending
	now := s.now().UTC()
	switch {
	case challenge.FailedAttempts >= maxOTPFailedAttempts:
		state = OTPChallengeLocked
	case challenge.ConsumedAt != nil:
		state = OTPChallengeUsed
	case challenge.InvalidatedAt != nil || !challenge.SuperAdmin.IsActive:
		state = OTPChallengeInvalid
	case !challenge.ExpiresAt.After(now):
		state = OTPChallengeExpired
	case challenge.ActiveVersion == nil || *challenge.ActiveVersion != challenge.DeliveryVersion:
		return ChallengeState{}, ErrAuthenticationUnavailable
	}

	return ChallengeState{
		PendingChallenge: pendingChallenge(
			challenge.ID,
			challenge.SuperAdmin.Email,
			challenge.ExpiresAt,
			challenge.ResendAvailableAt,
		),
		State: state,
	}, nil
}

func (s *Service) invalidateFailedDelivery(
	ctx context.Context,
	challengeID string,
	deliveryVersion int,
) {
	// Cleanup gets a short independent deadline so a disconnected browser does
	// not strand provider-failure state as nonterminal. If PostgreSQL is also
	// unavailable, active_version remains NULL and verification still fails
	// closed; the original delivery/activation failure remains authoritative.
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), otpCleanupTimeout)
	defer cancel()
	_ = s.store.InvalidateOTPChallengeDelivery(
		cleanupCtx,
		challengeID,
		deliveryVersion,
		s.now().UTC(),
	)
}

func (s *Service) realOTPAvailable() bool {
	return !s.otpBypassEnabled && s.otpLimiter != nil && s.emailSender != nil && s.otpHasher != nil
}

func pendingChallenge(
	challengeID string,
	email string,
	expiresAt time.Time,
	resendAvailableAt time.Time,
) PendingChallenge {
	return PendingChallenge{
		ID:                challengeID,
		ExpiresAt:         expiresAt,
		ResendAvailableAt: resendAvailableAt,
		DestinationHint:   destinationHint(email),
	}
}

func publicOTPError(err error) error {
	switch {
	case errors.Is(err, ErrOTPChallengeNotFound):
		return ErrOTPInvalid
	case errors.Is(err, ErrOTPInvalid),
		errors.Is(err, ErrOTPExpired),
		errors.Is(err, ErrOTPAttemptsExceeded),
		errors.Is(err, ErrOTPResendTooEarly),
		errors.Is(err, ErrOTPInvalidated),
		errors.Is(err, ErrOTPConsumed),
		errors.Is(err, ErrOTPVerifyRateLimited),
		errors.Is(err, ErrOTPResendRateLimited):
		return err
	default:
		return ErrAuthenticationUnavailable
	}
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
