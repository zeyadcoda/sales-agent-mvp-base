package auth

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrSuperAdminNotFound and ErrSessionNotFound are repository boundary
	// errors. HTTP callers must never receive either error directly because
	// doing so would disclose account or security-state details.
	ErrSuperAdminNotFound   = errors.New("super admin not found")
	ErrSessionNotFound      = errors.New("super admin session not found")
	ErrOTPChallengeNotFound = errors.New("OTP challenge not found")
	ErrSuperAdminExists     = errors.New("super admin already exists")

	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrOTPRequired               = errors.New("email OTP required")
	ErrOTPInvalid                = errors.New("invalid OTP")
	ErrOTPExpired                = errors.New("OTP challenge expired")
	ErrOTPAttemptsExceeded       = errors.New("OTP attempts exceeded")
	ErrOTPResendTooEarly         = errors.New("OTP resend requested too early")
	ErrOTPInvalidated            = errors.New("OTP challenge invalidated")
	ErrOTPConsumed               = errors.New("OTP challenge consumed")
	ErrOTPDeliveryPending        = errors.New("OTP delivery pending")
	ErrOTPRotationCollision      = errors.New("new OTP matches current OTP")
	ErrUnauthenticated           = errors.New("valid authentication session required")
	ErrInvalidCSRFToken          = errors.New("invalid CSRF token")
	ErrAuthenticationUnavailable = errors.New("authentication unavailable")
)

// OTPResendTooEarlyError carries only a relative wait duration. It lets the
// HTTP boundary set Retry-After without disclosing authoritative timestamps.
type OTPResendTooEarlyError struct {
	RetryAfter time.Duration
}

func (err *OTPResendTooEarlyError) Error() string {
	return ErrOTPResendTooEarly.Error()
}

func (err *OTPResendTooEarlyError) Unwrap() error {
	return ErrOTPResendTooEarly
}

// SuperAdmin contains the account fields needed by authentication. The
// password hash is intentionally unexported from all HTTP response DTOs.
type SuperAdmin struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Session is authoritative server-side authentication state loaded from
// PostgreSQL. A raw session token is never represented in this model.
type Session struct {
	ID         string
	SuperAdmin SuperAdmin
	CSRFToken  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	LastSeenAt time.Time
}

// NewSession contains the server-derived fields needed to persist a session.
// TokenHash is SHA-256 output; only the raw token sent in the HttpOnly cookie
// can reproduce it.
type NewSession struct {
	SuperAdminID string
	TokenHash    []byte
	CSRFToken    string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastSeenAt   time.Time
}

// OTPEmail is the only plaintext-code boundary. Implementations deliver it
// without logging or persisting its contents and return only success/failure.
type OTPEmail struct {
	RecipientEmail string
	DisplayName    string
	OTP            string
	ExpiresAt      time.Time
}

// OTPEmailSender is implemented by environment-specific notification
// adapters. Authentication never imports a provider SDK.
type OTPEmailSender interface {
	SendOTP(ctx context.Context, message OTPEmail) error
}

// NewOTPChallenge contains one inactive delivery version. The hash becomes
// verifiable only after ActivateOTPChallenge conditionally activates the same
// version following successful email delivery.
type NewOTPChallenge struct {
	ID                string
	SuperAdminID      string
	OTPHash           []byte
	CreatedAt         time.Time
	ExpiresAt         time.Time
	ResendAvailableAt time.Time
	DeliveryVersion   int
	DeliveryStartedAt time.Time
}

// OTPChallenge is internal authoritative state loaded from PostgreSQL. It is
// never serialized because it contains the OTP hash and account identity.
type OTPChallenge struct {
	ID                string
	SuperAdmin        SuperAdmin
	OTPHash           []byte
	CreatedAt         time.Time
	ExpiresAt         time.Time
	FailedAttempts    int
	ResendAvailableAt time.Time
	DeliveryVersion   int
	ActiveVersion     *int
	DeliveryStartedAt time.Time
	ActivatedAt       *time.Time
	ConsumedAt        *time.Time
	InvalidatedAt     *time.Time
}

// OTPDelivery is returned after PostgreSQL has atomically invalidated the old
// code and installed a new inactive delivery version. Email I/O happens only
// after that transaction releases its row lock.
type OTPDelivery struct {
	ChallengeID       string
	SuperAdmin        SuperAdmin
	DeliveryVersion   int
	ExpiresAt         time.Time
	ResendAvailableAt time.Time
}

// NewSuperAdmin is used only by the deployment/local bootstrap command. It is
// deliberately separate from browser request DTOs so privileged account
// fields can never become mass assignable.
type NewSuperAdmin struct {
	Email        string
	PasswordHash string
	DisplayName  string
	IsActive     bool
}

// Store is the narrow persistence contract owned by the authentication
// domain. PostgreSQL implements it; Redis must never implement or replace it.
type Store interface {
	FindSuperAdminByEmail(ctx context.Context, normalizedEmail string) (SuperAdmin, error)
	HasActiveSuperAdminRecovery(
		ctx context.Context,
		superAdminID string,
		checkedAt time.Time,
	) (bool, error)
	ConsumeSuperAdminRecoveryAndCreateSession(
		ctx context.Context,
		superAdminID string,
		consumedAt time.Time,
		session NewSession,
		correlationID string,
	) error
	CreateSession(ctx context.Context, session NewSession) error
	CreateOTPChallenge(ctx context.Context, challenge NewOTPChallenge) error
	ActivateOTPChallenge(
		ctx context.Context,
		challengeID string,
		deliveryVersion int,
		activatedAt time.Time,
	) error
	InvalidateOTPChallengeDelivery(
		ctx context.Context,
		challengeID string,
		deliveryVersion int,
		invalidatedAt time.Time,
	) error
	BeginOTPChallengeResend(
		ctx context.Context,
		challengeID string,
		otpHash []byte,
		startedAt time.Time,
		expiresAt time.Time,
		resendAvailableAt time.Time,
	) (OTPDelivery, error)
	VerifyOTPChallengeAndCreateSession(
		ctx context.Context,
		challengeID string,
		candidateHash []byte,
		verifiedAt time.Time,
		session NewSession,
	) (SuperAdmin, error)
	FindOTPChallenge(ctx context.Context, challengeID string) (OTPChallenge, error)
	FindSessionByTokenHash(ctx context.Context, tokenHash []byte) (Session, error)
	TouchSession(ctx context.Context, sessionID string, seenAt time.Time) error
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error
}

// ProvisioningStore is intentionally separate from Store because normal API
// authentication has no authority to create privileged accounts.
type ProvisioningStore interface {
	CreateSuperAdmin(ctx context.Context, account NewSuperAdmin) (SuperAdmin, error)
}

// AuthenticatedSession is the safe subset returned to the same-origin web
// application. It contains CSRF material but never the raw session token or
// password hash.
type AuthenticatedSession struct {
	SuperAdmin SuperAdmin
	CSRFToken  string
	ExpiresAt  time.Time
}

// AuthenticatedLogin keeps the raw cookie token inside the application/HTTP
// handoff. The token must never be serialized in a response body.
type AuthenticatedLogin struct {
	Session         AuthenticatedSession
	RawSessionToken string
}

// PendingChallenge is the complete browser-safe context for OTP navigation.
// Possession of its ID does not authenticate or identify an account.
type PendingChallenge struct {
	ID                string
	ExpiresAt         time.Time
	ResendAvailableAt time.Time
	DestinationHint   string
}

// LoginResult is deliberately discriminated: password authentication yields
// either a completed server-authorized session or an email-OTP challenge,
// never both. It contains no marker revealing which server-side path created
// the normal session.
type LoginResult struct {
	Authenticated *AuthenticatedLogin
	Challenge     *PendingChallenge
}

type OTPChallengeState string

const (
	OTPChallengePending OTPChallengeState = "PENDING"
	OTPChallengeExpired OTPChallengeState = "EXPIRED"
	OTPChallengeLocked  OTPChallengeState = "ATTEMPTS_EXCEEDED"
	OTPChallengeInvalid OTPChallengeState = "INVALIDATED"
	OTPChallengeUsed    OTPChallengeState = "CONSUMED"
)

// ChallengeState exposes lifecycle and display metadata only. Hashes, account
// IDs, failure counts, and delivery internals remain server-side.
type ChallengeState struct {
	PendingChallenge
	State OTPChallengeState
}
